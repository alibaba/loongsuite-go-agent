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

package rueidisgo

import (
	"context"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/alibaba/loongsuite-go/pkg/rules/rueidisgo/rueidisnil"
	"github.com/redis/rueidis"
	"go.opentelemetry.io/otel/trace"
	"os"
	"strings"
	"time"
	_ "unsafe"
)

type rueidisInnerEnabler struct {
	enabled bool
}

func (r rueidisInnerEnabler) Enable() bool {
	return r.enabled
}

var goRueidisInstrumenter = BuildGoRueidisOtelInstrumenter()

var rueidisStartOptions = []trace.SpanStartOption{}

var rueidisEnabler = rueidisInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_REDIGO_ENABLED") != "false"}

//go:linkname rueidisNewClientOnEnter github.com/redis/rueidis.rueidisNewClientOnEnter
func rueidisNewClientOnEnter(call api.CallContext, option rueidis.ClientOption) {
	if option.InitAddress != nil && len(option.InitAddress) > 0 {
		rueidis.RedisAdders = option.InitAddress[0]
	}
}

//go:linkname clusterClientDoOnEnter github.com/redis/rueidis.clusterClientDoOnEnter
func clusterClientDoOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoOnExit github.com/redis/rueidis.clusterClientDoOnExit
func clusterClientDoOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname clusterClientDoMultiOnEnter github.com/redis/rueidis.clusterClientDoMultiOnEnter
func clusterClientDoMultiOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoMultiOnExit github.com/redis/rueidis.clusterClientDoMultiOnExit
func clusterClientDoMultiOnExit(call api.CallContext, resp []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(resp))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(resp))
	}
}

//
//go:linkname clusterClientReceiveOnEnter github.com/redis/rueidis.clusterClientReceiveOnEnter
func clusterClientReceiveOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) {
	cc := processCommand(subscribe)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientReceiveOnExit github.com/redis/rueidis.clusterClientReceiveOnExit
func clusterClientReceiveOnExit(call api.CallContext, err error) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(err))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(err))
	}
}

//go:linkname clusterClientDoCacheOnEnter github.com/redis/rueidis.clusterClientDoCacheOnEnter
func clusterClientDoCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Cacheable, ttl time.Duration) {
	cc := processCache(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoCacheOnExit github.com/redis/rueidis.clusterClientDoCacheOnExit
func clusterClientDoCacheOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname clusterClientDoMultiCacheOnEnter github.com/redis/rueidis.clusterClientDoMultiCacheOnEnter
func clusterClientDoMultiCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.CacheableTTL) {
	cc := processCacheMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoMultiCacheOnExit github.com/redis/rueidis.clusterClientDoMultiCacheOnExit
func clusterClientDoMultiCacheOnExit(call api.CallContext, r []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(r))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(r))
	}
}

//go:linkname clusterClientDoStreamOnEnter github.com/redis/rueidis.clusterClientDoStreamOnEnter
func clusterClientDoStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoStreamOnExit github.com/redis/rueidis.clusterClientDoStreamOnExit
func clusterClientDoStreamOnExit(call api.CallContext, r rueidis.RedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname clusterClientDoMultiStreamOnEnter github.com/redis/rueidis.clusterClientDoMultiStreamOnEnter
func clusterClientDoMultiStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname clusterClientDoMultiStreamOnExit github.com/redis/rueidis.clusterClientDoMultiStreamOnExit
func clusterClientDoMultiStreamOnExit(call api.CallContext, r rueidis.MultiRedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname singleClientDoOnEnter github.com/redis/rueidis.singleClientDoOnEnter
func singleClientDoOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoOnExit github.com/redis/rueidis.singleClientDoOnExit
func singleClientDoOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname singleClientDoMultiOnEnter github.com/redis/rueidis.singleClientDoMultiOnEnter
func singleClientDoMultiOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoMultiOnExit github.com/redis/rueidis.singleClientDoMultiOnExit
func singleClientDoMultiOnExit(call api.CallContext, resp []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(resp))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(resp))
	}
}

//
//go:linkname singleClientReceiveOnEnter github.com/redis/rueidis.singleClientReceiveOnEnter
func singleClientReceiveOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) {
	cc := processCommand(subscribe)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientReceiveOnExit github.com/redis/rueidis.singleClientReceiveOnExit
func singleClientReceiveOnExit(call api.CallContext, err error) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(err))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(err))
	}
}

