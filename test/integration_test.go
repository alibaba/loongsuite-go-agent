// Copyright (c) 2024 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package test

import (
	_ "github.com/alibaba/opentelemetry-go-auto-instrumentation/test/version"
	"testing"
)

func TestPlugins(t *testing.T) {
	for _, c := range TestCases {
		if c.TestName != "logrus-test" {
			continue
		}
		if c == nil {
			continue
		}
		if c.IsMuzzleCheck || c.IsLatestDepthCheck {
			continue
		}
		t.Run(c.TestName, func(t *testing.T) {
			c.TestFunc(t)
		})
	}
}

func TestMuzzle(t *testing.T) {
	for _, c := range TestCases {
		if c == nil {
			continue
		}
		if !c.IsMuzzleCheck {
			continue
		}
		t.Run(c.TestName, func(t *testing.T) {
			ExecMuzzle(t, c.DependencyName, c.ModuleName, c.MinVersion, c.MaxVersion, c.MuzzleClasses)
		})
	}
}

func TestLatest(t *testing.T) {
	for _, c := range TestCases {
		if c == nil {
			continue
		}
		if !c.IsLatestDepthCheck {
			continue
		}
		t.Run(c.TestName, func(t *testing.T) {
			ExecLatestTest(t, c.DependencyName, c.ModuleName, c.MinVersion, c.MaxVersion, c.LatestDepthFunc)
		})
	}
}
