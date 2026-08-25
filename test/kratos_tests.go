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

const kratos_dependency_name = "github.com/go-kratos/kratos/v2"
const kratos_module_name = "kratos"

const kratos_v3_dependency_name = "github.com/go-kratos/kratos/v3"
const kratos_v3_module_name = "kratos"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("kratos-http-test", kratos_module_name, "", "", "1.18", "", TestKratosHttp),
		NewGeneralTestCase("kratos-grpc-test", kratos_module_name, "", "", "1.18", "", TestKratosGrpc),
		NewLatestDepthTestCase("kratos-latest-depth-grpc", kratos_dependency_name, kratos_module_name, "v2.6.3", "", "1.22", "", TestKratosGrpc),
		NewLatestDepthTestCase("kratos-latest-depth-http", kratos_dependency_name, kratos_module_name, "v2.6.3", "", "1.22", "", TestKratosHttp),
		NewMuzzleTestCase("kratos-muzzle-grpc", kratos_dependency_name, kratos_module_name, "v2.6.3", "", "1.22", "", []string{"go", "build", "test_kratos_grpc.go", "server.go"}),
		NewMuzzleTestCase("kratos-muzzle-http", kratos_dependency_name, kratos_module_name, "v2.6.3", "", "1.22", "", []string{"go", "build", "test_kratos_http.go", "server.go"}),

		NewGeneralTestCase("kratos-v3-http-test", kratos_v3_module_name, "", "", "1.25", "", TestKratosHttpV3),
		NewGeneralTestCase("kratos-v3-grpc-test", kratos_v3_module_name, "", "", "1.25", "", TestKratosGrpcV3),
		NewLatestDepthTestCase("kratos-v3-latest-depth-grpc", kratos_v3_dependency_name, kratos_v3_module_name, "v3.0.0", "", "1.25", "", TestKratosGrpcV3),
		NewLatestDepthTestCase("kratos-v3-latest-depth-http", kratos_v3_dependency_name, kratos_v3_module_name, "v3.0.0", "", "1.25", "", TestKratosHttpV3),
		NewMuzzleTestCase("kratos-v3-muzzle-grpc", kratos_v3_dependency_name, kratos_v3_module_name, "v3.0.0", "", "1.25", "", []string{"go", "build", "test_kratos_grpc.go", "server.go"}),
		NewMuzzleTestCase("kratos-v3-muzzle-http", kratos_v3_dependency_name, kratos_v3_module_name, "v3.0.0", "", "1.25", "", []string{"go", "build", "test_kratos_http.go", "server.go"}),
	)
}

func TestKratosGrpc(t *testing.T, env ...string) {
	UseApp("kratos/v2.6.3")
	RunGoBuild(t, "go", "build", "test_kratos_grpc.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_grpc", env...)
}

func TestKratosHttp(t *testing.T, env ...string) {
	UseApp("kratos/v2.6.3")
	RunGoBuild(t, "go", "build", "test_kratos_http.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_http", env...)
}

func TestKratosGrpcV3(t *testing.T, env ...string) {
	UseApp("kratos/v3.0.0")
	RunGoBuild(t, "go", "build", "test_kratos_grpc.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_grpc", env...)
}

func TestKratosHttpV3(t *testing.T, env ...string) {
	UseApp("kratos/v3.0.0")
	RunGoBuild(t, "go", "build", "test_kratos_http.go", "server.go")
	env = append(env, "OTEL_INSTRUMENTATION_KRATOS_EXPERIMENTAL_SPAN_ENABLE=true")
	RunApp(t, "test_kratos_http", env...)
}
