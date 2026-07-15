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

package gopg

import (
	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/db"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/instrumenter"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/utils"
	"github.com/alibaba/loongsuite-go/pkg/inst-api/version"
	"go.opentelemetry.io/otel/sdk/instrumentation"
)

type gogpAttrsGetter struct{}

func (g gogpAttrsGetter) GetSystem(_ gopgRequest) string {
	return "postgresql"
}

func (g gogpAttrsGetter) GetServerAddress(gopgRequest gopgRequest) string {
	return gopgRequest.Addr
}

func (g gogpAttrsGetter) GetStatement(gopgRequest gopgRequest) string {
	return gopgRequest.Statement
}

func (g gogpAttrsGetter) GetCollection(_ gopgRequest) string {
	// TBD: We need to implement retrieving the collection later.
	return ""
}

func (g gogpAttrsGetter) GetOperation(gopgRequest gopgRequest) string {
	return string(gopgRequest.QueryOp)
}

func (g gogpAttrsGetter) GetParameters(_ gopgRequest) []any {
	return nil
}

func (g gogpAttrsGetter) GetDbNamespace(gopgRequest gopgRequest) string {
	return gopgRequest.DbName
}

func (g gogpAttrsGetter) GetBatchSize(_ gopgRequest) int {
	return 0
}

func BuildGopgInstrumenter() instrumenter.Instrumenter[gopgRequest, interface{}] {
	builder := instrumenter.Builder[gopgRequest, interface{}]{}
	getter := gogpAttrsGetter{}
	return builder.Init().SetSpanNameExtractor(&db.DBSpanNameExtractor[gopgRequest]{Getter: getter}).SetSpanKindExtractor(&instrumenter.AlwaysClientExtractor[gopgRequest]{}).
		AddAttributesExtractor(&db.DbClientAttrsExtractor[gopgRequest, any, gogpAttrsGetter]{Base: db.DbClientCommonAttrsExtractor[gopgRequest, any, gogpAttrsGetter]{Getter: getter}}).
		AddOperationListeners(db.DbClientMetrics("database.gopg")).
		SetInstrumentationScope(instrumentation.Scope{
			Name:    utils.GOPG_SCOPE_NAME,
			Version: version.Tag,
		}).
		BuildInstrumenter()
}
