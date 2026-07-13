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

package http

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

// NormalizeHTTPClientErrorType follows the upstream HTTPClientErrorType behavior
// and uses the concrete Go error type as the error.type value.
func NormalizeHTTPClientErrorType(err error) string {
	if err == nil {
		return ""
	}

	// 1. 先进行 ErrorType() 接口判定，允许自定义错误控制
	if et, ok := err.(interface{ ErrorType() string }); ok {
		if s := et.ErrorType(); s != "" {
			return s
		}
	}

	// 2. 解包 fmt.Errorf("%w") 和 errors.Join 产生的标准库包装类，获取底层具体错误
	err = unwrapFmtWrapped(err)
	if err == nil {
		return ""
	}

	// 3. 对解包后的具体错误，如果实现了 ErrorType() 同样优先调用
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
