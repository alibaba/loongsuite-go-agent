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
	"fmt"
	"os"
	"strconv"

	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api-semconv/instrumenter/message"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go-agent/pkg/inst-api/version"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
	"go.opentelemetry.io/otel/trace"
)

// Instrumentation control
var (
	rocketmqEnabler           = rocketmqInnerEnabler{os.Getenv("OTEL_INSTRUMENTATION_ROCKETMQ_ENABLED") != "false"}
	producerInst              = newProducerInstrumenter()
	singleProcessConsumerInst = newConsumerInstrumenter(message.PROCESS, false)
	batchProcessConsumerInst  = newConsumerInstrumenter(message.PROCESS, true)
	receiveConsumerInst       = newReceiveInstrumenter()
)

type rocketmqInnerEnabler struct {
	enabled bool
}

func (g rocketmqInnerEnabler) Enable() bool {
	return g.enabled
}

// sendStatusToString converts SendStatus to readable string
func sendStatusToString(status primitive.SendStatus) string {
	switch status {
	case primitive.SendOK:
		return "SEND_OK"
	case primitive.SendFlushDiskTimeout:
		return "SEND_FLUSH_DISK_TIMEOUT"
	case primitive.SendFlushSlaveTimeout:
		return "SEND_FLUSH_SLAVE_TIMEOUT"
	case primitive.SendSlaveNotAvailable:
		return "SEND_SLAVE_NOT_AVAILABLE"
	default:
		return "UNKNOWN"
	}
}

// ProducerCarrier implements propagation.TextMapCarrier for producer messages
type ProducerCarrier struct {
	Msg *primitive.Message
}

func (c ProducerCarrier) Get(key string) string {
	return c.Msg.GetProperty(key)
}

func (c ProducerCarrier) Set(key, value string) {
	c.Msg.WithProperty(key, value)
}

