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

package neo4j

import (
	"context"
	"os"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
)

type neo4jInnerEnabler struct {
	enabled bool
}

func (n neo4jInnerEnabler) Enable() bool {
	return n.enabled
}

var neo4jEnabler = neo4jInnerEnabler{
	enabled: os.Getenv("OTEL_INSTRUMENTATION_NEO4J_ENABLED") != "false",
}

var neo4jInstrumenter = BuildNeo4jInstrumenter()

type neo4jSpanData struct {
	request neo4jRequest
	ctx     context.Context
}

//go:linkname sessionRunOnEnter github.com/neo4j/neo4j-go-driver/v6/neo4j.sessionRunOnEnter
func sessionRunOnEnter(call api.CallContext, s interface{}, ctx context.Context, cypher string, params map[string]any, configurers ...func(*neo4j.TransactionConfig)) {
	if !neo4jEnabler.Enable() {
		return
	}
	request := neo4jRequest{
		Statement: cypher,
		Op:        extractOpType(cypher),
		BatchSize: 1,
	}
	call.SetData(neo4jSpanData{
		request: request,
		ctx:     neo4jInstrumenter.Start(ctx, request),
	})
}

//go:linkname sessionRunOnExit github.com/neo4j/neo4j-go-driver/v6/neo4j.sessionRunOnExit
func sessionRunOnExit(call api.CallContext, result neo4j.Result, err error) {
	if !neo4jEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(neo4jSpanData)
	if !ok {
		return
	}
	neo4jInstrumenter.End(data.ctx, data.request, result, err)
}

//go:linkname transactionRunOnEnter github.com/neo4j/neo4j-go-driver/v6/neo4j.transactionRunOnEnter
func transactionRunOnEnter(call api.CallContext, tx interface{}, ctx context.Context, cypher string, params map[string]any) {
	if !neo4jEnabler.Enable() {
		return
	}
	request := neo4jRequest{
		Statement: cypher,
		Op:        extractOpType(cypher),
		BatchSize: 1,
	}
	call.SetData(neo4jSpanData{
		request: request,
		ctx:     neo4jInstrumenter.Start(ctx, request),
	})
}

//go:linkname transactionRunOnExit github.com/neo4j/neo4j-go-driver/v6/neo4j.transactionRunOnExit
func transactionRunOnExit(call api.CallContext, result neo4j.Result, err error) {
	if !neo4jEnabler.Enable() {
		return
	}
	data, ok := call.GetData().(neo4jSpanData)
	if !ok {
		return
	}
	neo4jInstrumenter.End(data.ctx, data.request, result, err)
}
