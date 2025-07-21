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

package rocketmq

import (
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

// ProducerRequest represents a RocketMQ producer request
type ProducerRequest struct {
	Message *primitive.Message // The message to be sent
}

// ProducerResponse represents a RocketMQ producer response
type ProducerResponse struct {
	BrokerAddr string                // Address of the broker that handled the request
	Result     *primitive.SendResult // Result of the send operation
}

// ConsumerRequest represents a RocketMQ consumer request
type ConsumerRequest struct {
	BrokerAddr string                // Address of the broker serving the message
	Message    *primitive.MessageExt // The received message
}

// ConsumerResponse represents a RocketMQ consumer processing result
type ConsumerResponse struct {
	Result consumer.ConsumeResult // Result of message consumption
}
