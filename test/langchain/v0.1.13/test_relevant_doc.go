package main

import (
	"context"
	"github.com/alibaba/opentelemetry-go-auto-instrumentation/test/verifier"
	"github.com/tmc/langchaingo/vectorstores"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func main() {
	vec := vectorstores.ToRetriever(fakeVectorDb{}, 1)
	_, err := vec.GetRelevantDocuments(context.Background(), "123")
	if err != nil {
		panic(err)
	}
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyLLMAttributes(stubs[0][0], "relevantDocuments")
	}, 1)
}
