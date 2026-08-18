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
	"errors"
	_ "unsafe"

	"github.com/alibaba/loongsuite-go/pkg/api"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
)

var trpcServerInstrumenter = BuildTrpcServerInstrumenter()

// func (s *service) handle(ctx context.Context, msg codec.Msg, reqBodyBuf []byte) (interface{}, error)
//
//go:linkname serverTrpcOnEnter trpc.group/trpc-go/trpc-go/server.serverTrpcOnEnter
func serverTrpcOnEnter(call api.CallContext, _ interface{}, ctx context.Context, msg codec.Msg, reqBodyBuf []byte) {
	if !trpcEnabler.Enable() {
		return
	}
	request := trpcReq{
		msg: msg,
	}
	newCtx := trpcServerInstrumenter.Start(ctx, request)
	data := make(map[string]interface{}, 3)
	data["ctx"] = newCtx
	data["request"] = request
	call.SetData(data)
}

//go:linkname serverTrpcOnExit trpc.group/trpc-go/trpc-go/server.serverTrpcOnExit
func serverTrpcOnExit(call api.CallContext, _ interface{}, err error) {
	if !trpcEnabler.Enable() {
		return
	}
	data := call.GetData().(map[string]interface{})
	ctx := data["ctx"].(context.Context)
	request := data["request"].(trpcReq)
	statusCode := 0
	if err != nil {
		var trpcErr *errs.Error
		if errors.As(err, &trpcErr) {
			statusCode = int(trpcErr.Code)
		}
	}
	trpcServerInstrumenter.End(ctx, request, trpcRes{
		statusCode: statusCode,
	}, err)
}
