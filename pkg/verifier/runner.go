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

package verifier

import (
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"log"
	"sort"
	"time"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type node struct {
	root       bool
	childNodes []*node
	span       tracetest.SpanStub
}

func WaitAndAssertTraces(traceVerifiers func([]tracetest.SpanStubs), numTraces int) {
	traces := waitForTraces(numTraces)
	traceVerifiers(traces)
}

func WaitAndAssertMetrics(metricName string, metricVerifiers ...func(metricdata.ResourceMetrics)) {
	mrs, err := waitForMetrics(metricName)
	if err != nil {
		log.Fatalf("Failed to wait for metric %s: %v", metricName, err)
	}
	for _, v := range metricVerifiers {
		v(mrs)
	}
}

func waitForMetrics(metricName string) (metricdata.ResourceMetrics, error) {
	// 最多等30s
	var (
		mrs metricdata.ResourceMetrics
		err error
	)
	finish := false
	for !finish {
		select {
		case <-time.After(30 * time.Second):
			finish = true
		default:
			mrs, err = filterMetricByName(metricName)
			if err == nil {
				finish = true
				break
			}
		}
	}
	return mrs, err
}

func filterMetricByName(name string) (metricdata.ResourceMetrics, error) {
	data := GetTestMetrics()
	for i, s := range data.ScopeMetrics {
		scms := make([]metricdata.Metrics, 0)
		for _, sm := range s.Metrics {
			if sm.Name == name {
				scms = append(scms, sm)
			}
		}
		data.ScopeMetrics[i].Metrics = scms
	}
	return data, nil
}

func waitForTraces(numberOfTraces int) []tracetest.SpanStubs {
	defer ResetTestSpans()
	// 最多等20s
	finish := false
	var traces []tracetest.SpanStubs
	var i int
	for !finish {
		select {
		case <-time.After(20 * time.Second):
			log.Printf("Timeout waiting for traces!")
			finish = true
		default:
			traces = groupAndSortTrace()
			if len(traces) >= numberOfTraces {
				finish = true
			}
			i++
		}
		if i == 10 {
			break
		}
	}
	return traces
}

func groupAndSortTrace() []tracetest.SpanStubs {
	spans := GetTestSpans()
	traceMap := make(map[string][]tracetest.SpanStub)
	for _, span := range *spans {
		if span.SpanContext.HasTraceID() && span.SpanContext.TraceID().IsValid() {
			traceId := span.SpanContext.TraceID().String()
			spans, ok := traceMap[traceId]
			if !ok {
				spans = make([]tracetest.SpanStub, 0)
			}
			spans = append(spans, span)
			traceMap[traceId] = spans
		}
	}
	return sortTrace(traceMap)
}

func sortTrace(traceMap map[string][]tracetest.SpanStub) []tracetest.SpanStubs {
	traces := make([][]tracetest.SpanStub, 0)
	for _, trace := range traceMap {
		traces = append(traces, trace)
	}
	// 按开始时间从小到大排
	sort.Slice(traces, func(i, j int) bool {
		return traces[i][0].StartTime.UnixNano() < traces[j][0].StartTime.UnixNano()
	})
	for i, _ := range traces {
		traces[i] = sortSingleTrace(traces[i])
	}
	stubs := make([]tracetest.SpanStubs, 0)
	for i, _ := range traces {
		stubs = append(stubs, traces[i])
	}
	return stubs
}

func sortSingleTrace(stubs []tracetest.SpanStub) []tracetest.SpanStub {
	// 同一条trace的按span的父子关系排
	lookup := make(map[string]*node)
	for _, stub := range stubs {
		lookup[stub.SpanContext.SpanID().String()] = &node{
			root:       true,
			childNodes: make([]*node, 0),
			span:       stub,
		}
	}
	for _, stub := range stubs {
		n, ok := lookup[stub.SpanContext.SpanID().String()]
		if !ok {
			panic("no span id in stub " + stub.Name)
		}
		// 发现了父节点，就添加到父节点的子节点列表里面去
		if n.span.Parent.SpanID().IsValid() {
			parentSpanId := n.span.Parent.SpanID().String()
			parentNode, ok := lookup[parentSpanId]
			if ok {
				parentNode.childNodes = append(parentNode.childNodes, n)
				n.root = false
			}
		}
	}
	// 寻找根节点
	rootNodes := make([]*node, 0)
	for _, stub := range stubs {
		n, ok := lookup[stub.SpanContext.SpanID().String()]
		if !ok {
			panic("no span id in stub " + stub.Name)
		}
		sort.Slice(n.childNodes, func(i, j int) bool {
			return n.childNodes[i].span.StartTime.UnixNano() < n.childNodes[j].span.StartTime.UnixNano()
		})
		if n.root {
			rootNodes = append(rootNodes, n)
		}
	}
	sort.Slice(rootNodes, func(i, j int) bool {
		return rootNodes[i].span.StartTime.UnixNano() < rootNodes[j].span.StartTime.UnixNano()
	})
	// 层序遍历，获取排序后的span
	t := make([]tracetest.SpanStub, 0)
	for _, rootNode := range rootNodes {
		traversePreOrder(rootNode, &t)
	}
	return t
}

func traversePreOrder(n *node, acc *[]tracetest.SpanStub) {
	*acc = append(*acc, n.span)
	for _, child := range n.childNodes {
		traversePreOrder(child, acc)
	}
}
