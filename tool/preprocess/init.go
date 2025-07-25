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
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/alibaba/loongsuite-go-agent/tool/ast"
	"github.com/alibaba/loongsuite-go-agent/tool/config"
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
	"github.com/dave/dst"
	"golang.org/x/tools/go/packages"
)

// $ go help packages
// Many commands apply to a set of packages:
//
//	go <action> [packages]
//
// Usually, [packages] is a list of import paths.
//
// An import path that is a rooted path or that begins with
// a . or .. element is interpreted as a file system path and
// denotes the package in that directory.
//
// Otherwise, the import path P denotes the package found in
// the directory DIR/src/P for some DIR listed in the GOPATH
// environment variable (For more details see: 'go help gopath').
//
// If no import paths are given, the action applies to the
// package in the current directory.
//
// There are four reserved names for paths that should not be used
// for packages to be built with the go tool:
//
// - "main" denotes the top-level package in a stand-alone executable.
//
// - "all" expands to all packages found in all the GOPATH
// trees. For example, 'go list all' lists all the packages on the local
// system. When using modules, "all" expands to all packages in
// the main module and their dependencies, including dependencies
// needed by tests of any of those.
//
// - "std" is like all but expands to just the packages in the standard
// Go library.
//
// - "cmd" expands to the Go repository's commands and their
// internal libraries.
//
// Import paths beginning with "cmd/" only match source code in
// the Go repository.
//
// An import path is a pattern if it includes one or more "..." wildcards,
// each of which can match any string, including the empty string and
// strings containing slashes. Such a pattern expands to all package
// directories found in the GOPATH trees with names matching the
// patterns.
//
// To make common patterns more convenient, there are two special cases.
// First, /... at the end of the pattern can match an empty string,
// so that net/... matches both net and packages in its subdirectories, like net/http.
// Second, any slash-separated pattern element containing a wildcard never
// participates in a match of the "vendor" element in the path of a vendored
// package, so that ./... does not match packages in subdirectories of
// ./vendor or ./mycode/vendor, but ./vendor/... and ./mycode/vendor/... do.
// Note, however, that a directory named vendor that itself contains code
// is not a vendored package: cmd/vendor would be a command named vendor,
// and the pattern cmd/... matches it.
// See golang.org/s/go15vendor for more about vendoring.
//
// An import path can also name a package to be downloaded from
// a remote repository. Run 'go help importpath' for details.
//
// Every package in a program must have a unique import path.
// By convention, this is arranged by starting each path with a
// unique prefix that belongs to you. For example, paths used
// internally at Google all begin with 'google', and paths
// denoting remote repositories begin with the path to the code,
// such as 'github.com/user/repo'.
//
// Packages in a program need not have unique package names,
// but there are two reserved package names with special meaning.
// The name main indicates a command, not a library.
// Commands are built into binaries and cannot be imported.
// The name documentation indicates documentation for
// a non-Go program in the directory. Files in package documentation
// are ignored by the go command.
//
// As a special case, if the package list is a list of .go files from a
// single directory, the command is applied to a single synthesized
// package made up of exactly those files, ignoring any build constraints
// in those files and ignoring any other files in the directory.
//
// Directory and file names that begin with "." or "_" are ignored
// by the go tool, as are directories named "testdata".

func tryLoadPackage(path string) ([]*packages.Package, error) {
	cfg := &packages.Config{
		// Change it unless you know what you are doing
		Mode: packages.NeedModule | packages.NeedFiles | packages.NeedName,
	}

	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return nil, ex.Error(err)
	}
	return pkgs, nil
}

