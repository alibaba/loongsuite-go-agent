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
	"context"
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/test/verifier"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/admin"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"log"
	"os"
	"time"
)

const (
	topicName            = "test_topic"
	groupName            = "test_group"
	defaultNamingSrvAddr = "127.0.0.1:9876"
	defaultBrokerAddr    = "127.0.0.1:10911"
	topicCreateTimeout   = 5 * time.Second
)

// initProducer initializes and returns a RocketMQ producer
func initProducer() rocketmq.Producer {
	// Get NameServer address from environment variable
	nameSrvAddr := os.Getenv("NAMESRV_ADDR")
	if nameSrvAddr == "" {
		nameSrvAddr = defaultNamingSrvAddr
	}

	// Create producer
	p, err := rocketmq.NewProducer(
		producer.WithNameServer([]string{nameSrvAddr}),
		producer.WithGroupName(groupName),
		producer.WithRetry(2),
	)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
		panic(err)
	}

	// Start producer
	if err = p.Start(); err != nil {
		log.Fatalf("Failed to start producer: %v", err)
		panic(err)
	}

	return p
}

// initTopic creates the test topic if it doesn't exist
func initTopic() {
	nameSrvAddr := os.Getenv("NAMESRV_ADDR")
	if nameSrvAddr == "" {
		nameSrvAddr = defaultNamingSrvAddr
	}
	brokerAddr := os.Getenv("BROKER_ADDR")
	if brokerAddr == "" {
		brokerAddr = defaultBrokerAddr
	}

	// Create admin client
	adminClient, err := admin.NewAdmin(
		admin.WithResolver(primitive.NewPassthroughResolver([]string{nameSrvAddr})),
	)
	if err != nil {
		log.Printf("Failed to create admin client: %v. Relying on RocketMQ auto topic creation", err)
		return
	}
	defer adminClient.Close()

	// Create topic with timeout
	ctx, cancel := context.WithTimeout(context.Background(), topicCreateTimeout)
	defer cancel()

	err = adminClient.CreateTopic(
		ctx,
		admin.WithTopicCreate(topicName),
		admin.WithBrokerAddrCreate(brokerAddr),
	)
	if err != nil {
		log.Printf("Failed to create topic: %v. Relying on RocketMQ auto topic creation", err)
		panic(err)
	} else {
		log.Printf("Topic %s created successfully", topicName)
	}
}

// initConsumer creates and configures a RocketMQ push consumer
func initConsumer() rocketmq.PushConsumer {
	// Get NameServer address from environment variable
	nameSrvAddr := os.Getenv("NAMESRV_ADDR")
	if nameSrvAddr == "" {
		nameSrvAddr = defaultNamingSrvAddr
	}

	// Create consumer instance
	c, err := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{nameSrvAddr}),
		consumer.WithGroupName(groupName),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeMessageBatchMaxSize(10),
	)
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
		panic(err)
	}

	return c
}

// VerifyRocketMQAttributes verifies the span attributes for RocketMQ producer and consumer
func VerifyRocketMQAttributes(producer tracetest.SpanStub, consumer tracetest.SpanStub, topic string, tag string) {
	verifier.VerifyMQPublishAttributes(producer, "", "", topic, "publish", topic, "rocketmq")
	verifier.VerifyMQConsumeAttributes(consumer, "", "", topic, "process", topic, "rocketmq")

	// Verify trace context between producer and consumer
	verifier.Assert(consumer.Parent.SpanID() == producer.SpanContext.SpanID(),
		"Expected span ID %s, got %s", producer.SpanContext.SpanID(), consumer.Parent.SpanID())
	verifier.Assert(consumer.SpanContext.TraceID() == producer.SpanContext.TraceID(),
		"Expected trace ID %s, got %s", producer.SpanContext.TraceID(), consumer.SpanContext.TraceID())
}

// VerifyRocketMQProduceAttributes verifies span attributes for RocketMQ producer
func VerifyRocketMQProduceAttributes(span tracetest.SpanStub, topic string, tag string, key string, operationName string, expectedError bool) {
	// Verify basic message attributes
	verifier.Assert(span.Name == topic+" "+operationName,
		"Expected span name %s, got %s", topic+" "+operationName, span.Name)

	// Verify standard messaging attributes
	verifyMessagingAttributes(span, topic, operationName)

	// Verify message body size
	bodySize := verifier.GetAttribute(span.Attributes, "messaging.message.body.size").AsInt64()
	verifier.Assert(bodySize > 0, "Expected message body size > 0, got %d", bodySize)

	// Verify RocketMQ specific attributes
	if tag != "" {
		verifyAttribute(span, "messaging.rocketmq.message.tag", tag)
	}
	if key != "" {
		verifyAttribute(span, "messaging.rocketmq.message.keys", key)
	}

	// Verify span kind
	verifier.Assert(span.SpanKind == trace.SpanKindProducer,
		"Expected producer span, got %d", span.SpanKind)

	// Verify error status
	verifyErrorStatus(span, expectedError)
}

