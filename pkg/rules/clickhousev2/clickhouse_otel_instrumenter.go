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

package clickhousev2

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type clickhouseAttrsGetter struct{}

func (g clickhouseAttrsGetter) GetSystem(_ clickhouseRequest) string {
	return "clickhouse"
}

func (g clickhouseAttrsGetter) GetServerAddress(clickhouseRequest clickhouseRequest) string {
	return clickhouseRequest.Addr
}

func (g clickhouseAttrsGetter) GetStatement(clickhouseRequest clickhouseRequest) string {
	return clickhouseRequest.Statement
}

func (g clickhouseAttrsGetter) GetCollection(_ clickhouseRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (g clickhouseAttrsGetter) GetOperation(clickhouseRequest clickhouseRequest) string {
	return clickhouseRequest.Op
}

func (g clickhouseAttrsGetter) GetParameters(clickhouseRequest clickhouseRequest) []any {
	return clickhouseRequest.Params
}

func (g clickhouseAttrsGetter) GetDbNamespace(clickhouseRequest clickhouseRequest) string {
	return clickhouseRequest.DbName
}

func (g clickhouseAttrsGetter) GetBatchSize(clickhouseRequest clickhouseRequest) int {
	return clickhouseRequest.BatchSize
}

func BuildClickhouseInstrumenter() instrumenter.Instrumenter[clickhouseRequest, interface{}] {
	builder := instrumenter.Builder[clickhouseRequest, interface{}]{}
	getter := clickhouseAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[clickhouseRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[clickhouseRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[clickhouseRequest, any, clickhouseAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[clickhouseRequest, any, clickhouseAttrsGetter]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("database.clickhouse")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.CLICKHOUSE_V2_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
