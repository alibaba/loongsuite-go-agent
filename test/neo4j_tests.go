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

package test

import (
	"context"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const neo4j_dependency_name = "github.com/neo4j/neo4j-go-driver/v6"
const neo4j_module_name = "neo4j"

func init() {
	TestCases = append(TestCases, NewGeneralTestCase("test_neo4j_basic", neo4j_module_name, "6.2.0", "", "1.23", "", TestNeo4jBasic))
}

func TestNeo4jBasic(t *testing.T, env ...string) {
	_, boltPort := initNeo4jContainer()
	UseApp("neo4j")
	RunGoBuild(t, "go", "build", "test_neo4j_basic.go")
	env = append(env, "NEO4J_BOLT_PORT="+boltPort.Port())
	RunApp(t, "test_neo4j_basic", env...)
}

func initNeo4jContainer() (testcontainers.Container, nat.Port) {
	containerReqeust := testcontainers.ContainerRequest{
		Image:        "neo4j:5.26.0",
		ExposedPorts: []string{"7687/tcp"},
		Env: map[string]string{
			"NEO4J_AUTH": "neo4j/password",
		},
		WaitingFor: wait.ForLog("Started.").WithStartupTimeout(180 * time.Second)}
	neo4jC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: containerReqeust, Started: true})
	if err != nil {
		panic(err)
	}
	port, err := neo4jC.MappedPort(context.Background(), "7687")
	if err != nil {
		panic(err)
	}
	return neo4jC, port
}
