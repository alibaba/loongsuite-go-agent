//go:build ignore

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

package runtime

import (
	_ "unsafe"
)

// When the program exit, we should call the exit hook to shutdown gracefully.
// See https://github.com/alibaba/loongsuite-go/blob/main/pkg/otel_setup.go
var ExitHook func()

//go:linkname otel_get_trace_context_from_gls otel_get_trace_context_from_gls
var otel_get_trace_context_from_gls = _otel_gls_get_trace_context_impl

//go:linkname otel_get_baggage_container_from_gls otel_get_baggage_container_from_gls
var otel_get_baggage_container_from_gls = _otel_gls_get_baggage_container_impl

//go:linkname otel_set_trace_context_to_gls otel_set_trace_context_to_gls
var otel_set_trace_context_to_gls = _otel_gls_set_trace_context_impl

//go:linkname otel_set_baggage_container_to_gls otel_set_baggage_container_to_gls
var otel_set_baggage_container_to_gls = _otel_gls_set_baggage_container_impl

//go:linkname otel_sync_producer_call_depth_enter otel_sync_producer_call_depth_enter
var otel_sync_producer_call_depth_enter = _otel_sync_producer_call_depth_enter_impl

//go:linkname otel_sync_producer_call_depth_exit otel_sync_producer_call_depth_exit
var otel_sync_producer_call_depth_exit = _otel_sync_producer_call_depth_exit_impl

//go:linkname otel_sync_producer_call_depth_active otel_sync_producer_call_depth_active
var otel_sync_producer_call_depth_active = _otel_sync_producer_call_depth_active_impl

//go:nosplit
func _otel_gls_get_trace_context_impl() interface{} {
	return getg().m.curg.otel_trace_context
}

//go:nosplit
func _otel_gls_get_baggage_container_impl() interface{} {
	return getg().m.curg.otel_baggage_container
}

//go:nosplit
func _otel_gls_set_trace_context_impl(v interface{}) {
	getg().m.curg.otel_trace_context = v
}

//go:nosplit
func _otel_gls_set_baggage_container_impl(v interface{}) {
	getg().m.curg.otel_baggage_container = v
}

//go:nosplit
func _otel_sync_producer_call_depth_enter_impl() {
	getg().m.curg.otel_sync_producer_call_depth++
}

//go:nosplit
func _otel_sync_producer_call_depth_exit_impl() {
	getg().m.curg.otel_sync_producer_call_depth--
}

//go:nosplit
func _otel_sync_producer_call_depth_active_impl() bool {
	return getg().m.curg.otel_sync_producer_call_depth > 0
}

type ContextSnapshoter interface {
	TakeSnapShot() interface{}
}

func contextPropagate(tls interface{}) interface{} {
	if tls == nil {
		return nil
	}
	if taker, ok := tls.(ContextSnapshoter); ok {
		return taker.TakeSnapShot()
	}
	return tls
}
