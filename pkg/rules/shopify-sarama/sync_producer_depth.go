// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package sarama

import _ "unsafe"

//go:linkname syncProducerCallDepthEnter otel_sync_producer_call_depth_enter
var syncProducerCallDepthEnter func()

//go:linkname syncProducerCallDepthExit otel_sync_producer_call_depth_exit
var syncProducerCallDepthExit func()

//go:linkname syncProducerCallDepthActive otel_sync_producer_call_depth_active
var syncProducerCallDepthActive func() bool