//go:linkname singleClientDoCacheOnEnter github.com/redis/rueidis.singleClientDoCacheOnEnter
func singleClientDoCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Cacheable, ttl time.Duration) {
	cc := processCache(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoCacheOnExit github.com/redis/rueidis.singleClientDoCacheOnExit
func singleClientDoCacheOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname singleClientDoMultiCacheOnEnter github.com/redis/rueidis.singleClientDoMultiCacheOnEnter
func singleClientDoMultiCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.CacheableTTL) {
	cc := processCacheMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoMultiCacheOnExit github.com/redis/rueidis.singleClientDoMultiCacheOnExit
func singleClientDoMultiCacheOnExit(call api.CallContext, r []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(r))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(r))
	}
}

//go:linkname singleClientDoStreamOnEnter github.com/redis/rueidis.singleClientDoStreamOnEnter
func singleClientDoStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoStreamOnExit github.com/redis/rueidis.singleClientDoStreamOnExit
func singleClientDoStreamOnExit(call api.CallContext, r rueidis.RedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname singleClientDoMultiStreamOnEnter github.com/redis/rueidis.singleClientDoMultiStreamOnEnter
func singleClientDoMultiStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname singleClientDoMultiStreamOnExit github.com/redis/rueidis.singleClientDoMultiStreamOnExit
func singleClientDoMultiStreamOnExit(call api.CallContext, r rueidis.MultiRedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname sentinelClientDoOnEnter github.com/redis/rueidis.sentinelClientDoOnEnter
func sentinelClientDoOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoOnExit github.com/redis/rueidis.sentinelClientDoOnExit
func sentinelClientDoOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname sentinelClientDoMultiOnEnter github.com/redis/rueidis.sentinelClientDoMultiOnEnter
func sentinelClientDoMultiOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoMultiOnExit github.com/redis/rueidis.sentinelClientDoMultiOnExit
func sentinelClientDoMultiOnExit(call api.CallContext, resp []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(resp))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(resp))
	}
}

//
//go:linkname sentinelClientReceiveOnEnter github.com/redis/rueidis.sentinelClientReceiveOnEnter
func sentinelClientReceiveOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) {
	cc := processCommand(subscribe)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientReceiveOnExit github.com/redis/rueidis.sentinelClientReceiveOnExit
func sentinelClientReceiveOnExit(call api.CallContext, err error) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(err))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(err))
	}
}

//go:linkname sentinelClientDoCacheOnEnter github.com/redis/rueidis.sentinelClientDoCacheOnEnter
func sentinelClientDoCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Cacheable, ttl time.Duration) {
	cc := processCache(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoCacheOnExit github.com/redis/rueidis.sentinelClientDoCacheOnExit
func sentinelClientDoCacheOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname sentinelClientDoMultiCacheOnEnter github.com/redis/rueidis.sentinelClientDoMultiCacheOnEnter
func sentinelClientDoMultiCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.CacheableTTL) {
	cc := processCacheMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoMultiCacheOnExit github.com/redis/rueidis.sentinelClientDoMultiCacheOnExit
func sentinelClientDoMultiCacheOnExit(call api.CallContext, r []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(r))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(r))
	}
}

//go:linkname sentinelClientDoStreamOnEnter github.com/redis/rueidis.sentinelClientDoStreamOnEnter
func sentinelClientDoStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoStreamOnExit github.com/redis/rueidis.sentinelClientDoStreamOnExit
func sentinelClientDoStreamOnExit(call api.CallContext, r rueidis.RedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname sentinelClientDoMultiStreamOnEnter github.com/redis/rueidis.sentinelClientDoMultiStreamOnEnter
func sentinelClientDoMultiStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname sentinelClientDoMultiStreamOnExit github.com/redis/rueidis.sentinelClientDoMultiStreamOnExit
func sentinelClientDoMultiStreamOnExit(call api.CallContext, r rueidis.MultiRedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname standaloneDoOnEnter github.com/redis/rueidis.standaloneDoOnEnter
func standaloneDoOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoOnExit github.com/redis/rueidis.standaloneDoOnExit
func standaloneDoOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname standaloneDoMultiOnEnter github.com/redis/rueidis.standaloneDoMultiOnEnter
func standaloneDoMultiOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoMultiOnExit github.com/redis/rueidis.standaloneDoMultiOnExit
func standaloneDoMultiOnExit(call api.CallContext, resp []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(resp))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(resp))
	}
}

//
//go:linkname standaloneReceiveOnEnter github.com/redis/rueidis.standaloneReceiveOnEnter
func standaloneReceiveOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, subscribe rueidis.Completed, fn func(msg rueidis.PubSubMessage)) {
	cc := processCommand(subscribe)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneReceiveOnExit github.com/redis/rueidis.standaloneReceiveOnExit
func standaloneReceiveOnExit(call api.CallContext, err error) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(err))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(err))
	}
}

