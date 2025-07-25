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

package config

import (
	"fmt"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

// @@This value is specified by the build system.
// This is the version of the tool, which will be printed when the -version flag
// is passed.
var ToolVersion = "1.0.0"

func PrintVersion() error {
	name, err := util.GetToolName()
	if err != nil {
		return ex.Error(err)
	}
	fmt.Printf("%s version %s\n", name, ToolVersion)
	return nil
}
