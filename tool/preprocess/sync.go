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

package preprocess

import (
	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

func (dp *DepProcessor) runModTidy() error {
	out, err := runCmdCombinedOutput(dp.getGoModDir(),
		nil, "go", "mod", "tidy")
	util.Log("Run go mod tidy: %v", out)
	if err != nil {
		return ex.Errorf(err, "failed to run go mod tidy %s", string(out))
	}
	return nil
}

func (dp *DepProcessor) runModVendor() error {
	out, err := runCmdCombinedOutput(dp.getGoModDir(),
		nil, "go", "mod", "vendor")
	util.Log("Run go mod vendor: %v", out)
	if err != nil {
		return ex.Errorf(err, "failed to run go mod vendor %s", string(out))
	}
	return nil
}

func (dp *DepProcessor) syncDeps() error {
	// Run go mod tidy to remove unused dependencies
	err := dp.runModTidy()
	if err != nil {
		return ex.Error(err)
	}

	// Run go mod vendor to update the vendor directory
	if dp.vendorMode {
		err = dp.runModVendor()
		if err != nil {
			return ex.Error(err)
		}
	}

	return nil
}
