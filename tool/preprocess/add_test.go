// Copyright (c) 2026 Alibaba Group Holding Ltd.
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
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create test dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}

func findReplaceVersion(t *testing.T, gomod, path string) (string, bool) {
	t.Helper()
	modfile, err := parseGoMod(gomod)
	if err != nil {
		t.Fatalf("failed to parse go.mod: %v", err)
	}
	for _, replace := range modfile.Replace {
		if replace.Old.Path == path {
			return replace.New.Version, true
		}
	}
	return "", false
}

func findRequireVersion(t *testing.T, gomod, path string) (string, bool) {
	t.Helper()
	modfile, err := parseGoMod(gomod)
	if err != nil {
		t.Fatalf("failed to parse go.mod: %v", err)
	}
	for _, require := range modfile.Require {
		if require.Mod.Path == path {
			return require.Mod.Version, true
		}
	}
	return "", false
}

func TestPinConflictingHookDependenciesUsesOriginalUserVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if version != "v1.0.0" {
		t.Fatalf("hook require version = %q, want %q", version, "v1.0.0")
	}
	if _, ok := findReplaceVersion(t, gomod, "example.com/dep"); ok {
		t.Fatalf("did not expect user go.mod replace for example.com/dep")
	}
}

func TestPinConflictingHookDependenciesIgnoresIndirectUserVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0 // indirect
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if version != "v1.2.3" {
		t.Fatalf("hook require version = %q, want %q", version, "v1.2.3")
	}
}

func TestPinConflictingHookDependenciesIgnoresPreV1Version(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v0.3.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v0.4.0
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if version != "v0.4.0" {
		t.Fatalf("hook require version = %q, want %q", version, "v0.4.0")
	}
}

func TestPinConflictingHookDependenciesIgnoresIndirectHookVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3 // indirect
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if version != "v1.2.3" {
		t.Fatalf("hook require version = %q, want %q", version, "v1.2.3")
	}
}

func TestPinConflictingHookDependenciesPreservesExistingReplace(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0

replace example.com/dep => ./localdep
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findReplaceVersion(t, gomod, "example.com/dep")
	if !ok {
		t.Fatalf("expected existing replace for example.com/dep")
	}
	if version != "" {
		t.Fatalf("replace version = %q, want existing local replace without version", version)
	}
	hookVersion, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if hookVersion != "v1.2.3" {
		t.Fatalf("hook require version = %q, want %q", hookVersion, "v1.2.3")
	}
}

func TestPinConflictingHookDependenciesIgnoresNonConflictingDependency(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/userdep v1.0.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/hookdep v1.2.3
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	if _, ok := findReplaceVersion(t, gomod, "example.com/hookdep"); ok {
		t.Fatalf("did not expect replace for non-conflicting hook dependency")
	}
}

func TestPinConflictingHookDependenciesPinsMultipleHooks(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir1 := filepath.Join(dir, "hook1")
	hookDir2 := filepath.Join(dir, "hook2")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0
`
	hookMod1 := `module example.com/hook1

go 1.24.0

require example.com/dep v1.2.3
`
	hookMod2 := `module example.com/hook2

go 1.24.0

require example.com/dep v1.4.5
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir1, "go.mod"), hookMod1)
	writeTestFile(t, filepath.Join(hookDir2, "go.mod"), hookMod2)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{
		{
			ImportPath:  "example.com/hook1",
			Replace:     true,
			ReplacePath: hookDir1,
		},
		{
			ImportPath:  "example.com/hook2",
			Replace:     true,
			ReplacePath: hookDir2,
		},
	})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	for _, hookDir := range []string{hookDir1, hookDir2} {
		version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
		if !ok {
			t.Fatalf("expected hook require for example.com/dep")
		}
		if version != "v1.0.0" {
			t.Fatalf("hook require version = %q, want %q", version, "v1.0.0")
		}
	}
}

func TestPinConflictingHookDependenciesPinsMultipleDepsInOneHook(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require (
	example.com/dep1 v1.0.0
	example.com/dep2 v1.1.0
)
`
	hookMod := `module example.com/hook

go 1.24.0

require (
	example.com/dep1 v1.2.3
	example.com/dep2 v1.4.5
)
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	tests := map[string]string{
		"example.com/dep1": "v1.0.0",
		"example.com/dep2": "v1.1.0",
	}
	for path, want := range tests {
		version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), path)
		if !ok {
			t.Fatalf("expected hook require for %s", path)
		}
		if version != want {
			t.Fatalf("hook require version for %s = %q, want %q", path, version, want)
		}
	}
}

func TestPinConflictingHookDependenciesIgnoresDifferentMajorVersion(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.5.0
`
	hookMod := `module example.com/hook

go 1.24.0

require example.com/dep v2.0.0+incompatible
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), hookMod)

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err != nil {
		t.Fatalf("pinConflictingHookDependencies() error = %v", err)
	}

	version, ok := findRequireVersion(t, filepath.Join(hookDir, "go.mod"), "example.com/dep")
	if !ok {
		t.Fatalf("expected hook require for example.com/dep")
	}
	if version != "v2.0.0+incompatible" {
		t.Fatalf("hook require version = %q, want %q", version, "v2.0.0+incompatible")
	}
}

func TestPinConflictingHookDependenciesReturnsMalformedHookGoModError(t *testing.T) {
	dir := t.TempDir()
	gomod := filepath.Join(dir, "go.mod")
	originalGoMod := filepath.Join(dir, "go.mod.bk")
	hookDir := filepath.Join(dir, "hook")

	userMod := `module example.com/app

go 1.24.0

require example.com/dep v1.0.0
`
	writeTestFile(t, gomod, userMod)
	writeTestFile(t, originalGoMod, userMod)
	writeTestFile(t, filepath.Join(hookDir, "go.mod"), "module example.com/hook\n\nrequire (\n")

	dp := &DepProcessor{
		backups: map[string]string{gomod: originalGoMod},
	}
	err := dp.pinConflictingHookDependencies(gomod, []Dependency{{
		ImportPath:  "example.com/hook",
		Replace:     true,
		ReplacePath: hookDir,
	}})
	if err == nil {
		t.Fatalf("pinConflictingHookDependencies() error = nil, want parse error")
	}
}