func findModule(buildCmd []string) ([]*packages.Package, error) {
	candidates := make([]*packages.Package, 0)
	found := false

	// Find from build arguments e.g. go build test.go or go build cmd/app
	for i := len(buildCmd) - 1; i >= 0; i-- {
		buildArg := buildCmd[i]

		// Stop canary when we see a build flag or a "build" command
		if strings.HasPrefix("-", buildArg) ||
			buildArg == "build" ||
			buildArg == "install" {
			break
		}

		// Special case. If the file named with test_ prefix, we create a fake
		// package for it. This is a workaround for the case that the test file
		// is compiled with other normal files.
		if strings.HasSuffix(buildArg, ".go") &&
			strings.HasPrefix(buildArg, "test_") {
			artificialPkg := &packages.Package{
				GoFiles: []string{buildArg},
				Name:    "main",
			}
			candidates = append(candidates, artificialPkg)
			found = true
			continue
		}

		// Trying to load package from the build argument, error is tolerated
		// because we dont know what the build argument is. One exception is
		// when we already found packages, in this case, we expect subsequent
		// build arguments are packages, so we should not tolerate any error.
		pkgs, err := tryLoadPackage(buildArg)
		if err != nil {
			if found {
				// If packages are already found, we expect subsequent build
				// arguments are packages, so we should not tolerate any error
				break
			}
			util.Log("Cannot load package from %v", buildArg)
			continue
		}
		for _, pkg := range pkgs {
			if pkg.Errors != nil {
				continue
			}
			found = true
			candidates = append(candidates, pkg)
		}
	}

	// If no import paths are given, the action applies to the package in the
	// current directory.
	if !found {
		pkgs, err := tryLoadPackage(".")
		if err != nil {
			return nil, ex.Error(err)
		}
		for _, pkg := range pkgs {
			if pkg.Errors != nil {
				continue
			}
			candidates = append(candidates, pkg)
		}
	}
	if len(candidates) == 0 {
		return nil, ex.Errorf(nil, "no package found")
	}

	return candidates, nil
}

// Find go.mod from dir and its parent recursively
func findGoMod(dir string) (string, error) {
	for dir != "" {
		mod := filepath.Join(dir, util.GoModFile)
		if util.PathExists(mod) {
			return mod, nil
		}
		par := filepath.Dir(dir)
		if par == dir {
			break
		}
		dir = par
	}
	return "", ex.Errorf(nil, "cannot find go.mod")
}

func (dp *DepProcessor) initCmd() {
	// There is a tricky, all arguments after the otel tool itself are saved for
	// later use, which means the subcommand "go build" itself are also included
	dp.goBuildCmd = make([]string, len(os.Args)-1)
	copy(dp.goBuildCmd, os.Args[1:])
	util.AssertGoBuild(dp.goBuildCmd)
	util.Log("Go build command: %v", dp.goBuildCmd)
}

func findMainDir(pkgs []*packages.Package) (string, error) {
	gofiles := make([]string, 0)
	for _, pkg := range pkgs {
		if pkg.GoFiles == nil {
			continue
		}
		gofiles = append(gofiles, pkg.GoFiles...)
	}
	for _, gofile := range gofiles {
		if !util.IsGoFile(gofile) {
			continue
		}
		root, err := ast.ParseAstFromFileFast(gofile)
		if err != nil {
			return "", ex.Error(err)
		}
		for _, decl := range root.Decls {
			if d, ok := decl.(*dst.FuncDecl); ok && d.Name.Name == "main" {
				// We found the main function, return the directory of the file
				return filepath.Dir(gofile), nil
			}
		}
	}
	return "", ex.Errorf(nil, "cannot find main function in the source files")
}

