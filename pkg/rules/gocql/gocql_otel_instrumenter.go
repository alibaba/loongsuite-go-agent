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

package gocql

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type gogpAttrsGetter struct{}

func (g gogpAttrsGetter) GetSystem(_ gocqlRequest) string {
	return "cassandra"
}

func (g gogpAttrsGetter) GetServerAddress(gocqlRequest gocqlRequest) string {
	return gocqlRequest.Addr
}

func (g gogpAttrsGetter) GetStatement(gocqlRequest gocqlRequest) string {
	return gocqlRequest.Statement
}

func (g gogpAttrsGetter) GetCollection(_ gocqlRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (g gogpAttrsGetter) GetOperation(gocqlRequest gocqlRequest) string {
	return gocqlRequest.Op
}

func (g gogpAttrsGetter) GetParameters(_ gocqlRequest) []any {
	return nil
}

func (g gogpAttrsGetter) GetDbNamespace(gocqlRequest gocqlRequest) string {
	return gocqlRequest.DbName
}

func (g gogpAttrsGetter) GetBatchSize(gocqlRequest gocqlRequest) int {
	return gocqlRequest.BatchSize
}

func BuildGocqlInstrumenter() instrumenter.Instrumenter[gocqlRequest, interface{}] {
	builder := instrumenter.Builder[gocqlRequest, interface{}]{}
	getter := gogpAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[gocqlRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[gocqlRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[gocqlRequest, any, gogpAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[gocqlRequest, any, gogpAttrsGetter]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("nosql.gocql")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GOCQL_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
