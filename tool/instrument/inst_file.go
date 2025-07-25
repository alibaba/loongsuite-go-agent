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

package instrument

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/rules"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

func (rp *RuleProcessor) applyFileRules(bundle *rules.RuleBundle) (err error) {
	for _, rule := range bundle.FileRules {
		if rule.FileName == "" {
			return ex.Errorf(nil, "no file name")
		}
		// Decorate the source code to remove //go:build exclude
		// and rename package name
		source, err := util.ReadFile(rule.FileName)
		if err != nil {
			return ex.Errorf(err, "filename %s", rule.FileName)
		}
		source = util.RemoveGoBuildComment(source)
		source = util.RenamePackage(source, bundle.PackageName)

		// Get last section of file path as file name
		fileName := filepath.Base(rule.FileName)
		target := filepath.Join(rp.workDir,
			fmt.Sprintf("otel_inst_file_%s", fileName))
		_, err = util.WriteFile(target, source)
		if err != nil {
			return ex.Error(err)
		}
		// Relocate the file dependency of the rule, any rules targeting the
		// file dependency specified by the rule should be updated to target the
		// new file
		rp.setRelocated(rule.FileName, target)

		// Append or replace the file to the compile arguments
		if rule.Replace {
			err = rp.replaceCompileArg(target, func(arg string) bool {
				return strings.HasSuffix(arg, fileName)
			})
			if err != nil {
				return ex.Errorf(err, "compileArgs %v, newArg %s",
					rp.compileArgs, target)
			}
		} else {
			rp.addCompileArg(target)
		}
		util.Log("Apply file rule %v (%v)", rule, rp.compileArgs)
		rp.saveDebugFile(target)
	}
	return nil
}
