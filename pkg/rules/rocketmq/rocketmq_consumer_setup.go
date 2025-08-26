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
	"context"
	"reflect"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

//go:linkname consumerConsumeInnerOnEnter github.com/apache/rocketmq-client-go/v2/consumer.consumerConsumeInnerOnEnter
func consumerConsumeInnerOnEnter(call api.CallContext, _ interface{}, ctx context.Context, msgs []*primitive.MessageExt) {
	if !rocketmqEnabler.Enable() {
		return
	}

	var addr string
	// reflection loss performance
	//if len(msgs) > 0 && msgs[0] != nil && msgs[0].Queue != nil {
	//	if clientValue := getClientValue(call); clientValue.IsValid() {
	//		addr = getBrokerAddrFromMessageExt(clientValue, msgs[0])
	//	}
	//}

	// Store context and messages directly
	call.SetData(map[string]interface{}{
		"ctx":  Start(ctx, msgs, addr),
		"msgs": msgs,
	})
}

//go:linkname consumerConsumeInnerOnExit github.com/apache/rocketmq-client-go/v2/consumer.consumerConsumeInnerOnExit
func consumerConsumeInnerOnExit(call api.CallContext, consumeResult consumer.ConsumeResult, err error) {
	if !rocketmqEnabler.Enable() {
		return
	}

	data, ok := call.GetData().(map[string]interface{})
	if !ok {
		return
	}

	ctx, ok := data["ctx"].(context.Context)
	if !ok {
		return
	}

	msgs, _ := data["msgs"].([]*primitive.MessageExt)
	End(ctx, msgs, consumeResult, err)
}

// getBrokerAddrFromMessage gets broker address from Message
func getBrokerAddrFromMessageExt(clientValue reflect.Value, msg *primitive.MessageExt) string {
	if msg == nil || msg.Queue == nil || msg.Queue.BrokerName == "" {
		return ""
	}
	return callMethod(clientValue, msg.Queue.BrokerName)
}
