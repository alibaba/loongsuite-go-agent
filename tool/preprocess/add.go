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

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
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
		return ex.Error(err)
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
				return ex.Error(err)
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
					return ex.Error(err)
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
			return ex.Error(err)
		}
	}
	return nil
}

func (dp *DepProcessor) newDeps(bundles []*rules.RuleBundle) error {
	content := "package main\n"
	builtin := map[string]string{
		// for go:linkname when declaring printstack/getstack variable
		"unsafe": "_",
		// for debug.Stack and log.Printf when declaring printstack/getstack
		// we do need import alias because user may declare global variable such
		// as "log" or "debug" in their code, which will conflict with the import
		"runtime/debug": "_otel_debug",
		// for log.Printf when declaring printstack/getstack variable
		"log": "_otel_log",
		// otel setup
		"github.com/alibaba/loongsuite-go-agent/pkg": "_",
		"go.opentelemetry.io/otel":                   "_",
		"go.opentelemetry.io/otel/sdk/trace":         "_",
		"go.opentelemetry.io/otel/baggage":           "_",
	}
	for pkg, alias := range builtin {
		content += fmt.Sprintf("import %s %q\n", alias, pkg)
	}

	// No rule bundles? We still need to generate the otel_importer.go file whose
	// purpose is to import the fundamental dependencies
	if len(bundles) == 0 {
		_, err := util.WriteFile(dp.otelRuntimeGo, content)
		if err != nil {
			return ex.Error(err)
		}
		return nil
	}

	// Generate the otel_importer.go file with the rule bundles
	paths := map[string]bool{}
	for _, bundle := range bundles {
		for _, funcRules := range bundle.File2FuncRules {
			for _, rules := range funcRules {
				for _, rule := range rules {
					if rule.GetPath() != "" {
						paths[rule.GetPath()] = true
					}
				}
			}
		}
	}
	addDeps := make([]Dependency, 0)
	for path := range paths {
		content += fmt.Sprintf("import _ %q\n", path)
		t := strings.TrimPrefix(path, pkgPrefix)
		addDeps = append(addDeps, Dependency{
			ImportPath:     path,
			Version:        "v0.0.0-00010101000000-000000000000", // use latest version for the rule import
			Replace:        true,
			ReplacePath:    filepath.Join(dp.pkgLocalCache, t),
			ReplaceVersion: "",
		})
	}
	cnt := 0
	for _, bundle := range bundles {
		tag := ""
		// If we occasionally instrument the main package, we don't need to add
		// the linkname directive, as the target variables are already defined
		// in the main package, adding new linkname for generated code will cause
		// the symbol redefinition error.
		if bundle.ImportPath != "main" {
			tag = fmt.Sprintf("//go:linkname _getstatck%d %s.OtelGetStackImpl\n",
				cnt, bundle.ImportPath)
		}
		content += tag
		s := fmt.Sprintf("var _getstatck%d = _otel_debug.Stack\n", cnt)
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
		return ex.Error(err)
	}

	err = dp.addDependency(dp.getGoModPath(), addDeps)
	if err != nil {
		return ex.Error(err)
	}
	return nil
}
