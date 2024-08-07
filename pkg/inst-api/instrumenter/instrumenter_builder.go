package instrumenter

import (
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/pkg/core/meter"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// TODO: add route updater here, now we do not support such controller layer to update route.
type OperationMetrics interface {
	Create(meter metric.Meter) *OperationListenerWrapper
	Match(meta meter.MeterMeta) bool
}

type InstrumentEnabler interface {
	IsEnabled() bool
}

type defaultInstrumentEnabler struct {
}

func NewDefaultInstrumentEnabler() InstrumentEnabler {
	return &defaultInstrumentEnabler{}
}

func (a *defaultInstrumentEnabler) IsEnabled() bool {
	return true
}

type Builder[REQUEST any, RESPONSE any] struct {
	SpanNameExtractor    SpanNameExtractor[REQUEST]
	SpanKindExtractor    SpanKindExtractor[REQUEST]
	SpanStatusExtractor  SpanStatusExtractor[REQUEST, RESPONSE]
	AttributesExtractors []AttributesExtractor[REQUEST, RESPONSE]
	OperationListeners   []*OperationListenerWrapper
	OperationMetrics     []OperationMetrics
	ContextCustomizers   []ContextCustomizer[REQUEST]
	SpanSuppressor       SpanSuppressor
	Tracer               trace.Tracer
	InstVersion          string
}

func (b *Builder[REQUEST, RESPONSE]) Init() *Builder[REQUEST, RESPONSE] {
	b.AttributesExtractors = make([]AttributesExtractor[REQUEST, RESPONSE], 0)
	b.OperationListeners = b.buildOperationListeners()
	b.ContextCustomizers = make([]ContextCustomizer[REQUEST], 0)
	b.SpanSuppressor = b.buildSpanSuppressor()
	b.SpanStatusExtractor = &defaultSpanStatusExtractor[REQUEST, RESPONSE]{}
	b.Tracer = otel.GetTracerProvider().Tracer("")
	return b
}

func (b *Builder[REQUEST, RESPONSE]) SetInstVersion(instVersion string) *Builder[REQUEST, RESPONSE] {
	b.InstVersion = instVersion
	return b
}

func (b *Builder[REQUEST, RESPONSE]) SetSpanNameExtractor(spanNameExtractor SpanNameExtractor[REQUEST]) *Builder[REQUEST, RESPONSE] {
	b.SpanNameExtractor = spanNameExtractor
	return b
}

func (b *Builder[REQUEST, RESPONSE]) SetSpanStatusExtractor(spanStatusExtractor SpanStatusExtractor[REQUEST, RESPONSE]) *Builder[REQUEST, RESPONSE] {
	b.SpanStatusExtractor = spanStatusExtractor
	return b
}

func (b *Builder[REQUEST, RESPONSE]) SetSpanKindExtractor(spanKindExtractor SpanKindExtractor[REQUEST]) *Builder[REQUEST, RESPONSE] {
	b.SpanKindExtractor = spanKindExtractor
	return b
}

func (b *Builder[REQUEST, RESPONSE]) AddAttributesExtractor(attributesExtractor ...AttributesExtractor[REQUEST, RESPONSE]) *Builder[REQUEST, RESPONSE] {
	b.AttributesExtractors = append(b.AttributesExtractors, attributesExtractor...)
	return b
}

func (b *Builder[REQUEST, RESPONSE]) AddOperationListeners(operationListener ...*OperationListenerWrapper) *Builder[REQUEST, RESPONSE] {
	b.OperationListeners = append(b.OperationListeners, operationListener...)
	return b
}

func (b *Builder[REQUEST, RESPONSE]) AddContextCustomizers(contextCustomizers ...ContextCustomizer[REQUEST]) *Builder[REQUEST, RESPONSE] {
	b.ContextCustomizers = append(b.ContextCustomizers, contextCustomizers...)
	return b
}

func (b *Builder[REQUEST, RESPONSE]) BuildInstrumenter() *Instrumenter[REQUEST, RESPONSE] {
	return &Instrumenter[REQUEST, RESPONSE]{
		spanNameExtractor:    b.SpanNameExtractor,
		spanKindExtractor:    b.SpanKindExtractor,
		spanStatusExtractor:  b.SpanStatusExtractor,
		attributesExtractors: b.AttributesExtractors,
		operationListeners:   b.OperationListeners,
		operationMetrics:     b.OperationMetrics,
		contextCustomizers:   b.ContextCustomizers,
		spanSuppressor:       b.SpanSuppressor,
		tracer:               b.Tracer,
		instVersion:          b.InstVersion,
	}
}

func (b *Builder[REQUEST, RESPONSE]) buildOperationListeners() []*OperationListenerWrapper {
	if len(b.OperationMetrics) == 0 {
		return make([]*OperationListenerWrapper, 0)
	}
	meterProvider := meter.GetMeterProvider()
	if meterProvider == nil {
		return make([]*OperationListenerWrapper, 0)
	}

	listeners := make([]*OperationListenerWrapper, 0, len(b.OperationMetrics)+len(b.OperationListeners))

	meters := meterProvider.GetMeters()
	for _, m := range meters {
		for _, factory := range b.OperationMetrics {
			if factory.Match(m.Metas()) {
				listeners = append(listeners, factory.Create(m.Meter()))
			}
		}
	}
	return listeners
}

// TODO: create suppressor by otel.instrumentation.experimental.span-suppression-strategy
func (b *Builder[REQUEST, RESPONSE]) buildSpanSuppressor() SpanSuppressor {
	kvs := make(map[attribute.Key]bool)
	for _, extractor := range b.AttributesExtractors {
		provider, ok := extractor.(SpanKeyProvider)
		if ok {
			kvs[provider.GetSpanKey()] = true
		}
	}
	kSlice := make([]attribute.Key, 0, len(kvs))
	for k := range kvs {
		kSlice = append(kSlice, k)
	}
	return NewSpanKeySuppressor(kSlice)
}
