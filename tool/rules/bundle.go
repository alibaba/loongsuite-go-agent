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

package rules

import (
	"encoding/json"
	"path/filepath"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

const (
	MatchedRulesJsonFile = "matched_rules.json"
)

// RuleBundle is a collection of rules that matched with one compilation action
type RuleBundle struct {
	PackageName      string
	ImportPath       string
	FileRules        []*InstFileRule
	File2FuncRules   map[string]map[string][]*InstFuncRule
	File2StructRules map[string]map[string][]*InstStructRule
}

func NewRuleBundle(importPath string) *RuleBundle {
	return &RuleBundle{
		PackageName:      "",
		ImportPath:       importPath,
		FileRules:        make([]*InstFileRule, 0),
		File2FuncRules:   make(map[string]map[string][]*InstFuncRule),
		File2StructRules: make(map[string]map[string][]*InstStructRule),
	}
}

func (rb *RuleBundle) String() string {
	bs, _ := json.Marshal(rb)
	return string(bs)
}

func (rb *RuleBundle) IsValid() bool {
	return rb != nil &&
		(len(rb.FileRules) > 0 ||
			len(rb.File2FuncRules) > 0 ||
			len(rb.File2StructRules) > 0)
}

func (rb *RuleBundle) AddFile2FuncRule(file string, rule *InstFuncRule) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return ex.Error(err)
	}
	fn := rule.Function + "," + rule.ReceiverType
	util.Assert(fn != "", "sanity check")
	if _, exist := rb.File2FuncRules[file]; !exist {
		rb.File2FuncRules[file] = make(map[string][]*InstFuncRule)
		rb.File2FuncRules[file][fn] = []*InstFuncRule{rule}
	} else {
		rb.File2FuncRules[file][fn] =
			append(rb.File2FuncRules[file][fn], rule)
	}
	return nil
}

func (rb *RuleBundle) AddFile2StructRule(file string, rule *InstStructRule) error {
	file, err := filepath.Abs(file)
	if err != nil {
		return ex.Error(err)
	}
	st := rule.StructType
	util.Assert(st != "", "sanity check")
	if _, exist := rb.File2StructRules[file]; !exist {
		rb.File2StructRules[file] = make(map[string][]*InstStructRule)
		rb.File2StructRules[file][st] = []*InstStructRule{rule}
	} else {
		rb.File2StructRules[file][st] =
			append(rb.File2StructRules[file][st], rule)
	}
	return nil
}

func (rb *RuleBundle) SetPackageName(name string) {
	rb.PackageName = name
}

func (rb *RuleBundle) AddFileRule(rule *InstFileRule) {
	rb.FileRules = append(rb.FileRules, rule)
}

func StoreRuleBundles(bundles []*RuleBundle) error {
	util.GuaranteeInPreprocess()
	ruleFile := util.GetPreprocessLogPath(MatchedRulesJsonFile)
	bs, err := json.Marshal(bundles)
	if err != nil {
		return ex.Error(err)
	}
	_, err = util.WriteFile(ruleFile, string(bs))
	if err != nil {
		return ex.Error(err)
	}
	return nil
}

func LoadRuleBundles() ([]*RuleBundle, error) {
	util.GuaranteeInInstrument()

	ruleFile := util.GetPreprocessLogPath(MatchedRulesJsonFile)
	data, err := util.ReadFile(ruleFile)
	if err != nil {
		return nil, ex.Error(err)
	}
	var bundles []*RuleBundle
	err = json.Unmarshal([]byte(data), &bundles)
	if err != nil {
		return nil, ex.Errorf(err, "bad "+ruleFile)
	}
	return bundles, nil
}
