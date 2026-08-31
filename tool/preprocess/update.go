// Copyright (c) 2024 Alibaba Group Holding Ltd.
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
	"os"
	"path/filepath"

	"github.com/alibaba/loongsuite-go/tool/config"
	"github.com/alibaba/loongsuite-go/tool/data"
	"github.com/alibaba/loongsuite-go/tool/ex"
	"github.com/alibaba/loongsuite-go/tool/rules"
	"github.com/alibaba/loongsuite-go/tool/util"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/semver"
)

const (
	pkgPrefix = "github.com/alibaba/loongsuite-go/pkg"
)

var otelDeps = map[string]string{
	"go.opentelemetry.io/otel":                                          "v1.45.0",
	"go.opentelemetry.io/otel/sdk":                                      "v1.45.0",
	"go.opentelemetry.io/otel/trace":                                    "v1.45.0",
	"go.opentelemetry.io/otel/metric":                                   "v1.45.0",
	"go.opentelemetry.io/otel/sdk/metric":                               "v1.45.0",
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace":                 "v1.45.0",
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc":   "v1.45.0",
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp":   "v1.45.0",
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc": "v1.45.0",
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp": "v1.45.0",
	"go.opentelemetry.io/otel/exporters/prometheus":                     "v0.67.0",
	"go.opentelemetry.io/contrib/instrumentation/runtime":               "v0.70.0",
	"google.golang.org/protobuf":                                        "v1.36.11",
	// Fix "ambiguous import" between legacy monolithic genproto and split genproto modules:
	//
	// Old versions of the monolithic module `google.golang.org/genproto` included packages
	// under `googleapis/rpc` and `googleapis/api`. Newer releases split these into dedicated
	// submodules (`google.golang.org/genproto/googleapis/rpc` and `/api`) by adding nested
	// go.mod files. If a user project pulls an old monolithic genproto (e.g. 2023-04-10)
	// and we pull split modules transitively (via gRPC / grpc-gateway / OTLP), Go sees the
	// same import path in two modules and fails with "ambiguous import".
	//
	// By forcing the monolithic module to a post-split version, it no longer "owns" those
	// packages, and the ambiguity is resolved without pinning the user's gRPC version.
	"google.golang.org/genproto":                             "v0.0.0-20251202230838-ff82c1b0f217",
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric": "v1.45.0",
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace":  "v1.45.0",
	"go.opentelemetry.io/otel/exporters/zipkin":              "v1.45.0",
}

func parseGoMod(gomod string) (*modfile.File, error) {
	data, err := util.ReadFile(gomod)
	if err != nil {
		return nil, err
	}
	modFile, err := modfile.Parse(util.GoModFile, []byte(data), nil)
	if err != nil {
		return nil, ex.Wrap(err)
	}
	return modFile, nil
}

func writeGoMod(gomod string, modfile *modfile.File) error {
	bs, err := modfile.Format()
	if err != nil {
		return ex.Wrap(err)
	}
	_, err = util.WriteFile(gomod, string(bs))
	if err != nil {
		return err
	}
	return nil
}

// Fetch the zipped pkg module from the embedded data section and extract it to
// a temporary directory, then return the path to the pkg directory.
// In -pkg flag is specified, we will use the path specified by the flag instead
// of the temporary directory. This flag is specifically used for a scenario where
// the user compiles a shared library and the main program separately. If the
// injected code uses different temporary pkg directories, it will cause the
// plugin to report a version conflict error. This flag ensures that the shared
// library and the main program share the same pkg directory to prevent this issue.
func findPkgModDir() (string, error) {
	pkgPath := config.GetConf().PkgPath
	// If -pkg flag is specified, check if the path exists. If it does,
	// return the absolute path. If it does not, extract the embedded pkg
	// module into the path.
	if pkgPath != "" {
		if util.PathExists(pkgPath) {
			return filepath.Join(pkgPath, "pkg"), nil
		}
		bs, err := data.UseEmbeddedPkg()
		if err != nil {
			return "", err
		}
		tempPkg := pkgPath
		err = extractGZip(bs, tempPkg)
		if err != nil {
			return "", err
		}
		return filepath.Join(tempPkg, "pkg"), nil
	}
	bs, err := data.UseEmbeddedPkg()
	if err != nil {
		return "", err
	}
	tempPkg := util.GetTempBuildDirWith("alibaba-pkg")
	if util.PathExists(tempPkg) {
		_ = os.RemoveAll(tempPkg)
	}
	err = extractGZip(bs, tempPkg)
	if err != nil {
		return "", err
	}
	return filepath.Join(tempPkg, "pkg"), nil
}

// updateRule rectifies the file rules path to the local module cache path.
func (dp *DepProcessor) updateRule(bundles []*rules.InstRuleSet) error {
	defer util.PhaseTimer("Fetch")()
	modfile, err := parseGoMod(dp.getGoModPath())
	if err != nil {
		return err
	}
	// Collect all the replace directives from go.mod file, we will use it to
	// rectify the custom rules.
	replaceMap := map[string]string{}
	for _, replace := range modfile.Replace {
		replaceMap[replace.Old.Path] = replace.New.Path
	}
	rectified := map[string]bool{}
	for _, bundle := range bundles {
		for _, funcRules := range bundle.FuncRules {
			for _, rule := range funcRules {
				if rule.UseRaw {
					continue
				}
				if rectified[rule.GetPath()] {
					continue
				}
				_, path, err := dp.findRuleDir(rule.GetPath())
				if err != nil {
					return err
				}
				rule.SetPath(path)
				rectified[path] = true
			}
		}
		for _, fileRule := range bundle.FileRules {
			if rectified[fileRule.GetPath()] {
				continue
			}
			_, path, err := dp.findRuleDir(fileRule.GetPath())
			if err != nil {
				return err
			}
			fileRule.SetPath(path)
			fileRule.FileName = filepath.Join(path, fileRule.FileName)
			rectified[path] = true
		}
	}
	return nil
}

