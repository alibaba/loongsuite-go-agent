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

package neo4j

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type neo4jAttrsGetter struct{}

func (g neo4jAttrsGetter) GetSystem(_ neo4jRequest) string {
	return "neo4j"
}

func (g neo4jAttrsGetter) GetServerAddress(_ neo4jRequest) string {
	return ""
}

func (g neo4jAttrsGetter) GetStatement(request neo4jRequest) string {
	return request.Statement
}

func (g neo4jAttrsGetter) GetCollection(_ neo4jRequest) string {
	return ""
}

func (g neo4jAttrsGetter) GetOperation(request neo4jRequest) string {
	return request.Op
}

func (g neo4jAttrsGetter) GetParameters(_ neo4jRequest) []any {
	return nil
}

func (g neo4jAttrsGetter) GetDbNamespace(_ neo4jRequest) string {
	return ""
}

func (g neo4jAttrsGetter) GetBatchSize(_ neo4jRequest) int {
	return 1
}

func BuildNeo4jInstrumenter() instrumenter.Instrumenter[neo4jRequest, interface{}] {
	builder := instrumenter.Builder[neo4jRequest, interface{}]{}
	getter := neo4jAttrsGetter{}
	return builder.Init().
		SetSpanNameExtractor(&db.DBSpanNameExtractor[neo4jRequest]{Getter: getter}).
		SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[neo4jRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[neo4jRequest, any, neo4jAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[neo4jRequest, any, neo4jAttrsGetter]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("neo4j")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.NEO4J_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
