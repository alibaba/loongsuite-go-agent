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

package utils

import "go.opentelemetry.io/otel/attribute"

// InstrumentationCategory represents the semantic category of an instrumentation
type InstrumentationCategory string

const (
	CategoryHTTP      InstrumentationCategory = "http"
	CategoryRPC       InstrumentationCategory = "rpc"
	CategoryDB        InstrumentationCategory = "db"
	CategoryMessaging InstrumentationCategory = "messaging"
	CategoryAI        InstrumentationCategory = "ai"
	CategoryOther     InstrumentationCategory = "other"
)

// InstrumentationMetadata contains metadata for each instrumentation scope
type InstrumentationMetadata struct {
	ScopeName string
	Category  InstrumentationCategory
	ClientKey attribute.Key
	ServerKey attribute.Key
}

// InstrumentationRegistry maps scope names to their metadata
var InstrumentationRegistry = map[string]*InstrumentationMetadata{
	// HTTP
	"loongsuite.instrumentation.fasthttp": {
		ScopeName: "loongsuite.instrumentation.fasthttp",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.nethttp": {
		ScopeName: "loongsuite.instrumentation.nethttp",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.hertz": {
		ScopeName: "loongsuite.instrumentation.hertz",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.fiber": {
		ScopeName: "loongsuite.instrumentation.fiber",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.elasticsearch": {
		ScopeName: "loongsuite.instrumentation.elasticsearch",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.kratos": {
		ScopeName: "loongsuite.instrumentation.kratos",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},
	"loongsuite.instrumentation.k8s-client-go": {
		ScopeName: "loongsuite.instrumentation.k8s-client-go",
		Category:  CategoryHTTP,
		ClientKey: HTTP_CLIENT_KEY,
		ServerKey: HTTP_SERVER_KEY,
	},

	// RPC
	"loongsuite.instrumentation.grpc": {
		ScopeName: "loongsuite.instrumentation.grpc",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},
	"loongsuite.instrumentation.trpc": {
		ScopeName: "loongsuite.instrumentation.trpc",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},
	"loongsuite.instrumentation.kitex": {
		ScopeName: "loongsuite.instrumentation.kitex",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},
	"loongsuite.instrumentation.dubbo": {
		ScopeName: "loongsuite.instrumentation.dubbo",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},
	"loongsuite.instrumentation.gomicro": {
		ScopeName: "loongsuite.instrumentation.gomicro",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},
	"loongsuite.instrumentation.mcp": {
		ScopeName: "loongsuite.instrumentation.mcp",
		Category:  CategoryRPC,
		ClientKey: RPC_CLIENT_KEY,
		ServerKey: RPC_SERVER_KEY,
	},

	// Database
	"loongsuite.instrumentation.databasesql": {
		ScopeName: "loongsuite.instrumentation.databasesql",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "", // DB only has client
	},
	"loongsuite.instrumentation.goredisv9": {
		ScopeName: "loongsuite.instrumentation.goredisv9",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.goredisv8": {
		ScopeName: "loongsuite.instrumentation.goredisv8",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.gomysql": {
		ScopeName: "loongsuite.instrumentation.gomysql",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.redigo": {
		ScopeName: "loongsuite.instrumentation.redigo",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.mongo": {
		ScopeName: "loongsuite.instrumentation.mongo",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.gorm": {
		ScopeName: "loongsuite.instrumentation.gorm",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.gopg": {
		ScopeName: "loongsuite.instrumentation.gopg",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.gocql": {
		ScopeName: "loongsuite.instrumentation.gocql",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},
	"loongsuite.instrumentation.sqlx": {
		ScopeName: "loongsuite.instrumentation.sqlx",
		Category:  CategoryDB,
		ClientKey: DB_CLIENT_KEY,
		ServerKey: "",
	},

	// Messaging
	"loongsuite.instrumentation.amqp091": {
		ScopeName: "loongsuite.instrumentation.amqp091",
		Category:  CategoryMessaging,
		ClientKey: "",
		ServerKey: "",
	},
	"loongsuite.instrumentation.kafka-go": {
		ScopeName: "loongsuite.instrumentation.kafka-go",
		Category:  CategoryMessaging,
		ClientKey: "",
		ServerKey: "",
	},
	"loongsuite.instrumentation.rocketmq": {
		ScopeName: "loongsuite.instrumentation.rocketmq",
		Category:  CategoryMessaging,
		ClientKey: "",
		ServerKey: "",
	},

	// AI/LLM
	"loongsuite.instrumentation.eino": {
		ScopeName: "loongsuite.instrumentation.eino",
		Category:  CategoryAI,
		ClientKey: "",
		ServerKey: "",
	},
	"loongsuite.instrumentation.langchain": {
		ScopeName: "loongsuite.instrumentation.langchain",
		Category:  CategoryAI,
		ClientKey: "",
		ServerKey: "",
	},

	// Other
	"loongsuite.instrumentation.sentinel": {
		ScopeName: "loongsuite.instrumentation.sentinel",
		Category:  CategoryOther,
		ClientKey: "",
		ServerKey: "",
	},
}

// GetInstrumentationMetadata returns metadata for a given scope name
func GetInstrumentationMetadata(scopeName string) *InstrumentationMetadata {
	return InstrumentationRegistry[scopeName]
}

// GetCategory returns the category for a given scope name
func GetCategory(scopeName string) InstrumentationCategory {
	if metadata := InstrumentationRegistry[scopeName]; metadata != nil {
		return metadata.Category
	}
	return CategoryOther
}
