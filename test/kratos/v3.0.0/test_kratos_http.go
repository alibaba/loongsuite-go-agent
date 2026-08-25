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
	"fmt"
	"github.com/alibaba/loongsuite-go/test/verifier"
	transhttp "github.com/go-kratos/kratos/v3/transport/http"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	pb "kratos/v3.0.0/pkg/api/helloworld/v1"
	"strings"
	"time"
)

// findKratosSpan returns the first span across all traces whose kratos.span.kind
// attribute matches the given kind. Looking up by attribute (rather than a fixed
// index) keeps the assertions stable regardless of how many transport spans the
// trace contains.
func findKratosSpan(stubs []tracetest.SpanStubs, spanKind string) *tracetest.SpanStub {
	for _, trace := range stubs {
		for i := range trace {
			if verifier.GetAttribute(trace[i].Attributes, "kratos.span.kind").AsString() == spanKind {
				return &trace[i]
			}
		}
	}
	return nil
}

func main() {
	go func() {
		startup()
	}()
	time.Sleep(5 * time.Second)
	conn, err := transhttp.NewClient(
		context.Background(),
		transhttp.WithEndpoint("localhost:8000"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewGreeterHTTPClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "client"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("[http] SayHello %+v\n", reply)

	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		server := findKratosSpan(stubs, "server")
		if server == nil {
			panic("kratos server span not found")
		}
		protocolType := verifier.GetAttribute(server.Attributes, "kratos.protocol.type").AsString()
		if protocolType != "http" {
			panic("protocol type should be http, actually got " + protocolType)
		}
		serviceName := verifier.GetAttribute(server.Attributes, "kratos.service.name").AsString()
		if serviceName != "opentelemetry-kratos-server" {
			panic("service name should be opentelemetry-kratos-server, actually got " + serviceName)
		}
		serviceId := verifier.GetAttribute(server.Attributes, "kratos.service.id").AsString()
		if serviceId != "opentelemetry-id" {
			panic("service id should be opentelemetry-id, actually got " + serviceId)
		}
		serviceVersion := verifier.GetAttribute(server.Attributes, "kratos.service.version").AsString()
		if serviceVersion != "v1" {
			panic("service version should be v1, actually got " + serviceVersion)
		}
		serviceMetaAgent := verifier.GetAttribute(server.Attributes, "kratos.service.meta.agent").AsString()
		if serviceMetaAgent != "opentelemetry-go" {
			panic("service meta agent should be opentelemetry-go, actually got " + serviceMetaAgent)
		}
		serviceEndpoint := verifier.GetAttribute(server.Attributes, "kratos.service.endpoint").AsStringSlice()
		if !strings.Contains(serviceEndpoint[0], ":9000") || !strings.Contains(serviceEndpoint[1], ":8000") {
			panic("service endpoint should be grpc://30.221.144.142:9000 http://30.221.144.142:8000, actually got " + fmt.Sprintf("%v", serviceEndpoint))
		}

		client := findKratosSpan(stubs, "client")
		if client == nil {
			panic("kratos client span not found")
		}
		clientProtocol := verifier.GetAttribute(client.Attributes, "kratos.protocol.type").AsString()
		if clientProtocol != "http" {
			panic("client protocol type should be http, actually got " + clientProtocol)
		}
		clientOperation := verifier.GetAttribute(client.Attributes, "kratos.operation").AsString()
		if !strings.Contains(clientOperation, "SayHello") {
			panic("client operation should contain SayHello, actually got " + clientOperation)
		}
	}, 1)
}
