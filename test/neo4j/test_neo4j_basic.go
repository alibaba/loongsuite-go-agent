// Copyright (c) 2025 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"os"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	port := os.Getenv("NEO4J_BOLT_PORT")
	if port == "" {
		port = "7687"
	}
	ctx := context.Background()

	driver, err := neo4j.NewDriver("bolt://127.0.0.1:"+port, neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		panic(err)
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err = session.Run(ctx, "MATCH (n) RETURN n LIMIT $limit", map[string]any{"limit": 10})
	if err != nil {
		panic(err)
	}

	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "CREATE (n:Person {name: $name})", map[string]any{"name": "loongsuite"})
		return nil, err
	})
	if err != nil {
		panic(err)
	}

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		var matchSpan, createSpan tracetest.SpanStub
		var foundMatch, foundCreate bool

		for _, trace := range stubs {
			for _, span := range trace {
				if verifier.GetAttribute(span.Attributes, "db.system.name").AsString() != "neo4j" {
					continue
				}
				operation := verifier.GetAttribute(span.Attributes, "db.operation.name").AsString()
				switch operation {
				case "MATCH":
					matchSpan = span
					foundMatch = true
				case "CREATE":
					createSpan = span
					foundCreate = true
				}
			}
		}

		if !foundMatch {
			panic("Expected to find span with operation MATCH")
		}
		if !foundCreate {
			panic("Expected to find span with operation CREATE")
		}

		verifier.VerifyDbAttributes(matchSpan, "MATCH", "neo4j", "", "MATCH (n) RETURN n LIMIT $limit", "MATCH", "", nil)
		verifier.VerifyDbAttributes(createSpan, "CREATE", "neo4j", "", "CREATE (n:Person {name: $name})", "CREATE", "", nil)
	}, 2)

	verifier.WaitAndAssertMetrics(map[string]func(metricdata.ResourceMetrics){
		"db.client.request.duration": func(mrs metricdata.ResourceMetrics) {
			if len(mrs.ScopeMetrics) <= 0 {
				panic("No db.client.request.duration metrics received!")
			}
			point := mrs.ScopeMetrics[0].Metrics[0].Data.(metricdata.Histogram[float64])
			if len(point.DataPoints) <= 0 {
				panic("db.client.request.duration has no datapoints")
			}
			var foundMatch, foundCreate bool
			for _, dp := range point.DataPoints {
				if dp.Count <= 0 {
					continue
				}
				attrs := dp.Attributes.ToSlice()
				if verifier.GetAttribute(attrs, "db.system.name").AsString() != "neo4j" {
					continue
				}
				switch verifier.GetAttribute(attrs, "db.operation.name").AsString() {
				case "MATCH":
					foundMatch = true
				case "CREATE":
					foundCreate = true
				}
			}
			if !foundMatch {
				panic("db.client.request.duration missing MATCH datapoint for neo4j")
			}
			if !foundCreate {
				panic("db.client.request.duration missing CREATE datapoint for neo4j")
			}
		},
	})
}
