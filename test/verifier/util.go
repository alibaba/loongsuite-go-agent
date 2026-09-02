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

package verifier

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
)

func GetAttribute(attrs []attribute.KeyValue, name string) attribute.Value {
	for _, attr := range attrs {
		if string(attr.Key) == name {
			return attr.Value
		}
	}
	return attribute.Value{}
}

func HasAttribute(attrs []attribute.KeyValue, name string) bool {
	for _, attr := range attrs {
		if string(attr.Key) == name {
			return true
		}
	}
	return false
}

func AssertAttributeMissing(attrs []attribute.KeyValue, name string) {
	Assert(!HasAttribute(attrs, name), "Expected attribute %s to be absent, but it was present", name)
}

// Get all attrs with a prefix like `db.query.parameter.0` with `db.query.parameter`
func GetAttributesWithPrefix(attrs []attribute.KeyValue, prefix string) []string {
	var results []string

	for _, attr := range attrs {
		attrKey := string(attr.Key)

		// The attr keys must conform to the format like `db.query.parameter.{idx}`.
		if strings.HasPrefix(attrKey, prefix) {
			remainder := attrKey[len(prefix):]
			if len(remainder) > 0 && strings.HasPrefix(remainder, ".") {
				indexStr := remainder[1:]

				if _, err := strconv.Atoi(indexStr); err == nil {
					results = append(results, attr.Value.AsString())
				}
			}
		}
	}

	return results
}

func Assert(cond bool, format string, args ...interface{}) {
	if !cond {
		panic(fmt.Sprintf(format, args...))
	}
}

type mockTestingT struct{}

func (m mockTestingT) Errorf(format string, args ...interface{}) {

}

// Compare two attrs of slice type
func SliceAttrsAssert(expectAttrs []any, actualAttrs []string, format string, args ...interface{}) {
	convertedSlice := make([]string, len(expectAttrs))
	for i, v := range expectAttrs {
		candidate, ok := v.(string)
		if ok {
			convertedSlice[i] = candidate
		} else {
			convertedSlice[i] = fmt.Sprintf("%v", v)
		}
	}

	mockT := &struct {
		assert.TestingT
	}{}
	mockT.TestingT = mockTestingT{}

	if !assert.ElementsMatch(mockT, actualAttrs, convertedSlice) {
		errorMsg := fmt.Sprintf(format, args...)
		errorMsg += fmt.Sprintf("\nExpected contains same elements:\nexpected: %v\nactual: %v",
			convertedSlice, actualAttrs)
		panic(errorMsg)
	}
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		panic("Failed to create a free port: " + err.Error())
	}
	cli, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic("Failed to create a free port: " + err.Error())
	}
	defer cli.Close()
	return cli.Addr().(*net.TCPAddr).Port, nil
}

func GetServer(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return resp.Status, nil
}
