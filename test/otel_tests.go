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
		NewGeneralTestCase("otel-span-from-context-test", "otel", "", "", "1.18", "", TestSpanFromContext),
	)
}

func TestSpanFromContext(t *testing.T, env ...string) {
	UseApp("otel")
	RunGoBuild(t, "go", "build", "--", "test_span_from_context.go")
	stdout, _ := RunApp(t, "test_span_from_context", env...)
	ExpectContains(t, stdout, " GET ")
}
