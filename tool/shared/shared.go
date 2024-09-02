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

package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alibaba/opentelemetry-go-auto-instrumentation/tool/util"
)

const GoBuildIgnoreComment = "//go:build ignore"

const GoModFile = "go.mod"

const DebugLogFile = "debug.log"

const (
	TInstrument = "instrument"
	TPreprocess = "preprocess"
)

func IsCompileCommand(line string) bool {
	return strings.Contains(line, "compile -o") &&
		strings.Contains(line, "buildid")
}

func GetLogPath(name string) string {
	if InToolexec {
		return filepath.Join(TempBuildDir, TInstrument, name)
	} else {
		return filepath.Join(TempBuildDir, TPreprocess, name)
	}
}

func GetInstrumentLogPath(name string) string {
	return filepath.Join(TempBuildDir, TInstrument, name)
}

func GetPreprocessLogPath(name string) string {
	return filepath.Join(TempBuildDir, TPreprocess, name)
}

func GetVarNameOfFunc(fn string) string {
	const varDeclSuffix = "Impl"
	fn = strings.Title(fn)
	return fn + varDeclSuffix
}

func SaveDebugFile(prefix string, path string) {
	targetName := filepath.Base(path)
	util.Assert(IsGoFile(targetName), "sanity check")
	counterpart := GetLogPath("debug_" + prefix + targetName)
	_ = util.CopyFile(path, counterpart)
}

var packageRegexp = regexp.MustCompile(`(?m)^package\s+\w+`)

func RenamePackage(source, newPkgName string) string {
	source = packageRegexp.ReplaceAllString(source, fmt.Sprintf("package %s\n", newPkgName))
	return source
}

func RemoveGoBuildComment(text string) string {
	text = strings.ReplaceAll(text, GoBuildIgnoreComment, "")
	return text
}

func HasGoBuildComment(text string) bool {
	return strings.Contains(text, GoBuildIgnoreComment)
}

// GetGoModPath returns the absolute path of go.mod file, if any.
func GetGoModPath() (string, error) {
	// @@ As said in the comment https://github.com/golang/go/issues/26500, the
	// expected way to get go.mod should be go list -m -f {{.GoMod}}, but it does
	// not work well when go.work presents, we use go env GOMOD instead.
	//
	// go env GOMOD
	// The absolute path to the go.mod of the main module.
	// If module-aware mode is enabled, but there is no go.mod, GOMOD will be
	// os.DevNull ("/dev/null" on Unix-like systems, "NUL" on Windows).
	// If module-aware mode is disabled, GOMOD will be the empty string.
	cmd := exec.Command("go", "env", "GOMOD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get go.mod directory: %w", err)
	}
	path := strings.TrimSpace(string(out))
	return path, nil
}

func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

func IsExistGoMod() (bool, error) {
	gomod, err := GetGoModPath()
	if err != nil {
		return false, fmt.Errorf("failed to get go.mod path: %w", err)
	}
	if gomod == "" {
		return false, errors.New("failed to get go.mod path: not module-aware")
	}
	return strings.HasSuffix(gomod, GoModFile), nil
}

func HashStruct(st interface{}) (uint64, error) {
	bs, err := json.Marshal(st)
	if err != nil {
		return 0, err
	}
	hasher := fnv.New64a()
	_, err = hasher.Write(bs)
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func InPreprocess() bool {
	return !InToolexec
}

func InInstrument() bool {
	return InToolexec
}

func GuaranteeInPreprocess() {
	util.Assert(!InToolexec, "not in preprocess stage")
}

func GuaranteeInInstrument() {
	util.Assert(InToolexec, "not in instrument stage")
}
