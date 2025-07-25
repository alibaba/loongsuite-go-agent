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
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

// -----------------------------------------------------------------------------
// Preprocess
//
// The preprocess package is used to preprocess the source code before the actual
// instrumentation. Instrumentation rules may introduces additional dependencies
// that are not present in original source code. The preprocess is responsible
// for preparing these dependencies in advance.

const (
	OtelRuntimeGo    = "otel.runtime.go"
	OtelBackups      = "backups"
	OtelBackupSuffix = ".bk"
	DryRunLog        = "dry_run.log"
	CompileRemix     = "remix"
	VendorDir        = "vendor"
	GoCacheDir       = "gocache"
)

type DepProcessor struct {
	backups       map[string]string
	moduleName    string // Module name from go.mod
	modulePath    string // Where go.mod is located
	goBuildCmd    []string
	vendorMode    bool
	pkgLocalCache string // Local module cache path of alibaba-otel pkg module
	otelRuntimeGo string // Path to the otel.runtime.go file
}

func newDepProcessor() *DepProcessor {
	dp := &DepProcessor{
		backups:       map[string]string{},
		vendorMode:    false,
		pkgLocalCache: "",
		otelRuntimeGo: "",
	}
	return dp
}

func (dp *DepProcessor) String() string {
	return fmt.Sprintf("moduleName: %s, modulePath: %s, goBuildCmd: %v, vendorMode: %v, pkgLocalCache: %s, OtelRuntimeGo: %s",
		dp.moduleName, dp.modulePath, dp.goBuildCmd, dp.vendorMode,
		dp.pkgLocalCache, dp.otelRuntimeGo)
}

func (dp *DepProcessor) getGoModPath() string {
	util.Assert(dp.modulePath != "", "modulePath is empty")
	return dp.modulePath
}

func (dp *DepProcessor) getGoModDir() string {
	return filepath.Dir(dp.getGoModPath())
}

// Run runs the command and returns the combined standard output and standard
// error. dir specifies the working directory of the command. If dir is the
// empty string, run runs the command in the calling process's current directory.
func runCmdCombinedOutput(dir string, env []string, args ...string) (string, error) {
	path := args[0]
	args = args[1:]
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", ex.Errorf(err, "%s", string(out))
	}
	return string(out), nil
}

func runCmdOutput(dir string, args ...string) (string, error) {
	path := args[0]
	args = args[1:]
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", ex.Errorf(err, "command %v", args)
	}
	return string(out), nil
}

func (dp *DepProcessor) postProcess() {
	util.GuaranteeInPreprocess()

	// Using -debug? Leave all changes for debugging
	if config.GetConf().Debug {
		return
	}

	_ = os.RemoveAll(dp.otelRuntimeGo)
	_ = os.RemoveAll(util.GetTempBuildDirWith("alibaba-pkg"))
	_ = dp.restoreBackupFiles()
}

func (dp *DepProcessor) backupFile(origin string) error {
	util.GuaranteeInPreprocess()
	backup := filepath.Base(origin) + OtelBackupSuffix
	backup = util.GetLogPath(filepath.Join(OtelBackups, backup))
	err := os.MkdirAll(filepath.Dir(backup), 0777)
	if err != nil {
		return ex.Error(err)
	}
	if _, exist := dp.backups[origin]; !exist {
		err = util.CopyFile(origin, backup)
		if err != nil {
			return ex.Error(err)
		}
		dp.backups[origin] = backup
		util.Log("Backup %v", origin)
	} else if config.GetConf().Verbose {
		util.Log("Backup %v already exists", origin)
	}
	return nil
}

func (dp *DepProcessor) restoreBackupFiles() error {
	util.GuaranteeInPreprocess()
	for origin, backup := range dp.backups {
		err := util.CopyFile(backup, origin)
		if err != nil {
			return ex.Error(err)
		}
		util.Log("Restore %v", origin)
	}
	return nil
}

func getTempGoCache() (string, error) {
	goCachePath, err := filepath.Abs(filepath.Join(util.TempBuildDir, GoCacheDir))
	if err != nil {
		return "", ex.Error(err)
	}

	if !util.PathExists(goCachePath) {
		err = os.MkdirAll(goCachePath, 0755)
		if err != nil {
			return "", ex.Error(err)
		}
	}
	return goCachePath, nil
}

func buildGoCacheEnv(value string) []string {
	return []string{"GOCACHE=" + value}
}

