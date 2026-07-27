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

import "testing"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("gin-test-html", "gin", "", "", "1.21", "", TestGinHTML),
		NewGeneralTestCase("gin-test-pattern", "gin", "", "", "1.21", "", TestGinPattern),
		NewGeneralTestCase("gin-test-static", "gin", "", "", "1.21", "", TestGinStatic),
	)
}

func TestGinHTML(t *testing.T, env ...string) {
	UseApp("gin")
	RunGoBuild(t, "go", "build", "test_gin_html.go")
	RunApp(t, "test_gin_html", env...)
}

func TestGinPattern(t *testing.T, env ...string) {
	UseApp("gin")
	RunGoBuild(t, "go", "build", "test_gin_pattern.go")
	RunApp(t, "test_gin_pattern", env...)
}

func TestGinStatic(t *testing.T, env ...string) {
	UseApp("gin")
	RunGoBuild(t, "go", "build", "test_gin_static.go")
	RunApp(t, "test_gin_static", env...)
}
