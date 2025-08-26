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
	"log"
	"sync"
	"time"

	"github.com/alibaba/loongsuite-go-agent/test/verifier"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

const (
	testTimeout   = 1 * time.Microsecond
	timeoutBuffer = 10 * time.Millisecond
)

type TestCase struct {
	name     string
	testFunc func(p rocketmq.Producer)
}

func main() {
	// Initialize test environment
	initTopic()
	producer := initProducer()
	defer producer.Shutdown()

	if err := producer.Start(); err != nil {
		log.Fatalf("Failed to start producer: %v", err)
	}

	// Define test cases
	testCases := []TestCase{
		{"Sync Send Success", testSyncSendSuccess},
		{"Sync Send Failure", testSyncSendFailure},
		{"Async Send Success", testAsyncSendSuccess},
		{"OneWay Send Success", testOneWaySendSuccess},
	}

	// Execute test cases
	for _, tc := range testCases {
		log.Printf("\n===== Testing %s =====\n", tc.name)
		tc.testFunc(producer)
	}
}

func testSyncSendSuccess(p rocketmq.Producer) {
	msg := createMessage("测试同步发送", "test_sync")
	result, err := p.SendSync(context.Background(), msg)
	if err != nil {
		log.Printf("Sync send failed: %v\n", err)
		return
	}
	log.Printf("Sync send succeeded: %s\n", result.String())
	verifyProdcuerTraces(topicName, "test_sync", false)
}

func testSyncSendFailure(p rocketmq.Producer) {
	msg := createMessage("测试同步发送失败", "test_sync_failure")
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
	defer cancel()
	time.Sleep(timeoutBuffer)

	_, err := p.SendSync(ctx, msg)
	if err == nil {
		log.Println("Error: Expected sync send to fail but it succeeded")
		return
	}
	log.Printf("Expected sync send failure: %v\n", err)
	verifyProdcuerTraces(topicName, "test_sync_failure", true)
}

func testAsyncSendSuccess(p rocketmq.Producer) {
	msg := createMessage("测试异步发送", "test_async")
	var wg sync.WaitGroup
	wg.Add(1)

	err := p.SendAsync(context.Background(), func(ctx context.Context, result *primitive.SendResult, err error) {
		defer wg.Done()
		if err != nil {
			log.Printf("Async callback error: %v\n", err)
			panic(err)
		} else {
			log.Printf("Async send succeeded: %s\n", result.String())
		}
	}, msg)

	if err != nil {
		log.Printf("Async send request failed: %v\n", err)
		panic(err)
	}
	wg.Wait()
	time.Sleep(timeoutBuffer)
	verifyProdcuerTraces(topicName, "test_async", false)
}

func testOneWaySendSuccess(p rocketmq.Producer) {
	msg := createMessage("测试单向发送", "test_oneway")
	if err := p.SendOneWay(context.Background(), msg); err != nil {
		log.Printf("OneWay send failed: %v\n", err)
		panic(err)
	}
	log.Println("OneWay send completed")
	verifyProdcuerTraces(topicName, "test_oneway", false)
}

func createMessage(body, tag string) *primitive.Message {
	msg := primitive.NewMessage(topicName, []byte(body))
	msg.WithTag(tag)
	return msg
}

func verifyProdcuerTraces(topic, tag string, expectError bool) {
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		span := stubs[0][0]
		VerifyRocketMQProduceAttributes(span, topic, tag, "", "publish", expectError)
	}, 1)
}
