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
	"bufio"
	"encoding/json"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alibaba/loongsuite-go/tool/ast"
	"github.com/alibaba/loongsuite-go/tool/config"
	"github.com/alibaba/loongsuite-go/tool/data"
	"github.com/alibaba/loongsuite-go/tool/ex"
	"github.com/alibaba/loongsuite-go/tool/rules"
	"github.com/alibaba/loongsuite-go/tool/util"
	"github.com/dave/dst"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/errgroup"
)

type ruleMatcher struct {
	availableRules    map[string][]rules.InstRule
	moduleVersions    []*vendorModule     // vendor used only
	projectDeps       map[string]bool     // actual dependencies from dry run commands
	targetIdentifiers map[string][]string // map importPath to target identifiers for pre-filtering
}

func newRuleMatcher(compileCmds []string) *ruleMatcher {
	availableRules := make(map[string][]rules.InstRule)
	targetIdentifiers := make(map[string][]string)

	for _, rule := range findAvailableRules() {
		// Precompute target identifiers for each import path
		importPath := rule.GetImportPath()
		switch r := rule.(type) {
		case *rules.InstFuncRule:
			// Add function name as a target identifier
			targetIdentifiers[importPath] = append(targetIdentifiers[importPath], r.Function)
			// Add receiver type if present
			if r.ReceiverType != "" {
				targetIdentifiers[importPath] = append(targetIdentifiers[importPath], r.ReceiverType)
			}
		case *rules.InstStructRule:
			// Add struct type as a target identifier
			targetIdentifiers[importPath] = append(targetIdentifiers[importPath], r.StructType)
		}

		availableRules[importPath] = append(availableRules[importPath], rule)
	}
	if config.GetConf().Verbose {
		util.Log("Available rules: %v", availableRules)
	}

	// Populated projectDeps from compileCmds
	projectDeps := populateDependenciesFromCmd(compileCmds)

	return &ruleMatcher{
		availableRules: availableRules,
		projectDeps:    projectDeps,
	}
}

// populateDependenciesFromCmd extracts import paths from the compile commands
func populateDependenciesFromCmd(compileCmds []string) map[string]bool {
	projectDeps := make(map[string]bool)

	for _, cmd := range compileCmds {
		cmdArgs := util.SplitCompileCmds(cmd)
		importPath := util.FindFlagValue(cmdArgs, util.BuildPattern)
		util.Assert(importPath != "", "sanity check")
		projectDeps[importPath] = true
	}

	if config.GetConf().Verbose {
		util.Log("Project dependencies from dry run commands: %v", projectDeps)
	}

	return projectDeps
}

func loadRuleFile(path string) ([]rules.InstRule, error) {
	content, err := util.ReadFile(path)
	if err != nil {
		currentDir, _ := os.Getwd()
		return nil, ex.Wrapf(err, "pwd %s", currentDir)
	}
	return loadRuleRaw(content)
}

func loadRuleRaw(content string) ([]rules.InstRule, error) {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal([]byte(content), &rawMessages); err != nil {
		return nil, ex.Wrap(err)
	}

	rulesSlice := make([]rules.InstRule, 0, len(rawMessages))
	for _, raw := range rawMessages {
		var obj map[string]interface{}
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, ex.Wrap(err)
		}

		var rule rules.InstRule
		if _, ok := obj["Function"]; ok {
			var funcRule rules.InstFuncRule
			if err := json.Unmarshal(raw, &funcRule); err != nil {
				return nil, ex.Wrap(err)
			}
			rule = &funcRule
		} else if _, ok := obj["StructType"]; ok {
			var structRule rules.InstStructRule
			if err := json.Unmarshal(raw, &structRule); err != nil {
				return nil, ex.Wrap(err)
			}
			rule = &structRule
		} else if _, ok := obj["FileName"]; ok {
			var fileRule rules.InstFileRule
			if err := json.Unmarshal(raw, &fileRule); err != nil {
				return nil, ex.Wrap(err)
			}
			rule = &fileRule
		} else {
			return nil, ex.Newf("unknown rule type: %s", string(raw))
		}
		rulesSlice = append(rulesSlice, rule)
	}

	return rulesSlice, nil
}

type chunk []rules.InstRule

