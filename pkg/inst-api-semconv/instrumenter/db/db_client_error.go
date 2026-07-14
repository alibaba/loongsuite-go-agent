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
	"errors"
	"fmt"
	"reflect"
)

func unwrapFmtWrapped(err error) error {
	for err != nil {
		t := reflect.TypeOf(err)
		if t == nil {
			break
		}
		name := t.String()
		if name == "*fmt.wrapError" {
			err = errors.Unwrap(err)
		} else if name == "*errors.joinError" {
			if je, ok := err.(interface{ Unwrap() []error }); ok {
				errs := je.Unwrap()
				if len(errs) > 0 {
					err = errs[0]
				} else {
					break
				}
			} else {
				break
			}
		} else {
			break
		}
	}
	return err
}

// NormalizeDBClientErrorType returns a low-cardinality error.type value for
// failed DB client operations, matching HTTP/otelc style (concrete Go type name).
// Empty when err is nil.
func NormalizeDBClientErrorType(err error) string {
	if err == nil {
		return ""
	}

	if et, ok := err.(interface{ ErrorType() string }); ok {
		if s := et.ErrorType(); s != "" {
			return s
		}
	}

	err = unwrapFmtWrapped(err)
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
