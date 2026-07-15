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

package db

import (
	"fmt"
	"reflect"

	"github.com/alibaba/loongsuite-go/pkg/inst-api-semconv/instrumenter/utils"
)

// NormalizeDBClientErrorType returns a low-cardinality error.type value for
// failed DB client operations, matching HTTP/otelc style (concrete Go type name).
// Empty when err is nil.
func NormalizeDBClientErrorType(err error) string {
	if err == nil {
		return ""
	}
	// Guard against typed nil (e.g. (*MyError)(nil) stored in error interface).
	if v := reflect.ValueOf(err); v.Kind() == reflect.Ptr && v.IsNil() {
		return ""
	}

	if et, ok := err.(interface{ ErrorType() string }); ok {
		if s := et.ErrorType(); s != "" {
			return s
		}
	}

	err = utils.UnwrapFmtWrapped(err)
	if err == nil {
		return ""
	}

	if et, ok := err.(interface{ ErrorType() string }); ok {
		if s := et.ErrorType(); s != "" {
			return s
		}
	}

	t := reflect.TypeOf(err)
	if t == nil {
		return ""
	}
	if t.PkgPath() == "" && t.Name() == "" {
		return t.String()
	}
	return fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
}