func loadDefaultRules() []rules.InstRule {
	// Read all default embedded rule files
	files, err := data.ListRuleFiles()
	if err != nil {
		util.Log("Failed to list default rule json files: %v", err)
		return nil
	}

	// Disable specific rules if specified
	filteredFiles := make([]string, 0)
	disable := config.GetConf().GetDisabledRules()
	switch disable {
	case "all":
		// Disable all rules except base.json
		filteredFiles = append(filteredFiles, "base.json")
	case "":
		// Enable all rules
		filteredFiles = files
	default:
		// Disable specific rules
		disabledRules := strings.Split(disable, ",")
		for _, name := range files {
			found := false
			for _, disabled := range disabledRules {
				if disabled == name {
					found = true
				}
			}
			if !found {
				filteredFiles = append(filteredFiles, name)
			}
		}
	}

	// Load and parse each rule file concurrently
	ruleChunks := make([]chunk, len(filteredFiles))
	group := &errgroup.Group{}
	foundBase := false
	for i, name := range filteredFiles {
		i, name := i, name // capture loop variables
		if name == "base.json" {
			foundBase = true
		}

		group.Go(func() error {
			raw, err := data.ReadRuleFile(name)
			if err != nil {
				util.Log("Failed to read rule file %s: %v", name, err)
				return err
			}

			// Parse JSON content into InstRule slice
			rule, err := loadRuleRaw(string(raw))
			if err != nil {
				util.Log("Failed to parse rule file %s: %v", name, err)
				return nil
			}

			ruleChunks[i] = rule
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		util.Log("One or more default rule files failed to load: %v", err)
		return nil
	}
	if !foundBase {
		util.Log("base.json is not found in the default rule files")
		return nil
	}

	// Merge all ruleChunks
	rules := make([]rules.InstRule, 0)
	for _, c := range ruleChunks {
		rules = append(rules, c...)
	}
	return rules
}

func findAvailableRules() []rules.InstRule {
	rules := make([]rules.InstRule, 0)

	// Load default rules (filtering is handled inside loadDefaultRules)
	defaultRules := loadDefaultRules()
	rules = append(rules, defaultRules...)

	// If rule files are provided, load them
	if config.GetConf().RuleJsonFiles != "" {
		// Load multiple rule files
		if strings.Contains(config.GetConf().RuleJsonFiles, ",") {
			ruleFiles := strings.Split(config.GetConf().RuleJsonFiles, ",")
			for _, ruleFile := range ruleFiles {
				r, err := loadRuleFile(ruleFile)
				if err != nil {
					util.Log("Failed to load rules: %v", err)
					continue
				}
				rules = append(rules, r...)
			}
			return rules
		}
		// Load the one rule file
		rs, err := loadRuleFile(config.GetConf().RuleJsonFiles)
		if err != nil {
			util.Log("Failed to load rules: %v", err)
			return nil
		}
		rules = append(rules, rs...)
	}
	return rules
}

var versionRegexp = regexp.MustCompile(`@v\d+\.\d+\.\d+(-.*?)?/`)

func extractVersion(path string) string {
	// Unify the path to Unix style
	path = filepath.ToSlash(path)
	version := versionRegexp.FindString(path)
	if version == "" {
		return ""
	}
	// Extract version number from the string
	return version[1 : len(version)-1]
}

// splitVersionRange splits the version range into two parts, start and end.
func splitVersionRange(vr string) (string, string) {
	util.Assert(strings.Contains(vr, ","), "invalid version range format")
	util.Assert(strings.Contains(vr, "["), "invalid version range format")
	util.Assert(strings.Contains(vr, ")"), "invalid version range format")

	start := vr[1:strings.Index(vr, ",")]
	end := vr[strings.Index(vr, ",")+1 : len(vr)-1]
	return "v" + start, "v" + end
}

// matchVersion checks if the version string matches the version range in the
// rule. The version range is in format [start, end), where start is inclusive
// and end is exclusive. If the rule version string is empty, it always matches.
func matchVersion(version string, ruleVersion string) (bool, error) {
	// Fast path, always match if the rule version is not specified
	if ruleVersion == "" {
		return true, nil
	}
	// Check if both rule version and package version are in sane
	if !strings.Contains(version, "v") {
		return false, ex.Newf("invalid version %v", version)
	}
	if !strings.Contains(ruleVersion, "[") ||
		!strings.Contains(ruleVersion, ")") ||
		!strings.Contains(ruleVersion, ",") ||
		strings.Contains(ruleVersion, "v") {
		return false, ex.Newf("invalid rule version %v", ruleVersion)
	}
	// Remove extra whitespace from the rule version string
	ruleVersion = strings.ReplaceAll(ruleVersion, " ", "")

	// Compare the version with the rule version, the rule version is in the
	// format [start, end), where start is inclusive and end is exclusive
	// and start or end can be omitted, which means the range is open-ended.
	ruleVersionStart, ruleVersionEnd := splitVersionRange(ruleVersion)
	switch {
	case ruleVersionStart != "v" && ruleVersionEnd != "v":
		// Full version range
		if semver.Compare(version, ruleVersionStart) >= 0 &&
			semver.Compare(version, ruleVersionEnd) < 0 {
			return true, nil
		}
	case ruleVersionStart == "v":
		// Only end is specified
		util.Assert(ruleVersionEnd != "v", "sanity check")
		if semver.Compare(version, ruleVersionEnd) < 0 {
			return true, nil
		}
	case ruleVersionEnd == "v":
		// Only start is specified
		util.Assert(ruleVersionStart != "v", "sanity check")
		if semver.Compare(version, ruleVersionStart) >= 0 {
			return true, nil
		}
	default:
		return false, ex.Newf("invalid rule version range %v", ruleVersion)
	}
	return false, nil
}

// matchDependencies checks if all required dependencies are present in the project
// Only InstFuncRule supports dependencies checking
func (rm *ruleMatcher) matchDependencies(rule rules.InstRule) bool {
	funcRule, ok := rule.(*rules.InstFuncRule)
	if !ok {
		return true
	}

	dependencies := funcRule.Dependencies
	if len(dependencies) == 0 {
		return true // No dependencies required
	}

	for _, dep := range dependencies {
		if !rm.projectDeps[dep] {
			if config.GetConf().Verbose {
				util.Log("Dependency %s not found for rule %s", dep, rule.GetImportPath())
			}
			return false
		}
	}

	return true
}

// match gives compilation arguments and finds out all interested rules
// for it.
func (rm *ruleMatcher) match(cmdArgs []string) *rules.InstRuleSet {
	importPath := util.FindFlagValue(cmdArgs, util.BuildPattern)
	util.Assert(importPath != "", "sanity check")
	if config.GetConf().Verbose {
		util.Log("RunMatch: %v (%v)", importPath, cmdArgs)
	}
	availables := make([]rules.InstRule, len(rm.availableRules[importPath]))

	// Okay, we are interested in these candidates, let's read it and match with
	// the instrumentation rule, but first we need to check if the package name
	// are already registered, to avoid futile effort
	copy(availables, rm.availableRules[importPath])
	if len(availables) == 0 {
		return nil // fast fail
	}
	// Early filtering: filter rules based on dependencies before processing any files
	filteredAvailables := make([]rules.InstRule, 0, len(availables))
	for _, rule := range availables {
		if rm.matchDependencies(rule) {
			filteredAvailables = append(filteredAvailables, rule)
		}
	}

	if len(filteredAvailables) == 0 {
		return nil // no rules match dependencies
	}

	parsedAst := make(map[string]*dst.File)
	bundle := rules.NewInstRuleSet(importPath)

	goVersion := util.FindFlagValue(cmdArgs, util.BuildGoVer)
	util.Assert(goVersion != "", "sanity check")
	util.Assert(strings.HasPrefix(goVersion, "go"), "sanity check")
	goVersion = strings.Replace(goVersion, "go", "v", 1)
	for _, candidate := range cmdArgs {
		// Rewrite the quoted file path to the unquoted one, this usually happens
		// on Windows, for example, "C:\\Program Files\\abc.go" will be rewritten
		// to C:\\Program Files\\abc.go
		candidate = strings.Trim(candidate, `"`)

		// It's not a go file, ignore silently
		if !util.IsGoFile(candidate) {
			continue
		}
		const cgoMark = "_cgo_import.go"
		if strings.HasSuffix(candidate, cgoMark) {
			bundle.HasCgo = true
			continue
		}
		file, err := filepath.Abs(candidate)
		if err != nil {
			util.Log("Failed to get absolute path of file %s: %v", candidate, err)
			continue
		}
		// If the file is generated by cgo at compilation time, ignore it
		if !util.PathExists(file) {
			util.Log("File %s does not exist, ignore", file)
			continue
		}

		// Pre-filtering: check if the file contains any target identifiers before parsing AST
		targetIdentifiers := rm.targetIdentifiers[importPath]
		if len(targetIdentifiers) > 0 && !rm.hasTargetIdentifiers(file, targetIdentifiers) {
			continue // Skip this file as it doesn't contain any target identifiers
		}

		// If it's a vendor build, we need to extract the version of the module
		// from vendor/modules.txt, otherwise we find the version from source
		// code file path
		version := extractVersion(file)
		if rm.moduleVersions != nil {
			recorded := findVendorModuleVersion(rm.moduleVersions, importPath)
			if recorded != "" {
				version = recorded
			}
		}

		// Fair enough, parse the file content. Since this is a heavy operation,
		// we cache the parsed AST to avoid redundant parsing.
		var tree *dst.File
		if _, ok := parsedAst[file]; !ok {
			fileAst, err := ast.ParseFileFast(file)
			if fileAst == nil || err != nil {
				util.Log("failed to parse file %s: %v", file, err)
				continue
			}
			parsedAst[file] = fileAst
			util.Assert(fileAst.Name.Name != "", "empty package name")
			bundle.SetPackageName(fileAst.Name.Name)
			tree = fileAst
		} else {
			tree = parsedAst[file]
		}

		if tree == nil {
			// Failed to parse the file, stop here and log only
			// since it's a tolerant failure
			util.Log("Failed to parse file %s", file)
			continue
		}

		for i := len(filteredAvailables) - 1; i >= 0; i-- {
			rule := filteredAvailables[i]

			// Check if the version is supported
			matched, err := matchVersion(version, rule.GetVersion())
			if err != nil {
				util.Log("Bad match: file %s, rule %s, version %s",
					file, rule, version)
				continue
			}
			if !matched {
				continue
			}
			// Check if the rule requires a specific Go version(range)
			if rule.GetGoVersion() != "" {
				matched, err = matchVersion(goVersion, rule.GetGoVersion())
				if err != nil {
					util.Log("Bad match: file %s, rule %s, go version %s",
						file, rule, goVersion)
					continue
				}
				if !matched {
					continue
				}
			}

			// Let's match with the rule precisely
			valid := false
			switch rl := rule.(type) {
			case *rules.InstFuncRule:
				funcDecls := ast.FindFuncDecl(tree, rl.Function, rl.ReceiverType)
				if len(funcDecls) > 0 {
					bundle.AddFuncRule(file, rl)
					util.Log("Match func rule %s with %v", rule, cmdArgs)
					valid = true
				}
			case *rules.InstStructRule:
				genDecl := ast.FindStructDecl(tree, rl.StructType)
				if genDecl != nil {
					bundle.AddStructRule(file, rl)
					util.Log("Match struct rule %s with %v", rule, cmdArgs)
					valid = true
				}
			case *rules.InstFileRule:
				// File rule is always matched
				util.Log("Match file rule %s with %v", rule, cmdArgs)
				bundle.AddFileRule(rl)
				bundle.SetPackageName(tree.Name.Name)
				valid = true
			default:
				util.ShouldNotReachHere()
			}
			if valid {
				// Remove the rule from the available rules
				filteredAvailables = append(filteredAvailables[:i], filteredAvailables[i+1:]...)
			}
		}
	}
	return bundle
}

// vendorModule represents a module in vendor/modules.txt file, it contains
// the module name, version and all submodules of the module, which looks like
//
// # golang.org/x/text v0.21.0
// ## explicit; go 1.18
// golang.org/x/text/secure/bidirule
// golang.org/x/text/transform
// golang.org/x/text/unicode/bidi
// golang.org/x/text/unicode/norm
// # golang.org/x/time v0.5.0
// ## explicit; go 1.18
// golang.org/x/time/rate
//
// The module name is the first line of the module, the version is the second
// part of the first line, and all submodules are listed in the following lines
// starting with the module name.
type vendorModule struct {
	path       string
	version    string
	submodules []string
}

func findVendorModuleVersion(modules []*vendorModule, importPath string) string {
	for _, module := range modules {
		if module.path == importPath {
			return module.version
		}
		for _, submodule := range module.submodules {
			if submodule == importPath {
				return module.version
			}
		}
	}
	return ""
}

func cutPrefix(s, prefix string) (after string, found bool) {
	// Compatible with go1.18 as we use this version internally
	if !strings.HasPrefix(s, prefix) {
		return s, false
	}
	return s[len(prefix):], true
}

//nolint:staticcheck // verbatim copy from go source
func parseVendorModules(projDir string) ([]*vendorModule, error) {
	vendorFile := filepath.Join(projDir, "vendor", "modules.txt")
	if util.PathNotExists(vendorFile) {
		return nil, ex.Newf("vendor/modules.txt not found")
	}
	file, err := os.Open(vendorFile)
	if err != nil {
		return nil, ex.Wrap(err)
	}
	defer func(dryRunLog *os.File) {
		err := dryRunLog.Close()
		if err != nil {
			util.Log("Failed to close dry run log file: %v", err)
		}
	}(file)

	scanner := bufio.NewScanner(file)
	// 10MB should be enough to accommodate most long line
	buffer := make([]byte, 0, 10*1024*1024)
	scanner.Buffer(buffer, cap(buffer))

	vms := make([]*vendorModule, 0)
	mod := &vendorModule{}
	vendorVersion := make(map[string]string)
	// From src/cmd/go/internal/modload/vendor.go
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			f := strings.Fields(line)

			if len(f) < 3 {
				continue
			}
			if semver.IsValid(f[2]) {
				// A module, but we don't yet know whether it is in the build list or
				// only included to indicate a replacement.
				mod = &vendorModule{path: f[1], version: f[2]}
				f = f[3:]
			} else if f[2] == "=>" {
				// A wildcard replacement found in the main module's go.mod file.
				mod = &vendorModule{path: f[1]}
				f = f[2:]
			} else {
				// Not a version or a wildcard replacement.
				// We don't know how to interpret this module line, so ignore it.
				mod = &vendorModule{}
				continue
			}
			if len(f) >= 2 && f[0] == "=>" {
				// Skip replacement lines
			}
			continue
		}

		// Not a module line. Must be a package within a module or a metadata
		// directive, either of which requires a preceding module line.
		if mod.path == "" {
			continue
		}

		if _, ok := cutPrefix(line, "## "); ok {
			// Skip annotations lines
			continue
		}

		if f := strings.Fields(line); len(f) == 1 && module.CheckImportPath(f[0]) == nil {
			// A package within the current module.
			mod.submodules = append(mod.submodules, f[0])

			// Since this module provides a package for the build, we know that it
			// is in the build list and is the selected version of its path.
			// If this information is new, record it.
			if v, ok := vendorVersion[mod.path]; !ok || semver.Compare(v, mod.version) < 0 {
				vms = append(vms, mod)
				vendorVersion[mod.path] = mod.version
			}
		}
	}
	err = scanner.Err()
	if err != nil {
		return nil, ex.Wrapf(err, "cannot parse vendor/modules.txt")
	}
	return vms, nil
}

