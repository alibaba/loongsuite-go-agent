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
	"unsafe"

	"github.com/alibaba/loongsuite-go-agent/pkg/api"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

//go:linkname producerSendSyncOnEnter github.com/apache/rocketmq-client-go/v2/producer.producerSendSyncOnEnter
func producerSendSyncOnEnter(call api.CallContext, _ interface{}, ctx context.Context, msg *primitive.Message, res *primitive.SendResult) {
	if !rocketmqEnabler.Enable() {
		return
	}

	req := ProducerRequest{
		Message: msg,
	}
	newCtx := producerInst.Start(ctx, req)

	call.SetData(map[string]interface{}{
		"ctx": newCtx,
		"res": res,
	})
}

//go:linkname producerSendSyncOnExit github.com/apache/rocketmq-client-go/v2/producer.producerSendSyncOnExit
func producerSendSyncOnExit(call api.CallContext, err error) {
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

	req := ProducerRequest{
		Message: call.GetParam(2).(*primitive.Message),
	}

	resObj, ok := data["res"]
	if !ok {
		producerInst.End(ctx, req, ProducerResponse{}, err)
		return
	}

	res, _ := resObj.(*primitive.SendResult)
	clientValue := getClientValue(call.GetParam(0))
	if !clientValue.IsValid() {
		producerInst.End(ctx, req, ProducerResponse{Result: res}, err)
		return
	}

	addr := getBrokerAddr(clientValue, res)

	producerInst.End(ctx, req, ProducerResponse{
		Result:     res,
		BrokerAddr: addr,
	}, err)
}

//go:linkname producerSendAsyncOnEnter github.com/apache/rocketmq-client-go/v2/producer.producerSendAsyncOnEnter
func producerSendAsyncOnEnter(call api.CallContext, _ interface{}, ctx context.Context, msg *primitive.Message, originalCallback func(context.Context, *primitive.SendResult, error)) {
	if !rocketmqEnabler.Enable() {
		return
	}

	req := ProducerRequest{
		Message: msg,
	}
	newCtx := producerInst.Start(ctx, req)

	wrappedCallback := func(callbackCtx context.Context, result *primitive.SendResult, callbackErr error) {
		originalCallback(callbackCtx, result, callbackErr)

		// reflection loss performance
		//clientValue := getClientValue(call.GetParam(0))
		if result == nil || result.MessageQueue == nil {
			producerInst.End(ctx, req, ProducerResponse{Result: result}, callbackErr)
			return
		}
		// reflection loss performance
		//addr := getBrokerAddr(clientValue, result)

		producerInst.End(newCtx, req, ProducerResponse{
			Result: result,
			//BrokerAddr: addr,
		}, callbackErr)
	}

	call.SetParam(3, wrappedCallback)
}

//go:linkname producerSendOneWayOnEnter github.com/apache/rocketmq-client-go/v2/producer.producerSendOneWayOnEnter
func producerSendOneWayOnEnter(call api.CallContext, _ interface{}, ctx context.Context, msg *primitive.Message) {
	if !rocketmqEnabler.Enable() {
		return
	}

	req := ProducerRequest{
		Message: msg,
	}
	newCtx := producerInst.Start(ctx, req)

	call.SetData(map[string]interface{}{
		"ctx": newCtx,
		"req": req,
	})
}

//go:linkname producerSendOneWayOnExit github.com/apache/rocketmq-client-go/v2/producer.producerSendOneWayOnExit
func producerSendOneWayOnExit(call api.CallContext, err error) {
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

	req, ok := data["req"].(ProducerRequest)
	if !ok {
		return
	}

	// reflection loss performance
	//clientValue := getClientValue(call.GetParam(0))
	if req.Message == nil || req.Message.Queue == nil {
		producerInst.End(ctx, req, ProducerResponse{}, err)
		return
	}
	// reflection loss performance
	//addr := getBrokerAddrFromMessage(clientValue, req.Message)

	producerInst.End(ctx, req, ProducerResponse{}, err)
}

// getClientValue safely extracts the client value from producer
func getClientValue(producer interface{}) reflect.Value {
	if producer == nil {
		return reflect.Value{}
	}

	val := reflect.ValueOf(producer).Elem()
	if !val.IsValid() {
		return reflect.Value{}
	}

	clientField := val.FieldByName("client")
	if !clientField.IsValid() {
		return reflect.Value{}
	}

	return reflect.NewAt(clientField.Type(), unsafe.Pointer(clientField.UnsafeAddr())).Elem()
}

// getBrokerAddr gets broker address from SendResult
func getBrokerAddr(clientValue reflect.Value, res *primitive.SendResult) string {
	if res == nil || res.MessageQueue == nil || res.MessageQueue.BrokerName == "" {
		return ""
	}
	return callMethod(clientValue, res.MessageQueue.BrokerName)
}

// getBrokerAddrFromMessage gets broker address from Message
func getBrokerAddrFromMessage(clientValue reflect.Value, msg *primitive.Message) string {
	if msg == nil || msg.Queue == nil || msg.Queue.BrokerName == "" {
		return ""
	}
	return callMethod(clientValue, msg.Queue.BrokerName)
}

// callMethod calls a method chain on clientValue
func callMethod(clientValue reflect.Value, brokerName string) string {
	if !clientValue.IsValid() {
		return ""
	}

	getNameSrvMethod := clientValue.MethodByName("GetNameSrv")
	if !getNameSrvMethod.IsValid() {
		return ""
	}

	nameSrvResults := getNameSrvMethod.Call(nil)
	if len(nameSrvResults) == 0 || !nameSrvResults[0].IsValid() {
		return ""
	}

	nameSrv := nameSrvResults[0]
	findMethod := nameSrv.MethodByName("FindBrokerAddrByName")
	if !findMethod.IsValid() {
		return ""
	}

	findResults := findMethod.Call([]reflect.Value{
		reflect.ValueOf(brokerName),
	})

	if len(findResults) == 0 || !findResults[0].IsValid() {
		return ""
	}

	return findResults[0].String()
}