func runBuildWithToolexec(goBuildCmd []string) error {
	exe, err := os.Executable()
	if err != nil {
		return ex.Error(err)
	}
	// go build/install
	args := []string{}
	args = append(args, goBuildCmd[:2]...)
	// Remix toolexec
	args = append(args, "-toolexec="+exe+" "+CompileRemix)

	// Leave the temporary compilation directory
	args = append(args, util.BuildWork)

	// Force rebuilding
	args = append(args, "-a")

	if config.GetConf().Debug {
		// Disable compiler optimizations for debugging mode
		args = append(args, "-gcflags=all=-N -l")
	}

	// Append additional build arguments provided by the user
	args = append(args, goBuildCmd[2:]...)

	util.Log("Run toolexec build: %v", args)
	util.AssertGoBuild(args)

	// get the temporary build cache path
	goCachePath, err := getTempGoCache()
	if err != nil {
		return ex.Error(err)
	}
	util.Log("Using isolated GOCACHE: %s", goCachePath)

	// @@ Note that we should not set the working directory here, as the build
	// with toolexec should be run in the same directory as the original build
	// command
	out, err := runCmdCombinedOutput("", buildGoCacheEnv(goCachePath), args...)
	util.Log("Output from toolexec build: %v", out)
	if err != nil {
		return ex.Error(err)
	}
	return nil
}

func precheck() error {
	// Check if the project is modularized
	go11module := os.Getenv("GO111MODULE")
	if go11module == "off" {
		return ex.Errorf(nil, "GO111MODULE is off")
	}

	// Check if the build arguments is sane
	if len(os.Args) < 3 {
		config.PrintVersion()
		os.Exit(0)
	}
	if !strings.Contains(os.Args[1], "go") {
		config.PrintVersion()
		os.Exit(0)
	}
	if os.Args[2] != "build" && os.Args[2] != "install" {
		// exec original go command
		err := util.RunCmd(os.Args[1:]...)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	return nil
}

func (dp *DepProcessor) saveDebugFiles() {
	dir := filepath.Join(util.GetTempBuildDir(), "changed")
	err := os.MkdirAll(dir, os.ModePerm)
	if err == nil {
		for origin := range dp.backups {
			util.CopyFile(origin, filepath.Join(dir, filepath.Base(origin)))
		}
	}
	_ = util.CopyFile(dp.otelRuntimeGo, filepath.Join(dir, OtelRuntimeGo))
}

func Preprocess() error {
	// Make sure the project is modularized otherwise we cannot proceed
	err := precheck()
	if err != nil {
		return ex.Error(err)
	}

	dp := newDepProcessor()

	err = dp.init()
	if err != nil {
		return ex.Error(err)
	}
	defer func() { dp.postProcess() }()
	{
		defer util.PhaseTimer("Preprocess")()
		defer dp.saveDebugFiles()

		// Backup go.mod and add additional repalce directives for the pkg module
		err = dp.updateGoMod()
		if err != nil {
			return ex.Error(err)
		}

		// Two round of rule matching
		//    {prepare->refresh}
		//        1st match
		//    {prepare->refresh}
		//        2nd match
		//    {prepare->refresh}
		// Let's break it down a little bit. We first prepare the rule import,
		// which is used to import foundational dependencies (e.g., otel, as we
		// will instrument the otel SDK itself). Then, we perform a refresh to
		// ensure dependencies are ready and proceed to the 1st match. During
		// this phase, some rules matching specific criteria are identified. We
		// then update the rule import again to include these newly matched rules.
		// Since these rules may (and likely will) break the original dependency
		// graph, a 2nd match is required to resolve the final set of rules.
		// These final rules are used to perform a final update of the rule import.
		// At this point, all preparations are complete, and the process can
		// advance to the second stage: instrumentation.
		bundles := make([]*rules.RuleBundle, 0)
		for i := 0; i < 3; i++ {
			util.Log("Round %d of rule matching", i+1)
			err = dp.newDeps(bundles)
			if err != nil {
				return ex.Error(err)
			}

			err = dp.syncDeps()
			if err != nil {
				return ex.Error(err)
			}
			if i == 2 {
				continue
			}
			bundles, err = dp.matchRules()
			if err != nil {
				return ex.Error(err)
			}
		}

		// Rectify file rules to make sure we can find them locally
		err = dp.updateRule(bundles)
		if err != nil {
			return ex.Error(err)
		}

		// From this point on, we no longer modify the rules
		err = rules.StoreRuleBundles(bundles)
		if err != nil {
			return ex.Error(err)
		}
	}

	{
		defer util.PhaseTimer("Instrument")()

		// Run go build with toolexec to start instrumentation
		err = runBuildWithToolexec(dp.goBuildCmd)
		if err != nil {
			return ex.Error(err)
		}
	}
	util.Log("Build completed successfully")
	return nil
}
