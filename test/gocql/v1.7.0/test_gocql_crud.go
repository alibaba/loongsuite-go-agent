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

package main

import (
	"log"
	"os"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/gocql/gocql"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

var session *gocql.Session

func TestCreateKeySpace() {
	if err := session.Query(`CREATE KEYSPACE IF NOT EXISTS cassandra WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : '1' };`).Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestCreateTable() {
	if err := session.Query(`CREATE TABLE IF NOT EXISTS cassandra.shopping_cart (userid text PRIMARY KEY,item_count int,last_update_timestamp timestamp);`).Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestInsert() {
	if err := session.Query("INSERT INTO cassandra.shopping_cart\n(userid, item_count, last_update_timestamp)\nVALUES ('9876', 2, toTimeStamp(now()));").Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestQuery() {
	var userid string
	if err := session.Query("SELECT userid FROM cassandra.shopping_cart;").Scan(&userid); err != nil {
		log.Printf("%v", err)
	}
}

func TestUpdate() {
	if err := session.Query("update cassandra.shopping_cart \nSET item_count = 6, last_update_timestamp = toTimeStamp(now()) \nWHERE userid = '1234';").Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestDelete() {
	if err := session.Query("delete FROM cassandra.shopping_cart \nWHERE userid = '1234';").Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestDropTable() {
	if err := session.Query("DROP table IF EXISTS cassandra.shopping_cart;").Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func TestDropKeyspace() {
	if err := session.Query("DROP KEYSPACE IF EXISTS cassandra;").Exec(); err != nil {
		log.Printf("%v", err)
	}
}

func main() {
	host := "127.0.0.1:" + os.Getenv("CASSANDRA_PORT")
	clusterCfg := gocql.NewCluster(host)
	clusterCfg.Authenticator = gocql.PasswordAuthenticator{Username: "cassandra"}
	s, err := clusterCfg.CreateSession()
	if err != nil {
		panic(err)
	}
	session = s
	defer session.Close()
	TestCreateKeySpace()
	TestCreateTable()
	TestInsert()
	TestQuery()
	TestUpdate()
	TestDelete()
	TestDropTable()
	TestDropKeyspace()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyDbAttributes(stubs[0][0], "CREATE KEYSPACE", "cassandra", host, "CREATE KEYSPACE IF NOT EXISTS cassandra WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : '1' };", "CREATE KEYSPACE", "", nil)
		verifier.VerifyDbAttributes(stubs[1][0], "CREATE TABLE", "cassandra", host, "CREATE TABLE IF NOT EXISTS cassandra.shopping_cart (userid text PRIMARY KEY,item_count int,last_update_timestamp timestamp);", "CREATE TABLE", "", nil)
		verifier.VerifyDbAttributes(stubs[2][0], "INSERT", "cassandra", host, "INSERT INTO cassandra.shopping_cart\n(userid, item_count, last_update_timestamp)\nVALUES ('9876', 2, toTimeStamp(now()));", "INSERT", "", nil)
		verifier.VerifyDbAttributes(stubs[3][0], "SELECT", "cassandra", host, "SELECT userid FROM cassandra.shopping_cart;", "SELECT", "", nil)
		verifier.VerifyDbAttributes(stubs[4][0], "UPDATE", "cassandra", host, "update cassandra.shopping_cart \nSET item_count = 6, last_update_timestamp = toTimeStamp(now()) \nWHERE userid = '1234';", "UPDATE", "", nil)
		verifier.VerifyDbAttributes(stubs[5][0], "DELETE", "cassandra", host, "delete FROM cassandra.shopping_cart \nWHERE userid = '1234';", "DELETE", "", nil)
		verifier.VerifyDbAttributes(stubs[6][0], "DROP TABLE", "cassandra", host, "DROP table IF EXISTS cassandra.shopping_cart;", "DROP TABLE", "", nil)
		verifier.VerifyDbAttributes(stubs[7][0], "DROP KEYSPACE", "cassandra", host, "DROP KEYSPACE IF EXISTS cassandra;", "DROP KEYSPACE", "", nil)
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
				if verifier.GetAttribute(attrs, "db.system.name").AsString() == "cassandra" &&
					verifier.GetAttribute(attrs, "db.operation.name").AsString() == "SELECT" &&
					verifier.GetAttribute(attrs, "server.address").AsString() == host {
					found = true
					break
				}
			}
			if !found {
				panic("db.client.request.duration missing SELECT datapoint for cassandra")
			}
		},
	})
}
