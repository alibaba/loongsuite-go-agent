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
	"os"
	"strings"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/go-mysql-org/go-mysql/client"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
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
	_ = conn.Close()

	// One connect should emit a dial span.
	verifier.WaitAndAssertTraces(func(traces []tracetest.SpanStubs) {
		// Find a span that looks like a MySQL connect/dial span.
		// We keep the name check flexible to avoid coupling to naming changes.
		var dialSpan *tracetest.SpanStub
		for _, trace := range traces {
			for i := range trace {
				sp := trace[i]
				system := verifier.GetAttribute(sp.Attributes, "db.system.name").AsString()
				serverAddr := verifier.GetAttribute(sp.Attributes, "server.address").AsString()
				connection_id := verifier.GetAttribute(sp.Attributes, "db.connectionId").AsString()
				verifier.Assert(connection_id != "", "expected db.connection_id attribute to be present in mysql dial span")
				if system == "mysql" && strings.Contains(serverAddr, strings.Split(addr, ":")[0]) {
					dialSpan = &sp
					break
				}
			}
			if dialSpan != nil {
				break
			}
		}
		verifier.Assert(dialSpan != nil, "did not find mysql dial span in exported traces")

		// Assert core DB attributes for connection span.
		// statement/operation/collection are empty for connect.
		verifier.VerifyDbAttributes(*dialSpan, dialSpan.Name, "mysql", addr, "", "", "", nil)
	}, 1)
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
