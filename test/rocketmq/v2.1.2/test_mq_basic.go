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
	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"log"
	"time"
)

const (
	consumerStartupDelay = 3 * time.Second  // Time for consumer to initialize
	testMessageContent   = "Hello RocketMQ" // Test message content
	testTag              = "test_tag"       // Test message tag
)

func main() {
	// Initialize test environment
	initTopic()
	p := initProducer()
	defer p.Shutdown()

	c := initConsumer()
	defer c.Shutdown()

	// Prepare and send test message
	msg := primitive.NewMessage(topicName, []byte(testMessageContent))
	msg.WithTag(testTag)

	result, err := p.SendSync(context.Background(), msg)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	log.Printf("Message sent successfully: %s", result.String())

	// Register message handler
	err = c.Subscribe(topicName, consumer.MessageSelector{}, createMessageHandler())
	if err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}

	// Start consumer with delay to ensure it's ready
	if err = c.Start(); err != nil {
		log.Fatalf("Failed to start consumer: %v", err)
	}
	time.Sleep(consumerStartupDelay)

	// Verify OpenTelemetry traces
	verifyBasicTraces()

	log.Println("Test completed successfully")
}

// createMessageHandler creates a message handler function
func createMessageHandler() func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	return func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			log.Printf("Received message: %s", string(msg.Body))
		}
		return consumer.ConsumeSuccess, nil
	}
}

// verifyBasicTraces verifies the OpenTelemetry traces
func verifyBasicTraces() {
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		if len(stubs) == 0 || len(stubs[0]) < 2 {
			log.Fatal("Insufficient spans collected for verification")
		}
		VerifyRocketMQAttributes(stubs[0][0], stubs[0][1], topicName, testTag)
	}, 1)
}
