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
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	go setupWithException()
	time.Sleep(5 * time.Second)
	GetException()
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8888/exception", "http", "", "tcp", "ipv4", "", "127.0.0.1:8888", 500, 0, 8888)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /exception", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8888", "Host", "http", "/exception", "", "/exception", 500)
		fmt.Printf("%v %v\n", stubs[0][0].Status.Code, stubs[0][1].Status.Code)
		if stubs[0][0].Status.Code != codes.Error {
			panic("span should be error state")
		}
	}, 1)
}
