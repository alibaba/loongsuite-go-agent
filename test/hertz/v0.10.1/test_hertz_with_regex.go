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
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	go setupWithRoute()
	time.Sleep(5 * time.Second)
	GetRoute()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8888/hertz/v1", "http", "", "tcp", "ipv4", "", "127.0.0.1:8888", 301, 0, 8888)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8888", "Host", "http", "/hertz/v1/", "", verifier.OmitHttpRoute, 301)
		verifier.VerifyHttpClientAttributes(stubs[1][0], "GET", "GET", "http://127.0.0.1:8888/hertz/v1/", "http", "", "tcp", "ipv4", "", "127.0.0.1:8888", 200, 0, 8888)
		verifier.VerifyHttpServerAttributes(stubs[1][1], "GET /hertz/:version/*action", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8888", "Host", "http", "/hertz/v1/", "", "/hertz/:version/*action", 200)
		verifier.VerifyHttpClientAttributes(stubs[2][0], "GET", "GET", "http://127.0.0.1:8888/hertz/v2/send", "http", "", "tcp", "ipv4", "", "127.0.0.1:8888", 200, 0, 8888)
		verifier.VerifyHttpServerAttributes(stubs[2][1], "GET /hertz/:version/*action", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8888", "Host", "http", "/hertz/v2/send", "", "/hertz/:version/*action", 200)
	}, 3)
}