func (c ProducerCarrier) Keys() []string {
	props := c.Msg.GetProperties()
	if props == nil {
		return nil
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	return keys
}

// ConsumerCarrier implements propagation.TextMapCarrier for consumer messages
type ConsumerCarrier struct {
	Msg *primitive.MessageExt
}

func (c ConsumerCarrier) Get(key string) string {
	return c.Msg.GetProperty(key)
}

func (c ConsumerCarrier) Set(key, value string) {
	c.Msg.WithProperty(key, value)
}

func (c ConsumerCarrier) Keys() []string {
	props := c.Msg.GetProperties()
	if props == nil {
		return nil
	}
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	return keys
}

// ProducerStatusExtractor extracts status for producer spans
type ProducerStatusExtractor struct{}

func (e *ProducerStatusExtractor) Extract(span trace.Span, req ProducerRequest, res ProducerResponse, err error) {
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// ConsumerStatusExtractor extracts status for consumer spans
type ConsumerStatusExtractor struct{}

func (e *ConsumerStatusExtractor) Extract(span trace.Span, req ConsumerRequest, res ConsumerResponse, err error) {
	switch {
	case err != nil:
		span.SetStatus(codes.Error, err.Error())
	case res.Result != consumer.ConsumeSuccess:
		span.SetStatus(codes.Error, fmt.Sprintf("consume result: %v", res.Result))
	default:
		span.SetStatus(codes.Ok, "")
	}
}

// ProducerAttrsGetter implements message attribute getter for producers
type ProducerAttrsGetter struct{}

func (r ProducerAttrsGetter) GetDestination(req ProducerRequest) string {
	if req.Message == nil {
		return ""
	}
	return req.Message.Topic
}

func (r ProducerAttrsGetter) GetMessageBodySize(req ProducerRequest) int64 {
	if req.Message == nil {
		return 0
	}
	return int64(len(req.Message.Body))
}

func (r ProducerAttrsGetter) IsAnonymousDestination(req ProducerRequest) bool {
	return false
}

func (r ProducerAttrsGetter) GetDestinationPartitionId(req ProducerRequest) string {
	return ""
}

func (r ProducerAttrsGetter) GetSystem(req ProducerRequest) string {
	return "rocketmq"
}

func (r ProducerAttrsGetter) GetDestinationTemplate(req ProducerRequest) string {
	return ""
}

func (r ProducerAttrsGetter) IsTemporaryDestination(req ProducerRequest) bool {
	return false
}

func (r ProducerAttrsGetter) isAnonymousDestination(req ProducerRequest) bool {
	return false
}

func (r ProducerAttrsGetter) GetConversationId(req ProducerRequest) string {
	return ""
}

func (r ProducerAttrsGetter) GetMessageEnvelopSize(req ProducerRequest) int64 {
	return 0
}

func (r ProducerAttrsGetter) GetMessageId(req ProducerRequest, res ProducerResponse) string {
	if res.Result != nil {
		return res.Result.MsgID
	}
	return ""
}

func (r ProducerAttrsGetter) GetClientId(req ProducerRequest) string {
	return ""
}

func (r ProducerAttrsGetter) GetBatchMessageCount(req ProducerRequest, res ProducerResponse) int64 {
	return 1
}

func (r ProducerAttrsGetter) GetMessageHeader(req ProducerRequest, name string) []string {
	if req.Message == nil {
		return nil
	}
	value := req.Message.GetProperty(name)
	if value == "" {
		return nil
	}
	return []string{value}
}

// ConsumerAttrsGetter implements message attribute getter for consumers
type ConsumerAttrsGetter struct{}

func (r ConsumerAttrsGetter) GetDestination(req ConsumerRequest) string {
	if req.Message == nil {
		return ""
	}
	return req.Message.Topic
}

func (r ConsumerAttrsGetter) GetMessageBodySize(req ConsumerRequest) int64 {
	if req.Message == nil {
		return 0
	}
	return int64(len(req.Message.Body))
}

func (r ConsumerAttrsGetter) IsAnonymousDestination(req ConsumerRequest) bool {
	return false
}

func (r ConsumerAttrsGetter) GetDestinationPartitionId(req ConsumerRequest) string {
	return ""
}

func (r ConsumerAttrsGetter) GetSystem(req ConsumerRequest) string {
	return "rocketmq"
}

func (r ConsumerAttrsGetter) GetDestinationTemplate(req ConsumerRequest) string {
	return ""
}

func (r ConsumerAttrsGetter) IsTemporaryDestination(req ConsumerRequest) bool {
	return false
}

func (r ConsumerAttrsGetter) isAnonymousDestination(req ConsumerRequest) bool {
	return false
}

func (r ConsumerAttrsGetter) GetConversationId(req ConsumerRequest) string {
	return ""
}

func (r ConsumerAttrsGetter) GetMessageEnvelopSize(req ConsumerRequest) int64 {
	return 0
}

func (r ConsumerAttrsGetter) GetMessageId(req ConsumerRequest, response ConsumerResponse) string {
	if req.Message == nil {
		return ""
	}
	return req.Message.MsgId
}

func (r ConsumerAttrsGetter) GetClientId(req ConsumerRequest) string {
	return ""
}

func (r ConsumerAttrsGetter) GetBatchMessageCount(req ConsumerRequest, res ConsumerResponse) int64 {
	return 1
}

func (r ConsumerAttrsGetter) GetMessageHeader(req ConsumerRequest, name string) []string {
	if req.Message == nil {
		return nil
	}
	value := req.Message.GetProperty(name)
	if value == "" {
		return nil
	}
	return []string{value}
}

// ConsumerProcessAttrsExtractor extracts additional attributes for consumer spans
type ConsumerProcessAttrsExtractor struct{}

func (e *ConsumerProcessAttrsExtractor) OnStart(attrs []attribute.KeyValue, ctx context.Context, req ConsumerRequest) ([]attribute.KeyValue, context.Context) {
	if req.Message == nil {
		return attrs, ctx
	}

	return append(attrs,
		semconv.MessagingRocketmqMessageTagKey.String(req.Message.GetTags()),
		semconv.MessagingRocketmqMessageKeysKey.String(req.Message.GetKeys()),
	), ctx
}

func (e *ConsumerProcessAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, req ConsumerRequest, res ConsumerResponse, err error) ([]attribute.KeyValue, context.Context) {
	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
	}

	if req.BrokerAddr != "" {
		attrs = append(attrs, attribute.String("messaging.rocketmq.broker_address", req.BrokerAddr))
	}

	if req.Message == nil {
		return attrs, ctx
	}

	attrs = append(attrs,
		attribute.String("messaging.rocketmq.queue_offset", strconv.FormatInt(req.Message.QueueOffset, 10)),
	)
	if req.Message.TransactionId != "" {
		attrs = append(attrs, attribute.String("messaging.rocketmq.transaction_id", req.Message.TransactionId))
	}
	if req.Message.Queue != nil {
		attrs = append(attrs,
			attribute.Int("messaging.rocketmq.queue_id", req.Message.Queue.QueueId),
			attribute.String("messaging.rocketmq.broker_name", req.Message.Queue.BrokerName),
		)
	}

	return attrs, ctx
}

// ProducerAttrsExtractor extracts additional attributes for producer spans
type ProducerAttrsExtractor struct{}

func (m *ProducerAttrsExtractor) OnStart(attrs []attribute.KeyValue, ctx context.Context, req ProducerRequest) ([]attribute.KeyValue, context.Context) {
	if req.Message == nil {
		return attrs, ctx
	}

	return append(attrs,
		semconv.MessagingSystemRocketmq,
		semconv.MessagingDestinationNameKey.String(req.Message.Topic),
		semconv.MessagingRocketmqMessageTagKey.String(req.Message.GetTags()),
		semconv.MessagingRocketmqMessageKeysKey.String(req.Message.GetKeys()),
	), ctx
}

func (e *ProducerAttrsExtractor) OnEnd(attrs []attribute.KeyValue, ctx context.Context, req ProducerRequest, res ProducerResponse, err error) ([]attribute.KeyValue, context.Context) {
	if err != nil {
		attrs = append(attrs, attribute.String("error", err.Error()))
	}
	if res.BrokerAddr != "" {
		attrs = append(attrs, attribute.String("messaging.rocketmq.broker_address", res.BrokerAddr))
	}

	if res.Result != nil {
		attrs = append(attrs,
			attribute.String("messaging.rocketmq.queue_offset", strconv.FormatInt(res.Result.QueueOffset, 10)),
			attribute.String("messaging.rocketmq.send_result", sendStatusToString(res.Result.Status)),
		)
		if res.Result.TransactionID != "" {
			attrs = append(attrs, attribute.String("messaging.rocketmq.transaction_id", res.Result.TransactionID))
		}
		if res.Result.MessageQueue != nil {
			attrs = append(attrs,
				attribute.Int("messaging.rocketmq.queue_id", res.Result.MessageQueue.QueueId),
				attribute.String("messaging.rocketmq.broker_name", res.Result.MessageQueue.BrokerName),
			)
		}
	}
	return attrs, ctx
}

// ConsumerSpanNameExtractor extracts span names for consumers
type ConsumerSpanNameExtractor struct{}

func (e *ConsumerSpanNameExtractor) Extract(req any) string {
	return "multiple_sources receive"
}

// newProducerInstrumenter constructs the producer instrumentation
func newProducerInstrumenter() instrumenter.Instrumenter[ProducerRequest, ProducerResponse] {
	builder := instrumenter.Builder[ProducerRequest, ProducerResponse]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ROCKETMQGO_PRODUCER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[ProducerRequest, ProducerResponse]{Getter: ProducerAttrsGetter{}, OperationName: message.PUBLISH}).
		SetSpanKindExtractor(&instrumenter.AlwaysProducerExtractor[ProducerRequest]{}).
		AddAttributesExtractor(&ProducerAttrsExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[ProducerRequest, ProducerResponse, ProducerAttrsGetter]{Operation: message.PUBLISH}).
		SetSpanStatusExtractor(&ProducerStatusExtractor{}).
		BuildPropagatingToDownstreamInstrumenter(
			func(req ProducerRequest) propagation.TextMapCarrier {
				return ProducerCarrier{Msg: req.Message}
			},
			otel.GetTextMapPropagator(),
		)
}

