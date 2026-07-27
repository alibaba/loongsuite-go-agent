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
		NewGeneralTestCase("gorestful-test-pattern", "gorestful", "", "", "1.21", "", TestGoRestfulPattern),
		NewGeneralTestCase("gorestful-test-handle-with-filter", "gorestful", "", "", "1.21", "", TestGoRestfulHandleWithFilter),
	)
}

func TestGoRestfulPattern(t *testing.T, env ...string) {
	UseApp("gorestful")
	RunGoBuild(t, "go", "build", "test_restful_pattern.go")
	RunApp(t, "test_restful_pattern", env...)
}

func TestGoRestfulHandleWithFilter(t *testing.T, env ...string) {
	UseApp("gorestful")
	RunGoBuild(t, "go", "build", "test_restful_handle_with_filter.go")
	RunApp(t, "test_restful_handle_with_filter", env...)
}
