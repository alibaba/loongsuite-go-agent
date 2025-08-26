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

package test

import (
	"fmt"
	"regexp"
	"testing"
)

const WorldAppName = "world"

const expectedImportCounts = 38

func TestCompileTheWorld(t *testing.T) {
	UseApp(WorldAppName)

	RunGoBuild(t, "go", "build")
	RunApp(t, WorldAppName)
	text := ReadLog(t)

	regex := `\"ImportPath\":\"([^"]+)\"`
	r := regexp.MustCompile(regex)

	importPaths := make(map[string]struct{})
	matches := r.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		importPath := match[1]
		importPaths[importPath] = struct{}{}
	}
	if len(importPaths) != expectedImportCounts {
		t.Logf("Expected %d import paths, but found %d", expectedImportCounts, len(importPaths))
		t.Log("Matched import paths:")
		// sort import paths for better readability
		// (not strictly necessary, but helps in debugging)
		for path := range importPaths {
			fmt.Println(path)
		}
		t.Fatalf("Rule matching is not complete")
	}
}
