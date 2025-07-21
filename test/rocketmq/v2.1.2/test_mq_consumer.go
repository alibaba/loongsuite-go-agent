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
	"fmt"
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/test/verifier"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"log"
	"time"
)

const (
	messageCount       = 2
	consumerReadyDelay = 5 * time.Second
)

func main() {
	// Initialize test environment
	initTopic()
	producer := initProducer()
	defer producer.Shutdown()

	// Initialize cluster consumer
	clusterConsumer := initClusterConsumer()
	defer clusterConsumer.Shutdown()

	// Prepare and send test messages
	messages := prepareTestMessages()
	sendMessages(producer, messages)

	// Start the cluster consumer
	clusterConsumer.Start()
	time.Sleep(consumerReadyDelay)
	// Verify OpenTelemetry traces
	verifyConsumerTraces()
}

// initClusterConsumer initializes and configures a cluster consumer
func initClusterConsumer() rocketmq.PushConsumer {
	c := initConsumer()

	err := c.Subscribe(topicName, consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			log.Printf("msg count: %d\n", len(msgs))
			for _, msg := range msgs {
				log.Printf("Cluster mode consumption: %s, Tags: %s\n",
					string(msg.Body), msg.GetTags())
			}
			return consumer.ConsumeSuccess, nil
		})
	if err != nil {
		log.Fatalf("Failed to subscribe topic (cluster mode): %v", err)
		panic(err)
	}

	return c
}

// prepareTestMessages creates test messages with unique tags
func prepareTestMessages() []*primitive.Message {
	messages := make([]*primitive.Message, messageCount)
	for i := 0; i < messageCount; i++ {
		msg := &primitive.Message{
			Topic: topicName,
			Body:  []byte(fmt.Sprintf("Consumption mode test message %d", i)),
		}
		msg.WithTag(fmt.Sprintf("Tag%d", i))
		messages[i] = msg
	}
	return messages
}

// sendMessages sends messages and handles errors
func sendMessages(p rocketmq.Producer, messages []*primitive.Message) {
	result, err := p.SendSync(context.Background(), messages...)
	if err != nil {
		log.Fatalf("Failed to send messages: %v", err)
		panic(err)
	}
	log.Printf("Messages sent successfully: %s\n", result.MsgID)
}

// verifyConsumerTraces verifies the OpenTelemetry traces
func verifyConsumerTraces() {
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		VerifyRocketMQReceive(stubs[0][0], stubs[1][0], stubs[1][1])
		VerifyRocketMQConsumeAttributes(stubs[1][1], topicName, "Tag0", "", "process", false)
		VerifyRocketMQConsumeAttributes(stubs[1][2], topicName, "Tag1", "", "process", false)
	}, 2)
}
