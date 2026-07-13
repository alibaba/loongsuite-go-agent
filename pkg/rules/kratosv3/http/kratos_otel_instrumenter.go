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

package http

import (
	"context"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"

	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"go.opentelemetry.io/otel/attribute"
)

const kratos_protocol_type = "kratos.protocol.type"
const kratos_service_name = "kratos.service.name"
const kratos_service_id = "kratos.service.id"
const kratos_service_version = "kratos.service.version"
const kratos_service_meta = "kratos.service.meta"
const kratos_service_endpoint = "kratos.service.endpoint"

type kratosExperimentalAttributeExtractor struct {
}

func (k kratosExperimentalAttributeExtractor) OnStart(attributes []attribute.KeyValue, parentContext context.Context, request kratosRequest) ([]attribute.KeyValue, context.Context) {
	attributes = append(attributes, attribute.KeyValue{
		Key:   kratos_protocol_type,
		Value: attribute.StringValue(request.protocolType),
	}, attribute.KeyValue{
		Key:   kratos_service_name,
		Value: attribute.StringValue(request.serviceName),
	}, attribute.KeyValue{
		Key:   kratos_service_id,
		Value: attribute.StringValue(request.serviceId),
	}, attribute.KeyValue{
		Key:   kratos_service_version,
		Value: attribute.StringValue(request.serviceVersion),
	}, attribute.KeyValue{
		Key:   kratos_service_endpoint,
		Value: attribute.StringSliceValue(request.serviceEndpoint),
	})
	if request.serviceMeta != nil {
		for k, v := range request.serviceMeta {
			attributes = append(attributes, attribute.KeyValue{
				Key:   attribute.Key(kratos_service_meta + "." + k),
				Value: attribute.StringValue(v),
			})
		}
	}
	return attributes, parentContext
}

func (k kratosExperimentalAttributeExtractor) OnEnd(attributes []attribute.KeyValue, context context.Context, request kratosRequest, response any, err error) ([]attribute.KeyValue, context.Context) {
	return attributes, context
}

type kratosExperimentalSpanNameExtractor struct {
}

func (k kratosExperimentalSpanNameExtractor) Extract(request kratosRequest) string {
	if request.protocolType == "grpc" {
		return "kratos.grpc." + request.serviceName
	}
	if request.protocolType == "http" {
		return "kratos.http." + request.serviceName
	}
	return "kratos.unknown"
}

func BuildKratosInternalInstrumenter() instrumenter.Instrumenter[kratosRequest, any] {
	builder := instrumenter.Builder[kratosRequest, any]{}
	return builder.Init().SetSpanNameExtractor(&kratosExperimentalSpanNameExtractor{}).
		SetSpanKindExtractor(&instrumenter.AlwaysInternalExtractor[kratosRequest]{}).
		AddAttributesExtractor(&kratosExperimentalAttributeExtractor{}).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.KRATOS_HTTP_INTERNAL_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
