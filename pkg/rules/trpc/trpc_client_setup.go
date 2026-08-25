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

package trpc

import (
	"context"
	"strings"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"trpc.group/trpc-go/trpc-go/client"
	"trpc.group/trpc-go/trpc-go/codec"
)

var trpcClientInstrumenter = BuildTrpcClientInstrumenter()

// func (c *client) Invoke(ctx context.Context, reqBody interface{}, rspBody interface{}, opt ...Option) (err error)
//
//go:linkname clientTrpcOnEnter trpc.group/trpc-go/trpc-go/client.clientTrpcOnEnter
func clientTrpcOnEnter(call api.CallContext, _ interface{}, ctx context.Context, reqBody interface{}, rspBody interface{}, opts ...client.Option) {
	if !trpcEnabler.Enable() {
		return
	}
	msg := codec.Message(ctx)

	// if the caller service is null, name it to `service`
	// https://github.com/trpc-group/trpc-go/blob/e025145c92d41417fb71574fb486441e629804ac/codec.go#L526
	if msg.CallerService() == "" {
		msg.WithCallerService("service")
	}

	inputOpts := &client.Options{}
	for _, o := range opts {
		o(inputOpts)
	}
	addr := inputOpts.Target

	req := trpcReq{
		msg:  msg,
		addr: parseTarget(addr),
	}
	newCtx := trpcClientInstrumenter.Start(context.Background(), req)
	data := make(map[string]interface{}, 3)
	data["ctx"] = newCtx
	data["request"] = req
	call.SetData(data)
}

// func (c *client) Invoke(ctx context.Context, reqBody interface{}, rspBody interface{}, opt ...Option) (err error)
//
//go:linkname clientTrpcOnExit trpc.group/trpc-go/trpc-go/client.clientTrpcOnExit
func clientTrpcOnExit(call api.CallContext, err error) {
	if !trpcEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)
	request := data["request"].(trpcReq)
	statusCode := 0
	if request.msg.ServerRspErr() != nil {
		statusCode = int(request.msg.ServerRspErr().Code)
	}
	trpcClientInstrumenter.End(ctx, request, trpcRes{
		statusCode: statusCode,
	}, err)
}

// https://github.com/trpc-group/trpc-go/blob/e025145c92d41417fb71574fb486441e629804ac/client/options.go#L667
func parseTarget(target string) string {
	if target == "" {
		return ""
	}
	substr := "://"
	index := strings.Index(target, substr)
	if index == -1 {
		return ""
	}
	return target[index+len(substr):]
}
