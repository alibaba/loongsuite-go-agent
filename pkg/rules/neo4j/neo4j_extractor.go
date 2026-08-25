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

import "strings"

const defaultOp = "QUERY"

// extractOpType returns the leading Cypher clause keyword, e.g. "MATCH",
// "CREATE", "MERGE" or "CALL", which is used as the db.operation.name.
func extractOpType(statement string) string {
	s := strings.TrimSpace(statement)
	if s == "" {
		return defaultOp
	}
	end := strings.IndexAny(s, " \t\r\n(")
	if end == -1 {
		return strings.ToUpper(s)
	}
	return strings.ToUpper(s[:end])
}
