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

package rueidisgo

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"unicode/utf8"
)

type goRueidisAttrsGetter struct {
}

func (d goRueidisAttrsGetter) GetSystem(request *goRueidisRequest) string {
	return "redis"
}

func (d goRueidisAttrsGetter) GetServerAddress(request *goRueidisRequest) string {
	return request.endpoint
}

func (d goRueidisAttrsGetter) GetStatement(request *goRueidisRequest) string {
	if request.cmd.statement == "" {
		return request.cmd.cmdName
	}
	if utf8.ValidString(request.cmd.statement) {
		return "not_support_type"
	} else {
		return request.cmd.statement
	}
}

func (d goRueidisAttrsGetter) GetOperation(request *goRueidisRequest) string {
	return request.cmd.cmdName
}

func (d goRueidisAttrsGetter) GetParameters(request *goRueidisRequest) []any {
	return nil
}

func (d goRueidisAttrsGetter) GetDbNamespace(request *goRueidisRequest) string {
	return ""
}

func (d goRueidisAttrsGetter) GetBatchSize(request *goRueidisRequest) int {
	return 0
}

func (d goRueidisAttrsGetter) GetCollection(request *goRueidisRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func BuildGoRueidisOtelInstrumenter() instrumenter.Instrumenter[*goRueidisRequest, any] {
	builder := instrumenter.Builder[*goRueidisRequest, any]{}
	getter := goRueidisAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[*goRueidisRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[*goRueidisRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[*goRueidisRequest, any, db.DbClientAttrsGetter[*goRueidisRequest]]{Base: db.DbClientCommonAttrsExtractor[*goRueidisRequest, any, db.DbClientAttrsGetter[*goRueidisRequest]]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("nosql.rueidis")).
		BuildInstrumenter()
}
