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

package elasticsearch

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type elasticSearchGetter struct {
}

func (e elasticSearchGetter) GetSystem(request *esRequest) string {
	return "elasticsearch"
}

func (e elasticSearchGetter) GetServerAddress(request *esRequest) string {
	return request.address
}

func (e elasticSearchGetter) GetStatement(request *esRequest) string {
	return request.request.URL.Path
}

func (e elasticSearchGetter) GetOperation(request *esRequest) string {
	return request.op
}

func (e elasticSearchGetter) GetCollection(request *esRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (e elasticSearchGetter) GetParameters(request *esRequest) []any {
	return request.params
}

func (e elasticSearchGetter) GetDbNamespace(request *esRequest) string {
	return ""
}

func (e elasticSearchGetter) GetBatchSize(request *esRequest) int {
	return 0
}

func BuildElasticSearchInstrumenter() instrumenter.Instrumenter[*esRequest, interface{}] {
	builder := instrumenter.Builder[*esRequest, any]{}
	getter := elasticSearchGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[*esRequest]{Getter: elasticSearchGetter{}}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*esRequest]{}).
		AddOperationListeners(db.DbClientMetrics("nosql.elasticsearch")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.ELASTICSEARCH_SCOPE_NAME,
			Version: version.Tag,
		}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[*esRequest, any, db.DbClientAttrsGetter[*esRequest]]{Base: db.DbClientCommonAttrsExtractor[*esRequest, any, db.DbClientAttrsGetter[*esRequest]]{Getter: getter}}).
		BuildInstrumenter()
}