// VerifyRocketMQConsumeAttributes verifies span attributes for RocketMQ consumer
func VerifyRocketMQConsumeAttributes(span tracetest.SpanStub, topic string, tag string, key string, operationName string, expectedError bool) {
	// Verify basic message attributes
	verifier.Assert(span.Name == topic+" "+operationName,
		"Expected span name %s, got %s", topic+" "+operationName, span.Name)

	// Verify standard messaging attributes
	verifyMessagingAttributes(span, topic, operationName)

	// Verify message body size
	bodySize := verifier.GetAttribute(span.Attributes, "messaging.message.body.size").AsInt64()
	verifier.Assert(bodySize > 0, "Expected message body size > 0, got %d", bodySize)

	// Verify RocketMQ specific attributes
	if tag != "" {
		verifyAttribute(span, "messaging.rocketmq.message.tag", tag)
	}
	if key != "" {
		verifyAttribute(span, "messaging.rocketmq.message.keys", key)
	}

	// Verify span kind
	verifier.Assert(span.SpanKind == trace.SpanKindConsumer,
		"Expected consumer span, got %d", span.SpanKind)

	// Verify error status
	verifyErrorStatus(span, expectedError)
}

// VerifyRocketMQReceive verifies span attributes for RocketMQ receive operation
func VerifyRocketMQReceive(producer tracetest.SpanStub, receive tracetest.SpanStub, process tracetest.SpanStub) {
	// Verify basic message attributes
	verifier.Assert(receive.Name == "multiple_sources receive",
		"Expected span name 'multiple_sources receive', got %s", receive.Name)

	// Verify standard messaging attributes
	actualSystem := verifier.GetAttribute(receive.Attributes, "messaging.system").AsString()
	verifier.Assert(actualSystem == "rocketmq",
		"Expected messaging.system 'rocketmq', got %s", actualSystem)

	// Verify span kind
	verifier.Assert(receive.SpanKind == trace.SpanKindConsumer,
		"Expected consumer span, got %d", receive.SpanKind)

	// Verify trace context
	verifier.Assert(process.Links[0].SpanContext.TraceID() == producer.SpanContext.TraceID(),
		"Expected trace ID %s, got %s", producer.SpanContext.TraceID(), process.Links[0].SpanContext.TraceID())
}

// verifyMessagingAttributes verifies common messaging attributes
func verifyMessagingAttributes(span tracetest.SpanStub, topic string, operation string) {
	actualSystem := verifier.GetAttribute(span.Attributes, "messaging.system").AsString()
	verifier.Assert(actualSystem == "rocketmq",
		"Expected messaging.system 'rocketmq', got %s", actualSystem)

	actualDestination := verifier.GetAttribute(span.Attributes, "messaging.destination.name").AsString()
	verifier.Assert(actualDestination == topic,
		"Expected messaging.destination.name '%s', got %s", topic, actualDestination)

	actualOperation := verifier.GetAttribute(span.Attributes, "messaging.operation.name").AsString()
	verifier.Assert(actualOperation == operation,
		"Expected messaging.operation.name '%s', got %s", operation, actualOperation)
}

// verifyAttribute verifies a specific attribute value
func verifyAttribute(span tracetest.SpanStub, key string, expectedValue string) {
	actualValue := verifier.GetAttribute(span.Attributes, key).AsString()
	verifier.Assert(actualValue == expectedValue,
		"Expected %s '%s', got %s", key, expectedValue, actualValue)
}

// verifyErrorStatus verifies the span error status
func verifyErrorStatus(span tracetest.SpanStub, expectedError bool) {
	if expectedError {
		verifier.Assert(span.Status.Code == codes.Error,
			"Expected error status, got %s", span.Status.Code)
		verifier.Assert(span.Status.Description != "",
			"Expected non-empty error description")
	} else {
		verifier.Assert(span.Status.Code != codes.Error,
			"Expected non-error status, got %s", span.Status.Code)
	}
}
