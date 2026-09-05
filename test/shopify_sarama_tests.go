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

package test

import (
	"context"
	"testing"
)

const shopifySaramaModuleName = "shopify-sarama"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("shopify-sarama-basic-test", shopifySaramaModuleName, "1.22.0", "", "1.22.0", "", TestShopifySaramaBasic),
		NewGeneralTestCase("shopify-sarama-async-produce-test", shopifySaramaModuleName, "1.22.0", "", "1.22.0", "", TestShopifySaramaAsyncProduce),
		NewGeneralTestCase("shopify-sarama-sync-batch-produce-test", shopifySaramaModuleName, "1.22.0", "", "1.22.0", "", TestShopifySaramaSyncBatchProduce),
		NewGeneralTestCase("shopify-sarama-concurrent-sync-producer-test", shopifySaramaModuleName, "1.22.0", "", "1.22.0", "", TestShopifySaramaConcurrentSyncProducer),
	)
}

func TestShopifySaramaConcurrentSyncProducer(t *testing.T, env ...string) {
	containers := initKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())
	UseApp("shopify-sarama/v1.22.0")
	RunGoBuild(t, "go", "build", "test_concurrent_sync_producer.go", "base.go")
	env = append(env, "KAFKA_ADDR="+containers.KafkaAddress)
	RunApp(t, "test_concurrent_sync_producer", env...)
}

func TestShopifySaramaBasic(t *testing.T, env ...string) {
	containers := initKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())
	UseApp("shopify-sarama/v1.22.0")
	RunGoBuild(t, "go", "build", "test_sarama_basic.go", "base.go")
	env = append(env, "KAFKA_ADDR="+containers.KafkaAddress)
	RunApp(t, "test_sarama_basic", env...)
}

func TestShopifySaramaAsyncProduce(t *testing.T, env ...string) {
	containers := initKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())
	UseApp("shopify-sarama/v1.22.0")
	RunGoBuild(t, "go", "build", "test_async_produce.go", "base.go")
	env = append(env, "KAFKA_ADDR="+containers.KafkaAddress)
	RunApp(t, "test_async_produce", env...)
}

func TestShopifySaramaSyncBatchProduce(t *testing.T, env ...string) {
	containers := initKafkaContainer(t)
	defer containers.CleanupContainers(context.Background())
	UseApp("shopify-sarama/v1.22.0")
	RunGoBuild(t, "go", "build", "test_sync_batch_produce.go", "base.go")
	env = append(env, "KAFKA_ADDR="+containers.KafkaAddress)
	RunApp(t, "test_sync_batch_produce", env...)
}