func (dp *DepProcessor) updateGoMod() error {
	// Backup go.mod and go.sum files
	gomodDir := dp.getGoModDir()
	files := []string{}
	files = append(files, filepath.Join(gomodDir, util.GoModFile))
	files = append(files, filepath.Join(gomodDir, util.GoSumFile))
	files = append(files, filepath.Join(gomodDir, util.GoWorkSumFile))
	for _, file := range files {
		if util.PathExists(file) {
			err := dp.backupFile(file)
			if err != nil {
				return err
			}
		}
	}
	// Add the alibaba-otel pkg module to the go.mod file
	addDeps := make([]Dependency, 0)
	dep := Dependency{
		ImportPath:     pkgPrefix,
		Version:        "v0.0.0-00010101000000-000000000000",
		Replace:        true,
		ReplacePath:    dp.pkgModDir,
		ReplaceVersion: "",
	}
	addDeps = append(addDeps, dep)
	// OTel dependencies may publish new versions that are not compatible
	// with the otel tool. In such cases, we need to add a replace directive
	// to use certain versions of the OTel dependencies for given otel tool,
	// otherwise, the otel tool may fail to run.
	for path, version := range otelDeps {
		addDeps = append(addDeps, Dependency{
			ImportPath:     path,
			Version:        version,
			Replace:        true,
			ReplacePath:    path,
			ReplaceVersion: version,
		})
	}
	err := dp.addDependency(dp.getGoModPath(), addDeps)
	if err != nil {
		return err
	}
	// Update the existing replace directives to use the local module cache
	// Very bad, we must guarantee the replace path is consistent either in
	// go.mod or vendor/modules.txt, otherwise, the go build toolchain will fail
	// so we must parse go.mod to check if there is any existing replace directive
	// and update the vendor/modules.txt accordingly.
	modfile, err := parseGoMod(dp.getGoModPath())
	if err != nil {
		return err
	}
	changed := false
	for _, replace := range modfile.Replace {
		if replace.Old.Path == pkgPrefix {
			err = modfile.DropReplace(pkgPrefix, "")
			if err != nil {
				return ex.Wrap(err)
			}
			err = modfile.AddReplace(pkgPrefix, "", dp.pkgModDir, "")
			if err != nil {
				return ex.Wrap(err)
			}
			changed = true
			break
		}
	}
	if changed {
		err = writeGoMod(dp.getGoModPath(), modfile)
		if err != nil {
			return err
		}
	}
	return nil
}

// warnDependencyUpgrades reports dependencies that the instrumentation graph
// resolves to a higher version than the one pinned in the user's original
// go.mod. The user's go.mod is restored after the build, but the binary is
// still compiled against the upgraded versions, which can break projects that
// depend on locked versions (#386, #387).
func (dp *DepProcessor) warnDependencyUpgrades() {
	backup, ok := dp.backups[dp.getGoModPath()]
	if !ok {
		return
	}
	orig, err := parseGoMod(backup)
	if err != nil {
		return
	}
	cur, err := parseGoMod(dp.getGoModPath())
	if err != nil {
		return
	}
	upgraded := func(from, to string) bool {
		if !semver.IsValid(from) {
			from = "v" + from // go directive versions carry no prefix
		}
		if !semver.IsValid(to) {
			to = "v" + to
		}
		return semver.IsValid(from) && semver.IsValid(to) &&
			semver.Compare(to, from) > 0
	}
	origVer := make(map[string]string, len(orig.Require))
	for _, r := range orig.Require {
		origVer[r.Mod.Path] = r.Mod.Version
	}
	lines := make([]string, 0, 4)
	for _, r := range cur.Require {
		if from, ok := origVer[r.Mod.Path]; ok && upgraded(from, r.Mod.Version) {
			lines = append(lines, fmt.Sprintf("  %s %s -> %s", r.Mod.Path, from, r.Mod.Version))
		}
	}
	if orig.Go != nil && cur.Go != nil && upgraded(orig.Go.Version, cur.Go.Version) {
		lines = append(lines, fmt.Sprintf("  go directive %s -> %s", orig.Go.Version, cur.Go.Version))
	}
	if len(lines) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "WARNING: instrumentation resolves dependencies pinned in your go.mod to newer versions,")
	fmt.Fprintln(os.Stderr, "the binary is built against them even though your go.mod is restored afterwards:")
	for _, line := range lines {
		fmt.Fprintln(os.Stderr, line)
	}
	fmt.Fprintln(os.Stderr, "If your project relies on the pinned versions, pin them with replace directives or")
	fmt.Fprintln(os.Stderr, "exclude the rule via OTEL_INSTRUMENTATION_*_ENABLED=false (see issue #386).")
}