func (dp *DepProcessor) initMod() (err error) {
	// Find compiling module and package information from the build command
	pkgs, err := findModule(dp.goBuildCmd)
	if err != nil {
		return ex.Error(err)
	}
	util.Log("Find Go packages %v", util.Jsonify(pkgs))
	for _, pkg := range pkgs {
		util.Log("Find Go package %v", util.Jsonify(pkg))
		if pkg.GoFiles == nil {
			continue
		}
		if pkg.Module != nil {
			// Build the module
			// Best case, we find the module information from the package field
			util.Log("Find Go module %v", util.Jsonify(pkg.Module))
			util.Assert(pkg.Module.Path != "", "pkg.Module.Path is empty")
			util.Assert(pkg.Module.GoMod != "", "pkg.Module.GoMod is empty")
			dp.moduleName = pkg.Module.Path
			dp.modulePath = pkg.Module.GoMod
			dir, err := findMainDir(pkgs)
			if err != nil {
				return ex.Error(err)
			}
			dp.otelRuntimeGo = filepath.Join(dir, OtelRuntimeGo)
		} else {
			// Build the source files
			// If we cannot find the module information from the package field,
			// we try to find it from the go.mod file, where go.mod file is in
			// the same directory as the source file.
			util.Assert(pkg.Name != "", "pkg.Name is empty")
			if pkg.Name == "main" {
				gofile := pkg.GoFiles[0]
				gomod, err := findGoMod(filepath.Dir(gofile))
				if err != nil {
					return ex.Error(err)
				}
				util.Assert(gomod != "", "gomod is empty")
				util.Assert(util.PathExists(gomod), "gomod does not exist")
				dp.modulePath = gomod
				// Get module name from go.mod file
				modfile, err := parseGoMod(gomod)
				if err != nil {
					return ex.Error(err)
				}
				dp.moduleName = modfile.Module.Mod.Path
				// We generate additional source file(otel_importer.go) in the
				// same directory as the go.mod file, we should append this file
				// into build commands to make sure it is compiled together with
				// the original source files.
				found := false
				for _, cmd := range dp.goBuildCmd {
					if strings.Contains(cmd, OtelRuntimeGo) {
						found = true
						break
					}
				}
				if !found {
					last := dp.goBuildCmd[len(dp.goBuildCmd)-1]
					dp.otelRuntimeGo = filepath.Join(filepath.Dir(last), OtelRuntimeGo)
					dp.goBuildCmd = append(dp.goBuildCmd, dp.otelRuntimeGo)
				}
			}
		}
	}
	if dp.moduleName == "" || dp.modulePath == "" {
		return ex.Errorf(nil, "cannot find compiled module")
	}
	if dp.otelRuntimeGo == "" {
		return ex.Errorf(nil, "cannot place otel_importer.go file")
	}

	// We will import alibaba-otel/pkg module in generated code, which is not
	// published yet, so we also need to add a replace directive to the go.mod file
	// to tell the go tool to use the local module cache instead of the remote
	// module, that's why we do this here.
	// TODO: Once we publish the alibaba-otel/pkg module, we can remove this code
	// along with the replace directive in the go.mod file.
	dp.pkgLocalCache, err = findModCacheDir()
	if err != nil {
		return ex.Error(err)
	}
	// In the further processing, we will edit the go.mod file, which is illegal
	// to use relative path, so we need to convert the relative path to an absolute
	dp.pkgLocalCache, err = filepath.Abs(dp.pkgLocalCache)
	if err != nil {
		return ex.Error(err)
	}
	if dp.pkgLocalCache == "" {
		return ex.Errorf(nil, "cannot find rule cache dir")
	}
	return nil
}
func (dp *DepProcessor) initBuildMode() {
	// Check if the build mode
	ignoreVendor := false
	for _, arg := range dp.goBuildCmd {
		// -mod=mod and -mod=readonly tells the go command to ignore the vendor
		// directory. We should not use the vendor directory in this case.
		if strings.HasPrefix(arg, "-mod=mod") ||
			strings.HasPrefix(arg, "-mod=readonly") {
			dp.vendorMode = false
			ignoreVendor = true
			break
		}
	}
	if !ignoreVendor {
		// FIXME: vendor directory name can be anything, but we assume it's "vendor"
		// for now
		vendor := filepath.Join(dp.getGoModDir(), VendorDir)
		dp.vendorMode = util.PathExists(vendor)
	}
	// If we are building with vendored dependencies, we should not pull any
	// additional dependencies online, which means all dependencies should be
	// available in the vendor directory. This requires users to add these
	// dependencies proactively
	util.Log("Vendor mode: %v", dp.vendorMode)
}

func (dp *DepProcessor) initSignalHandler() {
	// Register signal handler to catch up SIGINT/SIGTERM interrupt signals and
	// do necessary cleanup
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigc
		switch s {
		case syscall.SIGTERM, syscall.SIGINT:
			util.Log("Interrupted instrumentation, cleaning up")
		default:
		}
	}()
}

func (dp *DepProcessor) init() error {
	dp.initCmd()
	err := dp.initMod()
	if err != nil {
		return ex.Error(err)
	}
	dp.initBuildMode()
	dp.initSignalHandler()
	// Once all the initialization is done, let's log the configuration
	util.Log("ToolVersion: %s", config.ToolVersion)
	return nil
}
