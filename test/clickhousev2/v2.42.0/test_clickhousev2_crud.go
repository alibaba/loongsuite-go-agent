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
	"context"
	"log"
	"os"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var con driver.Conn

type User struct {
	ID   string `ch:"id"`
	Name string `ch:"name"`
	Age  *int32 `ch:"age"`
}

var (
	createTable = `CREATE TABLE IF NOT EXISTS users (
		id String, 
		name String,
		age Int32
	) ENGINE = MergeTree() ORDER BY id`

	insert = `INSERT INTO users (id, name, age) VALUES (?, ?, ?)`
)

func TestExec() {
	if err := con.Exec(context.Background(), createTable); err != nil {
		log.Fatalf("TestExec fail: %v\n", err)
	}
}

func TestAsyncInsert() {
	// AsyncInsert 需要参数化查询
	if err := con.AsyncInsert(context.Background(), insert, false, "1", "1", uint32(1)); err != nil {
		log.Fatalf("TestAsyncInsert fail: %v\n", err)
	}
}

func TestSelect() {
	var users []User
	// 使用字符串参数，因为id列是String类型
	if err := con.Select(context.Background(), &users, `SELECT * FROM users WHERE id = ?`, "1"); err != nil {
		log.Fatalf("TestSelect fail: %v\n", err)
	}
	log.Printf("Selected users: %+v\n", users)
}

func TestQuery() {
	rows, err := con.Query(context.Background(), `SELECT * FROM users WHERE id = ?`, "1")
	if err != nil {
		log.Fatalf("TestQuery fail: %v\n", err)
	}
	defer rows.Close()
}

func TestQueryRow() {
	row := con.QueryRow(context.Background(), `SELECT * FROM users WHERE id = ?`, "1")
	if row.Err() != nil {
		log.Fatalf("TestQueryRow fail: %v\n", row.Err())
	}
}

func TestPrepareBatch() {
	// 修正batch查询 - 使用INSERT而不是SELECT
	batch, err := con.PrepareBatch(context.Background(), `INSERT INTO users (id, name, age)`)
	if err != nil {
		log.Fatalf("TestPrepareBatch PrepareBatch fail: %v\n", err)
	}

	// 添加行数据
	err = batch.Append("2", "User2", int32(25))
	if err != nil {
		log.Fatalf("TestPrepareBatch Append fail: %v\n", err)
	}

	if err = batch.Send(); err != nil {
		log.Fatalf("TestPrepareBatch Send fail: %v\n", err)
	}
	log.Println("Batch insert completed")
}

func TestSelectVersion() {
	version, err := con.ServerVersion()
	if err != nil {
		log.Fatalf("TestSelectVersion fail: %v\n", err)
	}
	log.Printf("Server version: %s\n", version)
}

func TestPing() {
	if err := con.Ping(context.Background()); err != nil {
		log.Fatalf("TestPing fail: %v\n", err)
	}
	log.Println("Ping successful")
}

func main() {
	addr := "127.0.0.1:" + os.Getenv("CLICKHOUSE_PORT")
	tmpCon, err := clickhouse.Open(&clickhouse.Options{
		Addr:     []string{addr},
		Protocol: clickhouse.Native,
	})
	if err != nil {
		log.Fatalf("open connection fail, err: %v", err)
	}
	con = tmpCon
	TestExec()
	TestAsyncInsert()
	TestSelect()
	TestQuery()
	TestQueryRow()
	TestPrepareBatch()
	TestSelectVersion()
	TestPing()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "EXEC", "clickhouse", addr, createTable, "EXEC", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "ASYNC_INSERT", "clickhouse", addr, "INSERT INTO users (id, name, age) VALUES (?, ?, ?)", "ASYNC_INSERT", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "SELECT", "clickhouse", addr, "SELECT * FROM users WHERE id = ?", "SELECT", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "QUERY", "clickhouse", addr, "SELECT * FROM users WHERE id = ?", "QUERY", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "QUERY_ROW", "clickhouse", addr, "SELECT * FROM users WHERE id = ?", "QUERY_ROW", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "PREPARE_BATCH", "clickhouse", addr, "INSERT INTO users (id, name, age)", "PREPARE_BATCH", "", nil)
		verifier.VerifyDbAttributes(stubs[6][0], "SERVER_VERSION", "clickhouse", addr, "SERVER_VERSION", "SERVER_VERSION", "", nil)
		verifier.VerifyDbAttributes(stubs[7][0], "PING", "clickhouse", addr, "PING", "PING", "", nil)
	}, 1)
	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"db.client.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No db.client.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if len(point.DataPoints) <= 0 {
				panic("db.client.request.duration has no datapoints")
			}
			found := false
			for _, dp := range point.DataPoints {
				if dp.Count <= 0 {
					continue
				}
				attrs := dp.Attributes.ToSlice()
				if verifier.GetAttribute(attrs, "db.system.name").AsString() == "clickhouse" &&
					verifier.GetAttribute(attrs, "db.operation.name").AsString() == "EXEC" &&
					verifier.GetAttribute(attrs, "server.address").AsString() == addr {
					found = true
					break
				}
			}
			if !found {
				panic("db.client.request.duration missing EXEC datapoint for clickhouse")
			}
		},
	})
}
