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

package verifier

import (
	"strings"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// VerifyDbAttributes TODO: make attribute name to semconv attribute
func VerifyDbAttributes(span tracetest.SpanStub, name, system, address, statement, operation, collection string, params []any) {
	Assert(span.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", span.SpanKind)
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	actualSystem := GetAttribute(span.Attributes, "db.system.name").AsString()
	Assert(actualSystem == system, "Except client db system to be %s, got %s", system, actualSystem)
	actualConnStr := GetAttribute(span.Attributes, "server.address").AsString()
	Assert(strings.Contains(actualConnStr, address), "Except client address str to be %s, got %s", address, actualConnStr)
	actualStatement := GetAttribute(span.Attributes, "db.query.text").AsString()
	Assert(strings.Contains(actualStatement, statement), "Except client db statement to be %s, got %s", statement, actualStatement)
	actualOperation := GetAttribute(span.Attributes, "db.operation.name").AsString()
	Assert(actualOperation == operation, "Except client db operation to be %s, got %s", operation, actualOperation)
	actualCollection := GetAttribute(span.Attributes, "db.collection.name").AsString()
	Assert(actualCollection == collection, "Expect client collection to be %s, got %s", collection, actualCollection)
	actualParams := GetAttributesWithPrefix(span.Attributes, "db.query.parameter")
	SliceAttrsAssert(params, actualParams, "Expect client db params to be %#v, got %#v", params, actualParams)
}

func VerifyHttpClientAttributes(span tracetest.SpanStub, name, method, fullUrl, protocolName, protocolVersion, networkTransport, networkType, localAddr, peerAddr string, statusCode, localPort, peerPort int64) {
	Assert(span.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", span.SpanKind)
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	Assert(GetAttribute(span.Attributes, "http.request.method").AsString() == method, "Except method to be %s, got %s", method, GetAttribute(span.Attributes, "http.request.method").AsString())
	Assert(GetAttribute(span.Attributes, "url.full").AsString() == fullUrl, "Except full url to be %s, got %s", fullUrl, GetAttribute(span.Attributes, "url.full").AsString())
	Assert(GetAttribute(span.Attributes, "network.protocol.name").AsString() == protocolName, "Except protocol name to be %s, got %s", protocolName, GetAttribute(span.Attributes, "network.protocol.name").AsString())
	Assert(GetAttribute(span.Attributes, "network.protocol.version").AsString() == protocolVersion, "Except protocol version to be %s, got %s", protocolVersion, GetAttribute(span.Attributes, "network.protocol.version").AsString())
	Assert(GetAttribute(span.Attributes, "network.transport").AsString() == networkTransport, "Except network transport to be %s, got %s", networkTransport, GetAttribute(span.Attributes, "network.transport").AsString())
	Assert(GetAttribute(span.Attributes, "network.type").AsString() == networkType, "Except network type to be %s, got %s", networkType, GetAttribute(span.Attributes, "network.type").AsString())
	Assert(GetAttribute(span.Attributes, "network.local.address").AsString() == localAddr, "Except local addr to be %s, got %s", localAddr, GetAttribute(span.Attributes, "network.local.address").AsString())
	Assert(GetAttribute(span.Attributes, "network.peer.address").AsString() == peerAddr, "Except peer addr to be %s, got %s", peerAddr, GetAttribute(span.Attributes, "network.peer.address").AsString())
	Assert(GetAttribute(span.Attributes, "http.response.status_code").AsInt64() == statusCode, "Except status code to be %d, got %d", statusCode, GetAttribute(span.Attributes, "http.response.status_code").AsInt64())
	Assert(GetAttribute(span.Attributes, "network.peer.port").AsInt64() == peerPort, "Except peer port to be %d, got %d", peerPort, GetAttribute(span.Attributes, "network.peer.port").AsInt64())
	if localPort > 0 {
		Assert(GetAttribute(span.Attributes, "network.local.port").AsInt64() == localPort, "Except local port to be %d, got %d", localPort, GetAttribute(span.Attributes, "network.local.port").AsInt64())
	}
}

func VerifyHttpClientMetricsAttributes(attrs []attribute.KeyValue, method, serverAddress, errorType, protocolName, protocolVersion string, serverPort, statusCode int) {
	Assert(GetAttribute(attrs, "http.request.method").AsString() == method, "Except method to be %s, got %s", method, GetAttribute(attrs, "http.request.method").AsString())
	Assert(GetAttribute(attrs, "server.address").AsString() == serverAddress, "Except server.address to be %s, got %s", serverAddress, GetAttribute(attrs, "server.address").AsString())
	Assert(GetAttribute(attrs, "error.type").AsString() == errorType, "Except error.type to be %s, got %s", errorType, GetAttribute(attrs, "error.type").AsString())
	Assert(GetAttribute(attrs, "network.protocol.name").AsString() == protocolName, "Except network.protocol.name to be %s, got %s", protocolName, GetAttribute(attrs, "network.protocol.name").AsString())
	Assert(GetAttribute(attrs, "network.protocol.version").AsString() == protocolVersion, "Except network.protocol.version to be %s, got %s", protocolVersion, GetAttribute(attrs, "network.protocol.version").AsString())
	Assert(GetAttribute(attrs, "server.port").AsInt64() == int64(serverPort), "Except server.port to be %d, got %d", serverPort, GetAttribute(attrs, "server.port").AsInt64())
	Assert(GetAttribute(attrs, "http.response.status_code").AsInt64() == int64(statusCode), "Except status code to be %d, got %d", statusCode, GetAttribute(attrs, "http.response.status_code").AsInt64())
}

func VerifyHttpServerAttributes(span tracetest.SpanStub, name, method, protocolName, networkTransport, networkType, localAddr, peerAddr, agent, scheme, path, query, route string, statusCode int64) {
	Assert(span.SpanKind == trace.SpanKindServer, "Expect to be client span, got %d", span.SpanKind)
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	Assert(GetAttribute(span.Attributes, "http.request.method").AsString() == method, "Except method to be %s, got %s", method, GetAttribute(span.Attributes, "http.request.method").AsString())
	Assert(GetAttribute(span.Attributes, "network.protocol.name").AsString() == protocolName, "Except protocol name to be %s, got %s", protocolName, GetAttribute(span.Attributes, "network.protocol.name").AsString())
	Assert(GetAttribute(span.Attributes, "network.transport").AsString() == networkTransport, "Except network transport to be %s, got %s", networkTransport, GetAttribute(span.Attributes, "network.transport").AsString())
	Assert(GetAttribute(span.Attributes, "network.type").AsString() == networkType, "Except network type to be %s, got %s", networkType, GetAttribute(span.Attributes, "network.type").AsString())
	Assert(GetAttribute(span.Attributes, "network.local.address").AsString() == localAddr, "Except local addr to be %s, got %s", localAddr, GetAttribute(span.Attributes, "network.local.address").AsString())
	Assert(GetAttribute(span.Attributes, "network.peer.address").AsString() == peerAddr, "Except peer addr to be %s, got %s", peerAddr, GetAttribute(span.Attributes, "network.peer.address").AsString())
	Assert(GetAttribute(span.Attributes, "user_agent.original").AsString() == agent, "Except user agent to be %s, got %s", agent, GetAttribute(span.Attributes, "user_agent.original").AsString())
	Assert(GetAttribute(span.Attributes, "url.scheme").AsString() == scheme, "Except url scheme to be %s, got %s", scheme, GetAttribute(span.Attributes, "url.scheme").AsString())
	Assert(GetAttribute(span.Attributes, "url.path").AsString() == path, "Except url path to be %s, got %s", path, GetAttribute(span.Attributes, "url.path").AsString())
	Assert(GetAttribute(span.Attributes, "url.query").AsString() == query, "Except url query to be %s, got %s", query, GetAttribute(span.Attributes, "url.query").AsString())
	Assert(GetAttribute(span.Attributes, "http.route").AsString() == route, "Except http route to be %s, got %s", route, GetAttribute(span.Attributes, "http.route").AsString())
	Assert(GetAttribute(span.Attributes, "http.response.status_code").AsInt64() == statusCode, "Except status code to be %d, got %d", statusCode, GetAttribute(span.Attributes, "http.response.status_code").AsInt64())
}

func VerifyHttpServerMetricsAttributes(attrs []attribute.KeyValue, method, httpRoute, errorType, protocolName, protocolVersion, urlScheme string, statusCode int) {
	Assert(GetAttribute(attrs, "http.request.method").AsString() == method, "Except method to be %s, got %s", method, GetAttribute(attrs, "http.request.method").AsString())
	Assert(GetAttribute(attrs, "http.route").AsString() == httpRoute, "Except http.route to be %s, got %s", httpRoute, GetAttribute(attrs, "http.route").AsString())
	Assert(GetAttribute(attrs, "error.type").AsString() == errorType, "Except error.type to be %s, got %s", errorType, GetAttribute(attrs, "error.type").AsString())
	Assert(GetAttribute(attrs, "network.protocol.name").AsString() == protocolName, "Except network.protocol.name to be %s, got %s", protocolName, GetAttribute(attrs, "network.protocol.name").AsString())
	Assert(GetAttribute(attrs, "network.protocol.version").AsString() == protocolVersion, "Except network.protocol.version to be %s, got %s", protocolVersion, GetAttribute(attrs, "network.protocol.version").AsString())
	Assert(GetAttribute(attrs, "url.scheme").AsString() == urlScheme, "Except url.scheme to be %s, got %s", urlScheme, GetAttribute(attrs, "url.scheme").AsString())
	Assert(GetAttribute(attrs, "http.response.status_code").AsInt64() == int64(statusCode), "Except status code to be %d, got %d", statusCode, GetAttribute(attrs, "http.response.status_code").AsInt64())
}

func VerifyRpcServerAttributes(span tracetest.SpanStub, name, system, service, method string) {
	Assert(span.SpanKind == trace.SpanKindServer, "Expect to be server span, got %d", span.SpanKind)
	verifyRpcAttributes(span, name, system, service, method)
}

func VerifyRpcClientAttributes(span tracetest.SpanStub, name, system, service, method string) {
	Assert(span.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", span.SpanKind)
	verifyRpcAttributes(span, name, system, service, method)
}

func verifyRpcAttributes(span tracetest.SpanStub, name, system, service, method string) {
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	Assert(GetAttribute(span.Attributes, "rpc.system").AsString() == system, "Except rpc system to be %s, got %s", method, GetAttribute(span.Attributes, "rpc.system").AsString())
	Assert(GetAttribute(span.Attributes, "rpc.service").AsString() == service, "Except rpc service to be %s, got %s", method, GetAttribute(span.Attributes, "rpc.service").AsString())
	Assert(GetAttribute(span.Attributes, "rpc.method").AsString() == method, "Except rpc method to be %s, got %s", method, GetAttribute(span.Attributes, "rpc.method").AsString())
}

func VerifyDbMetricsAttributes(attrs []attribute.KeyValue, dbSystem, operationName, serverAddress string) {
	Assert(GetAttribute(attrs, string(semconv.DBSystemNameKey)).AsString() == dbSystem, "Expected db.system.name to be %s, got %s", dbSystem, GetAttribute(attrs, string(semconv.DBSystemNameKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.DBOperationNameKey)).AsString() == operationName, "Expected db.operation.name to be %s, got %s", operationName, GetAttribute(attrs, string(semconv.DBOperationNameKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.ServerAddressKey)).AsString() == serverAddress, "Expected server.address to be %s, got %s", serverAddress, GetAttribute(attrs, string(semconv.ServerAddressKey)).AsString())
}

func VerifyGenAIOperationDurationMetricsAttributes(attrs []attribute.KeyValue, operationName, system, requestModel, responseModel string) {
	Assert(GetAttribute(attrs, string(semconv.GenAIOperationNameKey)).AsString() == operationName, "Expected gen_ai.operation.name to be %s, got %s", operationName, GetAttribute(attrs, string(semconv.GenAIOperationNameKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAISystemKey)).AsString() == system, "Expected gen_ai.system to be %s, got %s", system, GetAttribute(attrs, string(semconv.GenAISystemKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAIRequestModelKey)).AsString() == requestModel, "Expected gen_ai.request.model to be %s, got %s", requestModel, GetAttribute(attrs, string(semconv.GenAIRequestModelKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAIResponseModelKey)).AsString() == responseModel, "Expected gen_ai.response.model to be %s, got %s", responseModel, GetAttribute(attrs, string(semconv.GenAIResponseModelKey)).AsString())
}

func VerifyGenAITokenUsageMetricsAttributes(attrs []attribute.KeyValue, operationName, system, requestModel, responseModel, tokenType string) {
	Assert(GetAttribute(attrs, string(semconv.GenAIOperationNameKey)).AsString() == operationName, "Expected gen_ai.operation.name to be %s, got %s", operationName, GetAttribute(attrs, string(semconv.GenAIOperationNameKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAISystemKey)).AsString() == system, "Expected gen_ai.system to be %s, got %s", system, GetAttribute(attrs, string(semconv.GenAISystemKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAIRequestModelKey)).AsString() == requestModel, "Expected gen_ai.request.model to be %s, got %s", requestModel, GetAttribute(attrs, string(semconv.GenAIRequestModelKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAIResponseModelKey)).AsString() == responseModel, "Expected gen_ai.response.model to be %s, got %s", responseModel, GetAttribute(attrs, string(semconv.GenAIResponseModelKey)).AsString())
	Assert(GetAttribute(attrs, string(semconv.GenAITokenTypeKey)).AsString() == tokenType, "Expected gen_ai.token.type to be %s, got %s", tokenType, GetAttribute(attrs, string(semconv.GenAITokenTypeKey)).AsString())
}

func VerifyRpcClientMetricsAttributes(attrs []attribute.KeyValue, method, service, system, serverAddr string) {
	Assert(GetAttribute(attrs, "rpc.method").AsString() == method, "Except rpc.method to be %s, got %s", method, GetAttribute(attrs, "rpc.method").AsString())
	Assert(GetAttribute(attrs, "rpc.service").AsString() == service, "Except rpc.service to be %s, got %s", service, GetAttribute(attrs, "rpc.service").AsString())
	Assert(GetAttribute(attrs, "rpc.system").AsString() == system, "Except rpc.system to be %s, got %s", system, GetAttribute(attrs, "rpc.system").AsString())
	Assert(GetAttribute(attrs, "server.address").AsString() == serverAddr, "Except rpc.system to be %s, got %s", serverAddr, GetAttribute(attrs, "server.address").AsString())
}

func VerifyRpcServerMetricsAttributes(attrs []attribute.KeyValue, method, service, system, serverAddr string) {
	Assert(GetAttribute(attrs, "rpc.method").AsString() == method, "Except rpc.method to be %s, got %s", method, GetAttribute(attrs, "rpc.method").AsString())
	Assert(GetAttribute(attrs, "rpc.service").AsString() == service, "Except rpc.service to be %s, got %s", service, GetAttribute(attrs, "rpc.service").AsString())
	Assert(GetAttribute(attrs, "rpc.system").AsString() == system, "Except rpc.system to be %s, got %s", system, GetAttribute(attrs, "rpc.system").AsString())
	Assert(GetAttribute(attrs, "server.address").AsString() == serverAddr, "Except rpc.system to be %s, got %s", serverAddr, GetAttribute(attrs, "server.address").AsString())
}

func VerifyLLMAttributes(span tracetest.SpanStub, name string, system string, model string) {
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	actualSystem := GetAttribute(span.Attributes, "gen_ai.system").AsString()
	Assert(actualSystem == system, "Except gen_ai.system to be %s, got %s", system, actualSystem)
	optName := GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
	Assert(optName == name, "Except gen_ai.operation.name to be %s, got %s", name, optName)
	optModel := GetAttribute(span.Attributes, "gen_ai.request.model").AsString()
	Assert(optModel == model, "Except gen_ai.request.model to be %s, got %s", model, optModel)
	Assert(span.SpanKind == trace.SpanKindClient, "Expect to be client span, got %d", span.SpanKind)
}
func VerifyLLMCommonAttributes(span tracetest.SpanStub, name string, system string, spanKind trace.SpanKind) {
	Assert(span.Name == name, "Except client span name to be %s, got %s", name, span.Name)
	actualSystem := GetAttribute(span.Attributes, "gen_ai.system").AsString()
	Assert(actualSystem == system, "Except gen_ai.system to be %s, got %s", system, actualSystem)
	optName := GetAttribute(span.Attributes, "gen_ai.operation.name").AsString()
	Assert(optName == name, "Except gen_ai.operation.name to be %s, got %s", name, optName)
	Assert(span.SpanKind == spanKind, "Expect to be %s span, got %d", spanKind, span.SpanKind)
}
func VerifyMQPublishAttributes(span tracetest.SpanStub, exchange, routing, queue, operationName, destination string, system string) {
	Assert(span.Name == destination+" "+operationName, "Except client span name to be %s, got %s", destination+" "+string(operationName), span.Name)
	actualDestination := GetAttribute(span.Attributes, "messaging.destination.name").AsString()
	Assert(actualDestination == destination, "Except messaging.destination.name to be %s, got %s", destination, actualDestination)
	optName := GetAttribute(span.Attributes, "messaging.operation.name").AsString()
	Assert(optName == operationName, "Except messaging.operation.name to be %s, got %s", operationName, optName)
	if routing != "" {
		routingKey := GetAttribute(span.Attributes, "messaging.rabbitmq.destination.routing_key").AsString()
		Assert(routingKey == routing, "Except messaging.rabbitmq.destination.routing_key to be %s, got %s", routing, routingKey)
	}
	actualSystem := GetAttribute(span.Attributes, "messaging.system").AsString()
	Assert(actualSystem == system, "Except messaging.system to be %s, got %s", system, actualSystem)
	Assert(span.SpanKind == trace.SpanKindProducer, "Expect to be producer span, got %d", span.SpanKind)
}
func VerifyMQConsumeAttributes(span tracetest.SpanStub, exchange, routing, queue, operationName, destination string, system string) {
	Assert(span.Name == destination+" "+operationName, "Except client span name to be %s, got %s", destination+" "+string(operationName), span.Name)
	actualDestination := GetAttribute(span.Attributes, "messaging.destination.name").AsString()
	Assert(actualDestination == destination, "Except messaging.destination.name to be %s, got %s", destination, actualDestination)
	optName := GetAttribute(span.Attributes, "messaging.operation.name").AsString()
	Assert(optName == operationName, "Except messaging.operation.name to be %s, got %s", operationName, optName)
	if routing != "" {
		routingKey := GetAttribute(span.Attributes, "messaging.rabbitmq.destination.routing_key").AsString()
		Assert(routingKey == routing, "Except messaging.rabbitmq.destination.routing_key to be %s, got %s", routing, routingKey)
	}
	actualSystem := GetAttribute(span.Attributes, "messaging.system").AsString()
	Assert(actualSystem == system, "Except messaging.system to be %s, got %s", system, actualSystem)
	Assert(span.SpanKind == trace.SpanKindConsumer, "Expect to be consumer span, got %d", span.SpanKind)
}
