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

package test

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"net"
	"strconv"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const go_mysql_module_name = "gomysql"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("gomysql-conn-test", go_mysql_module_name, "1.11.0", "1.14.0", "1.18", "", TestGoMySQLconn),
		NewGeneralTestCase("gomysql-execute-commands-test", go_mysql_module_name, "1.11.0", "1.14.0", "1.18", "", TestGoMySQLexecuteCommands))
}

func TestGoMySQLconn(t *testing.T, env ...string) {
	_, gomysqlPort := initgoMySQLContainer()
	UseApp("gomysql")
	RunGoBuild(t, "go", "build", "test_conn.go")
	env = append(env, "GOMYSQL_ADDR=127.0.0.1:"+gomysqlPort.Port(), "GOMYSQL_USER=root", "GOMYSQL_PASSWORD=secret", "GOMYSQL_DBNAME=test")
	RunApp(t, "test_conn", env...)
}
func TestGoMySQLexecuteCommands(t *testing.T, env ...string) {
	_, gomysqlPort := initgoMySQLContainer()
	UseApp("gomysql")
	RunGoBuild(t, "go", "build", "test_executing_commands.go")
	env = append(env, "GOMYSQL_ADDR=127.0.0.1:"+gomysqlPort.Port(), "GOMYSQL_USER=root", "GOMYSQL_PASSWORD=secret", "GOMYSQL_DBNAME=test")
	RunApp(t, "test_executing_commands", env...)
}

func initgoMySQLContainer() (testcontainers.Container, nat.Port) {
	randport, err := getRandomPort()
	if err != nil {
		panic(err)
	}
	req := testcontainers.ContainerRequest{
		Image:        "mysql:latest",
		ExposedPorts: []string{"3306/tcp"},
		Env: map[string]string{
			"MYSQL_ROOT_PASSWORD": "secret",
			"MYSQL_DATABASE":      "test",
		},
		WaitingFor: wait.ForLog("ready for connections"),
		HostConfigModifier: func(hostConfig *container.HostConfig) {
			hostConfig.PortBindings = nat.PortMap{
				"3306/tcp": []nat.PortBinding{{HostIP: "127.0.0.1", HostPort: strconv.Itoa(randport)}},
			}
			hostConfig.NetworkMode = "bridge"
		},
	}
	gomysqlC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		panic(err)
	}
	port, err := gomysqlC.MappedPort(context.Background(), "3306")
	if err != nil {
		panic(err)
	}
	return gomysqlC, port
}
func getRandomPort() (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
