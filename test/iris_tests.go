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
		NewGeneralTestCase("iris-test", "iris", "", "", "1.21", "", TestIris),
		NewGeneralTestCase("iris-pattern-test", "iris", "", "", "1.21", "", TestIrisPattern),
	)
}

func TestIris(t *testing.T, env ...string) {
	UseApp("iris")
	RunGoBuild(t, "go", "build", "test_iris.go")
	RunApp(t, "test_iris", env...)
}

func TestIrisPattern(t *testing.T, env ...string) {
	UseApp("iris")
	RunGoBuild(t, "go", "build", "test_iris_pattern.go")
	RunApp(t, "test_iris_pattern", env...)
}
