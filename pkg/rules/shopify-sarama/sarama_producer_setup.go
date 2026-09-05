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

package sarama

import (
	"context"
	"strconv"
	"sync"
	_ "unsafe"

	"github.com/Shopify/sarama"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const producerContextCacheMax = 16384

//go:linkname newAsyncProducerOnEnter github.com/Shopify/sarama.newAsyncProducerOnEnter
func newAsyncProducerOnEnter(call api.CallContext, client sarama.Client) {
	if !saramaEnabler.Enable() {
		return
	}
	if syncProducerCallDepthActive() {
		return
	}
	data := make(map[string]interface{}, 1)
	data["saramaConfig"] = client.Config()
	call.SetData(data)
}

//go:linkname newAsyncProducerOnExit github.com/Shopify/sarama.newAsyncProducerOnExit
func newAsyncProducerOnExit(call api.CallContext, p sarama.AsyncProducer, err error) {
	if !saramaEnabler.Enable() {
		return
	}
	if syncProducerCallDepthActive() {
		return
	}
	if err != nil || call.GetData() == nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	var saramaConfig *sarama.Config
	if data["saramaConfig"] == nil {
		saramaConfig = sarama.NewConfig()
	} else {
		saramaConfig = data["saramaConfig"].(*sarama.Config)
	}

	wrapped := &tracingAsyncProducer{
		AsyncProducer: p,
		input:         make(chan *sarama.ProducerMessage),
		successes:     make(chan *sarama.ProducerMessage),
		errors:        make(chan *sarama.ProducerError),
		closeErr:      make(chan error),
		closeSig:      make(chan struct{}),
		closeAsyncSig: make(chan struct{}),
	}

	var (
		mtx                     sync.Mutex
		producerMessageContexts = make(map[interface{}]saramaProducerMessageContext)
	)

	go func() {
		for {
			select {
			case <-wrapped.closeSig:
				wrapped.closeErr <- p.Close()
				return
			case <-wrapped.closeAsyncSig:
				p.AsyncClose()
				return
			case msg, ok := <-wrapped.input:
				if !ok {
					continue
				}

				request := saramaProducerRequest{
					msg:        msg,
					msgVersion: saramaConfig.Version,
				}
				ctx := context.Background()
				var haveTrace bool
				if msg.Headers != nil {
					for _, v := range msg.Headers {
						if string(v.Key) == "traceparent" {
							headerMap := make(propagation.MapCarrier)
							headerMap["traceparent"] = string(v.Value)
							ctx = otel.GetTextMapPropagator().Extract(ctx, headerMap)
							haveTrace = true
						}
					}
				}
				if haveTrace {
					ctx = producerInstrumenter.Start(ctx, request)
				} else {
					ctx = producerInstrumenter.Start(ctx, request, trace.WithNewRoot())
				}
				span := trace.SpanFromContext(ctx)

				mc := saramaProducerMessageContext{
					msg:            msg,
					traceCtx:       ctx,
					metadataBackup: msg.Metadata,
				}

				msg.Metadata = span.SpanContext().SpanID()
				if saramaConfig.Producer.Return.Successes ||
					len(producerMessageContexts) < producerContextCacheMax {
					mtx.Lock()
					producerMessageContexts[msg.Metadata] = mc
					mtx.Unlock()
				} else {
					producerInstrumenter.End(ctx, request, saramaProducerResponse{}, nil)
				}

				p.Input() <- msg
			}
		}
	}()

	var cleanupWg sync.WaitGroup

	cleanupWg.Add(1)
	go func() {
		defer func() {
			close(wrapped.successes)
			cleanupWg.Done()
		}()
		for msg := range p.Successes() {
			key := msg.Metadata
			mtx.Lock()
			if mc, ok := producerMessageContexts[key]; ok {
				delete(producerMessageContexts, key)
				partitionInt := int(msg.Partition)
				offsetStr := strconv.FormatInt(msg.Offset, 10)
				producerInstrumenter.End(mc.traceCtx, saramaProducerRequest{
					msg:        msg,
					msgVersion: saramaConfig.Version,
				}, saramaProducerResponse{
					partition: &partitionInt,
					offset:    &offsetStr,
				}, nil)
				msg.Metadata = mc.metadataBackup
			}
			mtx.Unlock()
			wrapped.successes <- msg
		}
	}()

	cleanupWg.Add(1)
	go func() {
		defer func() {
			close(wrapped.errors)
			cleanupWg.Done()
		}()
		for errMsg := range p.Errors() {
			key := errMsg.Msg.Metadata
			mtx.Lock()
			if mc, ok := producerMessageContexts[key]; ok {
				delete(producerMessageContexts, key)
				partitionInt := int(errMsg.Msg.Partition)
				offsetStr := strconv.FormatInt(errMsg.Msg.Offset, 10)
				producerInstrumenter.End(mc.traceCtx, saramaProducerRequest{
					msg:        errMsg.Msg,
					msgVersion: saramaConfig.Version,
				}, saramaProducerResponse{
					partition: &partitionInt,
					offset:    &offsetStr,
				}, errMsg.Err)
				errMsg.Msg.Metadata = mc.metadataBackup
			}
			mtx.Unlock()
			wrapped.errors <- errMsg
		}
	}()

	go func() {
		cleanupWg.Wait()
		mtx.Lock()
		for _, mc := range producerMessageContexts {
			producerInstrumenter.End(mc.traceCtx, saramaProducerRequest{
				msg:        mc.msg,
				msgVersion: saramaConfig.Version,
			}, saramaProducerResponse{}, nil)
		}
		mtx.Unlock()
	}()

	call.SetReturnVal(0, wrapped)
}

//go:linkname newSyncProducerOnEnter github.com/Shopify/sarama.newSyncProducerOnEnter
func newSyncProducerOnEnter(call api.CallContext, addrs []string, config *sarama.Config) {
	if !saramaEnabler.Enable() {
		return
	}
	syncProducerCallDepthEnter()
	data := make(map[string]interface{}, 1)
	data["saramaConfig"] = config
	call.SetData(data)
}

//go:linkname newSyncProducerOnExit github.com/Shopify/sarama.newSyncProducerOnExit
func newSyncProducerOnExit(call api.CallContext, producer sarama.SyncProducer, err error) {
	if call.GetData() == nil {
		return
	}
	defer syncProducerCallDepthExit()
	if !saramaEnabler.Enable() {
		return
	}
	if err != nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	var saramaConfig *sarama.Config
	if data["saramaConfig"] == nil {
		saramaConfig = sarama.NewConfig()
	} else {
		saramaConfig = data["saramaConfig"].(*sarama.Config)
	}

	wrapped := &tracingSyncProducer{
		SyncProducer: producer,
		saramaConfig: saramaConfig,
	}
	call.SetReturnVal(0, wrapped)
}

//go:linkname newSyncProducerFromClientOnEnter github.com/Shopify/sarama.newSyncProducerFromClientOnEnter
func newSyncProducerFromClientOnEnter(call api.CallContext, client sarama.Client) {
	if !saramaEnabler.Enable() {
		return
	}
	syncProducerCallDepthEnter()
	data := make(map[string]interface{}, 1)
	data["saramaConfig"] = client.Config()
	call.SetData(data)
}

//go:linkname newSyncProducerFromClientOnExit github.com/Shopify/sarama.newSyncProducerFromClientOnExit
func newSyncProducerFromClientOnExit(call api.CallContext, producer sarama.SyncProducer, err error) {
	if call.GetData() == nil {
		return
	}
	defer syncProducerCallDepthExit()
	if !saramaEnabler.Enable() {
		return
	}
	if err != nil {
		return
	}
	data, ok := call.GetData().(map[string]interface{})
	if !ok || data == nil {
		return
	}

	var saramaConfig *sarama.Config
	if data["saramaConfig"] == nil {
		saramaConfig = sarama.NewConfig()
	} else {
		saramaConfig = data["saramaConfig"].(*sarama.Config)
	}

	wrapped := &tracingSyncProducer{
		SyncProducer: producer,
		saramaConfig: saramaConfig,
	}
	call.SetReturnVal(0, wrapped)
}

func (p *tracingSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	request := saramaProducerRequest{
		msg:        msg,
		msgVersion: p.saramaConfig.Version,
	}
	ctx := producerInstrumenter.Start(context.Background(), request)

	partition, offset, err = p.SyncProducer.SendMessage(msg)

	partitionInt := int(partition)
	offsetStr := strconv.FormatInt(offset, 10)
	producerInstrumenter.End(ctx, request, saramaProducerResponse{
		partition: &partitionInt,
		offset:    &offsetStr,
	}, err)
	return partition, offset, err
}

func (p *tracingSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	producerMessageContexts := make(map[interface{}]saramaProducerMessageContext, len(msgs))

	for _, msg := range msgs {
		request := saramaProducerRequest{
			msg:        msg,
			msgVersion: p.saramaConfig.Version,
		}
		msgCtx := producerInstrumenter.Start(context.Background(), request)
		span := trace.SpanFromContext(msgCtx)

		mc := saramaProducerMessageContext{
			msg:            msg,
			traceCtx:       msgCtx,
			metadataBackup: msg.Metadata,
		}
		msg.Metadata = span.SpanContext().SpanID()
		producerMessageContexts[msg.Metadata] = mc
	}

	err := p.SyncProducer.SendMessages(msgs)

	if errors, ok := err.(sarama.ProducerErrors); ok {
		for _, singleErr := range errors {
			msg := singleErr.Msg
			key := msg.Metadata
			if mc, ok := producerMessageContexts[key]; ok {
				delete(producerMessageContexts, key)
				partitionInt := int(msg.Partition)
				offsetStr := strconv.FormatInt(msg.Offset, 10)
				producerInstrumenter.End(mc.traceCtx, saramaProducerRequest{
					msg:        msg,
					msgVersion: p.saramaConfig.Version,
				}, saramaProducerResponse{
					partition: &partitionInt,
					offset:    &offsetStr,
				}, singleErr.Err)
				msg.Metadata = mc.metadataBackup
			}
		}
	}

	for _, msg := range msgs {
		key := msg.Metadata
		if mc, ok := producerMessageContexts[key]; ok {
			delete(producerMessageContexts, key)
			partitionInt := int(msg.Partition)
			offsetStr := strconv.FormatInt(msg.Offset, 10)
			producerInstrumenter.End(mc.traceCtx, saramaProducerRequest{
				msg:        msg,
				msgVersion: p.saramaConfig.Version,
			}, saramaProducerResponse{
				partition: &partitionInt,
				offset:    &offsetStr,
			}, err)
			msg.Metadata = mc.metadataBackup
		}
	}

	return err
}