// newReceiveInstrumenter constructs the receive instrumentation
func newReceiveInstrumenter() instrumenter.Instrumenter[any, any] {
	builder := instrumenter.Builder[any, any]{}
	return builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ROCKETMQGO_CONSUMER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&ConsumerSpanNameExtractor{}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[any]{}).
		BuildInstrumenter()
}

// newConsumerInstrumenter constructs consumer instrumentation
func newConsumerInstrumenter(operation message.MessageOperation, isBatch bool) instrumenter.Instrumenter[ConsumerRequest, ConsumerResponse] {
	builder := instrumenter.Builder[ConsumerRequest, ConsumerResponse]{}
	buildInstrumenter := builder.Init().
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ROCKETMQGO_CONSUMER_SCOPE_NAME,
			Version: version.Tag,
		}).
		SetSpanNameExtractor(&message.MessageSpanNameExtractor[ConsumerRequest, ConsumerResponse]{
			Getter:        ConsumerAttrsGetter{},
			OperationName: operation,
		}).
		SetSpanKindExtractor(&instrumenter.AlwaysConsumerExtractor[ConsumerRequest]{}).
		AddAttributesExtractor(&ConsumerProcessAttrsExtractor{}).
		AddAttributesExtractor(&message.MessageAttrsExtractor[ConsumerRequest, ConsumerResponse, ConsumerAttrsGetter]{
			Operation: operation,
		})

	if !isBatch {
		return buildInstrumenter.SetSpanStatusExtractor(&ConsumerStatusExtractor{}).
			BuildPropagatingFromUpstreamInstrumenter(
				func(req ConsumerRequest) propagation.TextMapCarrier {
					return ConsumerCarrier{Msg: req.Message}
				},
				otel.GetTextMapPropagator(),
			)
	} else {
		return buildInstrumenter.BuildInstrumenter()
	}
}