//go:linkname standaloneDoCacheOnEnter github.com/redis/rueidis.standaloneDoCacheOnEnter
func standaloneDoCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Cacheable, ttl time.Duration) {
	cc := processCache(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoCacheOnExit github.com/redis/rueidis.standaloneDoCacheOnExit
func standaloneDoCacheOnExit(call api.CallContext, r rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname standaloneDoMultiCacheOnEnter github.com/redis/rueidis.standaloneDoMultiCacheOnEnter
func standaloneDoMultiCacheOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.CacheableTTL) {
	cc := processCacheMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoMultiCacheOnExit github.com/redis/rueidis.standaloneDoMultiCacheOnExit
func standaloneDoMultiCacheOnExit(call api.CallContext, r []rueidis.RedisResult) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, firstError(r))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, firstError(r))
	}
}

//go:linkname standaloneDoStreamOnEnter github.com/redis/rueidis.standaloneDoStreamOnEnter
func standaloneDoStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, cmd rueidis.Completed) {
	cc := processCommand(cmd)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoStreamOnExit github.com/redis/rueidis.standaloneDoStreamOnExit
func standaloneDoStreamOnExit(call api.CallContext, r rueidis.RedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}

//go:linkname standaloneDoMultiStreamOnEnter github.com/redis/rueidis.standaloneDoMultiStreamOnEnter
func standaloneDoMultiStreamOnEnter(call api.CallContext, client interface{}, ctx1 context.Context, multi ...rueidis.Completed) {
	cc := processCommandMulti(multi)
	if cc.cmdName == "" {
		return
	}
	if !rueidisEnabler.Enable() {
		return
	}

	request := &goRueidisRequest{
		cmd:      cc,
		endpoint: rueidis.RedisAdders,
	}
	ctx1 = goRueidisInstrumenter.Start(ctx1, request, rueidisStartOptions...)
	call.SetParam(1, ctx1)
	data := make(map[string]interface{}, 1)
	data["ctx"] = ctx1
	data["request"] = request
	call.SetData(data)
	return
}

//go:linkname standaloneDoMultiStreamOnExit github.com/redis/rueidis.standaloneDoMultiStreamOnExit
func standaloneDoMultiStreamOnExit(call api.CallContext, r rueidis.MultiRedisResultStream) {
	if !rueidisEnabler.Enable() || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil || data["ctx"] == nil {
		return
	}
	ctx := data["ctx"].(context.Context)
	request := data["request"].(*goRueidisRequest)
	if request != nil {
		goRueidisInstrumenter.End(ctx, request, nil, rueidisnil.SpanEndErr(r.Error()))
	} else {
		goRueidisInstrumenter.End(ctx, &goRueidisRequest{}, nil, rueidisnil.SpanEndErr(r.Error()))
	}
}
func firstError(s []rueidis.RedisResult) error {
	for _, result := range s {
		if err := result.Error(); rueidisnil.IsSpanError(err) {
			return err
		}
	}
	return nil
}

func processCommandMulti(multi []rueidis.Completed) command {
	var cmds []command
	for _, cmd := range multi {
		cmds = append(cmds, processCommand(cmd))
	}
	return multiCommand(cmds)
}

func processCacheMulti(multi []rueidis.CacheableTTL) command {
	var cmds []command
	for _, cmd := range multi {
		cmds = append(cmds, processCache(cmd.Cmd))
	}
	return multiCommand(cmds)
}

func multiCommand(cmds []command) command {
	// limit to the 5 first
	if len(cmds) > 5 {
		cmds = cmds[:5]
	}
	statement := strings.Builder{}
	raw := strings.Builder{}
	for i, cmd := range cmds {
		statement.WriteString(cmd.cmdName)
		raw.WriteString(cmd.statement)
		if i != len(cmds)-1 {
			statement.WriteString(" ")
			raw.WriteString(" ")
		}
	}
	return command{
		cmdName:   statement.String(),
		statement: raw.String(),
	}
}

func processCache(cmd rueidis.Cacheable) command {
	cmds := cmd.Commands()
	if len(cmds) == 0 {
		return command{}
	}
	statement := cmds[0]
	raw := strings.Join(cmds, " ")
	return command{
		cmdName:   statement,
		statement: raw,
	}
}

func processCommand(cmd rueidis.Completed) command {
	cmds := cmd.Commands()
	if len(cmds) == 0 {
		return command{}
	}
	statement := cmds[0]
	raw := strings.Join(cmds, " ")
	return command{
		cmdName:   statement,
		statement: raw,
	}
}
