// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package preprocess

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go/tool/ast"
	"github.com/alibaba/loongsuite-go/tool/ex"
	"github.com/alibaba/loongsuite-go/tool/rules"
	"github.com/alibaba/loongsuite-go/tool/util"
	"golang.org/x/mod/semver"
)

type Dependency struct {
	ImportPath     string // the import path of the dependency
	Version        string // the version of the dependency
	Replace        bool   // whether the dependency should be replaced
	ReplacePath    string // the path of the dependency
	ReplaceVersion string // the version of the dependency
}

func (dp *DepProcessor) addDependency(gomod string, dependencies []Dependency) error {
	modfile, err := parseGoMod(gomod)
	if err != nil {
		return err
	}
	// For each dependency, check if it is already in the go.mod file and add
	// it using require directive. If the dependency specifies a replace path,
	// then further add a replace directive if it is not already in the go.mod
	changed := false
	for _, dependency := range dependencies {
		alreadyRequire := false
		for _, r := range modfile.Require {
			if r.Mod.Path == dependency.ImportPath {
				alreadyRequire = true
				break
			}
		}
		if !alreadyRequire {
			err = modfile.AddRequire(dependency.ImportPath, dependency.Version)
			if err != nil {
				return ex.Wrap(err)
			}
			changed = true
			util.Log("Add require dependency %s %s",
				dependency.ImportPath, dependency.Version)
		}
		if dependency.Replace {
			alreadyReplace := false
			for _, r := range modfile.Replace {
				if r.Old.Path == dependency.ImportPath {
					alreadyReplace = true
					break
				}
			}
			if !alreadyReplace {
				err = modfile.AddReplace(dependency.ImportPath, "",
					dependency.ReplacePath, dependency.ReplaceVersion)
				if err != nil {
					return ex.Wrap(err)
				}
				changed = true
				util.Log("Add replace dependency %s %s => %s %s",
					dependency.ImportPath, dependency.Version,
					dependency.ReplacePath, dependency.ReplaceVersion)
			}
		}
	}
	// Once all dependencies are added and write back to go.mod
	if changed {
		err = writeGoMod(gomod, modfile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dp *DepProcessor) originalGoModPath(gomod string) string {
	if backup, ok := dp.backups[gomod]; ok && backup != "" {
		return backup
	}
	return gomod
}

func canPinHookDependency(userVersion, hookVersion string) bool {
	if !semver.IsValid(userVersion) || !semver.IsValid(hookVersion) {
		return false
	}
	userMajor := semver.Major(userVersion)
	hookMajor := semver.Major(hookVersion)
	if userMajor == "v0" || hookMajor == "v0" {
		return false
	}
	if userMajor != hookMajor {
		return false
	}
	return true
}

type hookDependencyPin struct {
	path    string
	version string
}

func (dp *DepProcessor) pinConflictingHookDependencies(gomod string, dependencies []Dependency) error {
	userModfile, err := parseGoMod(dp.originalGoModPath(gomod))
	if err != nil {
		return err
	}
	userVersions := make(map[string]string)
	for _, req := range userModfile.Require {
		if req.Indirect {
			continue
		}
		if req.Mod.Path != "" && req.Mod.Version != "" {
			userVersions[req.Mod.Path] = req.Mod.Version
		}
	}
	if len(userVersions) == 0 {
		return nil
	}

	existingReplaces := make(map[string]bool)
	for _, replace := range userModfile.Replace {
		existingReplaces[replace.Old.Path] = true
	}

	for _, dependency := range dependencies {
		if !dependency.Replace || dependency.ReplacePath == "" {
			continue
		}
		hookGoMod := filepath.Join(dependency.ReplacePath, util.GoModFile)
		if util.PathNotExists(hookGoMod) {
			continue
		}
		hookModfile, err := parseGoMod(hookGoMod)
		if err != nil {
			return err
		}
		changed := false
		pins := make([]hookDependencyPin, 0)
		for _, req := range hookModfile.Require {
			path := req.Mod.Path
			if req.Indirect {
				continue
			}
			userVersion, ok := userVersions[path]
			if !ok || userVersion == "" || userVersion == req.Mod.Version {
				continue
			}
			if !canPinHookDependency(userVersion, req.Mod.Version) {
				continue
			}
			if existingReplaces[path] {
				continue
			}
			pins = append(pins, hookDependencyPin{
				path:    path,
				version: userVersion,
			})
		}
		for _, pin := range pins {
			err = hookModfile.DropRequire(pin.path)
			if err != nil {
				return ex.Wrap(err)
			}
			err = hookModfile.AddRequire(pin.path, pin.version)
			if err != nil {
				return ex.Wrap(err)
			}
			changed = true
			util.Log("Pin hook dependency %s to user version %s", pin.path, pin.version)
		}

		if changed {
			err = writeGoMod(hookGoMod, hookModfile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (dp *DepProcessor) findRuleDir(path string) (string, string, error) {
	// The rule can be either a standard rule or a custom rule
	// We should identify it and define how to find it
	if util.PathExists(path) {
		modfile, err := parseGoMod(filepath.Join(path, util.GoModFile))
		if err != nil {
			return "", "", err
		}
		// Custom rule, find it locally
		moduleName := modfile.Module.Mod.Path
		replacePath := path
		return moduleName, replacePath, nil
	} else {
		// Standard rule, find it from the pkg module dir
		t := strings.TrimPrefix(path, pkgPrefix)
		moduleName := path
		replacePath := filepath.Join(dp.pkgModDir, t)
		return moduleName, replacePath, nil
	}
}

func (dp *DepProcessor) newDeps(bundles []*rules.InstRuleSet) error {
	content := "package main\n"
	builtin := map[string]string{
		// for go:linkname when declaring printstack/getstack variable
		"unsafe": ast.IdentIgnore,
		// for debug.Stack and log.Printf when declaring printstack/getstack
		// we do need import alias because user may declare global variable such
		// as "log" or "debug" in their code, which will conflict with the import
		"runtime/debug": "_otel_debug",
		// for log.Printf when declaring printstack/getstack variable
		"log": "_otel_log",
		// otel setup
		"github.com/alibaba/loongsuite-go/pkg": ast.IdentIgnore,
		"go.opentelemetry.io/otel":             ast.IdentIgnore,
		"go.opentelemetry.io/otel/sdk/trace":   ast.IdentIgnore,
		"go.opentelemetry.io/otel/baggage":     ast.IdentIgnore,
	}
	for pkg, alias := range builtin {
		content += fmt.Sprintf("import %s %q\n", alias, pkg)
	}

	// No rule bundles? We still need to generate the otel_importer.go file whose
	// purpose is to import the fundamental dependencies
	if len(bundles) == 0 {
		_, err := util.WriteFile(dp.otelRuntimeGo, content)
		if err != nil {
			return err
		}
		return nil
	}

	// Generate the otel.runtime.go file with the rule bundles
	addDeps := make([]Dependency, 0)
	for _, bundle := range bundles {
		for _, funcRules := range bundle.FuncRules {
			for _, rule := range funcRules {
				path := rule.GetPath()
				if path != "" {
					moduleName, replacePath, err := dp.findRuleDir(path)
					if err != nil {
						return err
					}
					content += fmt.Sprintf("import _ %q\n", moduleName)
					addDeps = append(addDeps, Dependency{
						ImportPath: moduleName,
						// use latest version for the rule import
						Version:        "v0.0.0-00010101000000-000000000000",
						Replace:        true,
						ReplacePath:    replacePath,
						ReplaceVersion: "",
					})
				}
			}
		}
	}
	cnt := 0
	for _, bundle := range bundles {
		tag := ""
		// If we occasionally instrument the main package, we don't need to add
		// the linkname directive, as the target variables are already defined
		// in the main package, adding new linkname for generated code will cause
		// the symbol redefinition error.
		if bundle.ImportPath != "main" {
			tag = fmt.Sprintf("//go:linkname _getstack%d %s.OtelGetStackImpl\n",
				cnt, bundle.ImportPath)
		}
		content += tag
		s := fmt.Sprintf("var _getstack%d = _otel_debug.Stack\n", cnt)
		content += s
		if bundle.ImportPath != "main" {
			tag = fmt.Sprintf("//go:linkname _printstack%d %s.OtelPrintStackImpl\n",
				cnt, bundle.ImportPath)
		}
		content += tag
		s = fmt.Sprintf("var _printstack%d = func (bt []byte){ _otel_log.Printf(string(bt)) }\n", cnt)
		content += s
		cnt++
	}
	_, err := util.WriteFile(dp.otelRuntimeGo, content)
	if err != nil {
		return err
	}

	err = dp.addDependency(dp.getGoModPath(), addDeps)
	if err != nil {
		return err
	}
	err = dp.pinConflictingHookDependencies(dp.getGoModPath(), addDeps)
	if err != nil {
		return err
	}
	return nil
}