// Start begins instrumentation for message processing
func Start(parentContext context.Context, msgs []*primitive.MessageExt, addr string) context.Context {
	if len(msgs) == 1 {
		msg := msgs[0]
		return singleProcessConsumerInst.Start(parentContext, ConsumerRequest{
			BrokerAddr: addr,
			Message:    msg,
		})
	}

	rootContext := receiveConsumerInst.Start(parentContext, nil, trace.WithAttributes(
		semconv.MessagingSystemRocketmq,
		attribute.String(string(semconv.MessagingOperationTypeKey), string(message.RECEIVE)),
	))

	for _, msg := range msgs {
		createChildSpan(rootContext, msg, addr)
	}
	return rootContext
}

func createChildSpan(ctx context.Context, msg *primitive.MessageExt, addr string) {
	childCtx := batchProcessConsumerInst.Start(ctx,
		ConsumerRequest{
			BrokerAddr: addr,
			Message:    msg,
		},
		trace.WithLinks(trace.Link{
			SpanContext: trace.SpanContextFromContext(
				otel.GetTextMapPropagator().Extract(ctx, ConsumerCarrier{Msg: msg})),
		}),
	)
	batchProcessConsumerInst.End(childCtx,
		ConsumerRequest{Message: msg},
		ConsumerResponse{},
		nil,
	)
}

// End completes instrumentation for message processing
func End(ctx context.Context, msgs []*primitive.MessageExt, result consumer.ConsumeResult, err error) {
	if len(msgs) == 1 {
		msg := msgs[0]
		singleProcessConsumerInst.End(ctx,
			ConsumerRequest{Message: msg},
			ConsumerResponse{Result: result},
			err,
		)
		return
	}
	receiveConsumerInst.End(ctx, msgs, result, err)
}
