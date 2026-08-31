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

package main

import (
	"fmt"
	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/go-mysql-org/go-mysql/client"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"os"
)

func main() {
	dbPrepared()
}

func dbPrepared() {
	// This test relies on an external MySQL instance.
	// Configure via env vars to match your CI/local environment.
	addr := getenvDefault("GOMYSQL_ADDR", "127.0.0.1:3306")
	user := getenvDefault("GOMYSQL_USER", "root")
	pass := os.Getenv("GOMYSQL_PASSWORD")
	dbName := getenvDefault("GOMYSQL_DBNAME", "mysql")
	conn, err := client.Connect(addr, user, pass, dbName)
	verifier.Assert(err == nil, "failed to connect to mysql at %s: %v", addr, err)

	// Attempt to create a table name x if not exists and then drop it. This is a simple command that should be supported by all MySQL versions and should be captured in a span.
	_, err = conn.Execute("CREATE TABLE IF NOT EXISTS testtablex (id INT)")
	verifier.Assert(err == nil, "failed to execute create table command: %v", err)
	_, err = conn.Execute("DROP TABLE testtablex")
	verifier.Assert(err == nil, "failed to execute drop table command: %v", err)
	_ = conn.Close()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		fmt.Println("Received spans:", stubs[1])
		fmt.Println("Received spans:", stubs[2])
		verifier.Assert(len(stubs) == 3, "expected 3 spans (dial, create table, drop table), got %d", len(stubs))
		verifier.VerifyDbAttributes(stubs[1][0], "CREATE", "mysql", addr, "CREATE TABLE IF NOT EXISTS testtablex (id INT)", "CREATE", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "DROP", "mysql", addr, "DROP TABLE testtablex", "DROP", "", nil)
	}, 2)
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