func runMatch(matcher *ruleMatcher, cmd string, ch chan *rules.InstRuleSet) {
	bundle := matcher.match(util.SplitCompileCmds(cmd))
	ch <- bundle
}

func (dp *DepProcessor) matchRules() ([]*rules.InstRuleSet, error) {
	defer util.PhaseTimer("Match")()
	compileCmds, err := dp.findDeps()
	if err != nil {
		return nil, err
	}

	matcher := newRuleMatcher(compileCmds)

	// If we are in vendor mode, we need to parse the vendor/modules.txt file
	// to get the version of each module for future matching
	if dp.vendorMode {
		modules, err := parseVendorModules(dp.getGoModDir())
		if err != nil {
			return nil, err
		}
		if config.GetConf().Verbose {
			util.Log("Vendor modules: %v", modules)
		}
		matcher.moduleVersions = modules
	}

	// Find used instrumentation rule according to compile commands
	ch := make(chan *rules.InstRuleSet)
	for _, cmd := range compileCmds {
		go runMatch(matcher, cmd, ch)
	}
	cnt := 0
	bundles := make([]*rules.InstRuleSet, 0)
	for cnt < len(compileCmds) {
		bundle := <-ch
		if bundle.IsValid() {
			bundles = append(bundles, bundle)
		}
		cnt++
	}
	return bundles, nil
}

func (rm *ruleMatcher) hasTargetIdentifiers(filePath string, targetIdentifiers []string) bool {
	if len(targetIdentifiers) == 0 {
		return true // If no targets, consider all files as matching
	}

	src, err := os.ReadFile(filePath)
	if err != nil {
		util.Log("Failed to read file: %v", err)
		return false
	}

	var scanner scanner.Scanner
	var fs token.FileSet

	f := fs.AddFile("", fs.Base(), len(src))
	scanner.Init(f, src, nil, 0)

	// Scan through all tokens in the file
	for {
		_, tok, lit := scanner.Scan() // ignore pos as we don't need it
		if tok == token.EOF {
			break
		}

		// Check if the token is an identifier that matches any target
		if tok == token.IDENT {
			for _, target := range targetIdentifiers {
				if lit == target {
					return true // Found a target identifier
				}
			}
		}
	}

	return false
}
