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
package rules

import (
	"encoding/json"
	"strings"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

// -----------------------------------------------------------------------------
// Instrumentation Rule
//
// Instrumentation rules are used to define the behavior of the instrumentation
// for a specific function call. The rules are defined in the init() function
// of rule.go in each package directory. The rules are then used by the instrument
// package to generate the instrumentation code. Multiple rules can be defined
// for a single function call, and the rules are executed in the order of their
// priority. The rules are executed
// in the order of their priority, from high to low.
// There are several types of rules for different purposes:
// - InstFuncRule: Instrumentation rule for a specific function call
// - InstStructRule: Instrumentation rule for a specific struct type
// - InstFileRule: Instrumentation rule for a specific file

type InstRule interface {
	GetVersion() string    // GetVersion returns the version of the rule
	GetGoVersion() string  // GetGoVersion returns the go version of the rule
	GetImportPath() string // GetImportPath returns import path of the rule
	GetPath() string       // GetPath returns the local path of the rule
	SetPath(path string)   // SetPath sets the local path of the rule
	String() string        // String returns string representation of rule
	Verify() error         // Verify checks the rule is valid
}

type InstBaseRule struct {
	// Local path of the rule, it designates where we can found the hook code
	Path string `json:"Path,omitempty"`
	// Version of the rule, e.g. "[1.9.1,1.9.2)" or "", it designates the
	// version range of rule, all other version will not be instrumented
	Version string `json:"Version,omitempty"`
	// Go version of the rule, e.g. "[1.22.0,)" or "", it designates the go
	// version range of rule, all other go version will not be instrumented
	GoVersion string `json:"GoVersion,omitempty"`
	// Import path of the rule, e.g. "github.com/gin-gonic/gin", it designates
	// the import path of rule, all other import path will not be instrumented
	ImportPath string `json:"ImportPath,omitempty"`
}

func (rule *InstBaseRule) GetVersion() string {
	return rule.Version
}

func (rule *InstBaseRule) GetGoVersion() string {
	return rule.GoVersion
}

func (rule *InstBaseRule) GetImportPath() string {
	return rule.ImportPath
}

func (rule *InstBaseRule) GetPath() string {
	return rule.Path
}

func (rule *InstBaseRule) SetPath(path string) {
	rule.Path = path
}

// InstFuncRule finds specific function call and instrument by adding new code
type InstFuncRule struct {
	InstBaseRule
	// Function name, e.g. "New"
	Function string `json:"Function,omitempty"`
	// Receiver type name, e.g. "*gin.Engine"
	ReceiverType string `json:"ReceiverType,omitempty"`
	// Order of the rule, higher is executed first
	Order int `json:"Order,omitempty"`
	// UseRaw indicates whether to insert raw code string
	UseRaw bool `json:"UseRaw,omitempty"`
	// OnEnter callback, called before original function
	OnEnter string `json:"OnEnter,omitempty"`
	// OnExit callback, called after original function
	OnExit string `json:"OnExit,omitempty"`
}

// InstStructRule finds specific struct type and instrument by adding new field
type InstStructRule struct {
	InstBaseRule
	// Struct type name, e.g. "Engine"
	StructType string `json:"StructType,omitempty"`
	// New field name, e.g. "Logger"
	FieldName string `json:"FieldName,omitempty"`
	// New field type, e.g. "zap.Logger"
	FieldType string `json:"FieldType,omitempty"`
}

// InstFileRule adds user file into compilation unit and do further compilation
type InstFileRule struct {
	InstBaseRule
	// File name, e.g. "engine.go"
	FileName string `json:"FileName,omitempty"`
	// Replace indicates whether to replace the original file
	Replace bool `json:"Replace,omitempty"`
}

// String returns string representation of the rule
func (rule *InstFuncRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}
func (rule *InstStructRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}
func (rule *InstFileRule) String() string {
	bs, _ := json.Marshal(rule)
	return string(bs)
}

// Verify checks the rule is valid
func verifyRule(rule *InstBaseRule, checkPath bool) error {
	if checkPath {
		if rule.Path == "" {
			return ex.Errorf(nil, "local path is empty")
		}
	}
	// Import path should not be empty
	if rule.ImportPath == "" {
		return ex.Errorf(nil, "import path is empty")
	}
	// If version is specified, it should be in the format of [start,end)
	for _, v := range []string{rule.Version, rule.GoVersion} {
		if v != "" {
			if !strings.Contains(v, "[") ||
				!strings.Contains(v, ")") ||
				!strings.Contains(v, ",") ||
				strings.Contains(v, "v") {
				return ex.Errorf(nil, "bad version "+v)
			}
		}
	}
	return nil
}

func verifyRuleBase(rule *InstBaseRule) error {
	return verifyRule(rule, false)
}

func verifyRuleBaseWithoutPath(rule *InstBaseRule) error {
	return verifyRule(rule, true)
}

func (rule *InstFileRule) Verify() error {
	err := verifyRuleBase(&rule.InstBaseRule)
	if err != nil {
		return ex.Error(err)
	}
	if rule.FileName == "" {
		return ex.Errorf(nil, "empty file name")
	}
	if !util.IsGoFile(rule.FileName) {
		return ex.Errorf(nil, "not a go file")
	}
	return nil
}

func (rule *InstFuncRule) Verify() error {
	var err error
	if rule.UseRaw {
		err = verifyRuleBaseWithoutPath(&rule.InstBaseRule)
	} else {
		err = verifyRuleBase(&rule.InstBaseRule)
	}
	if err != nil {
		return ex.Error(err)
	}
	if rule.Function == "" {
		return ex.Errorf(nil, "empty function name")
	}
	if rule.OnEnter == "" && rule.OnExit == "" {
		return ex.Errorf(nil, "empty hook")
	}
	return nil
}

func (rule *InstStructRule) Verify() error {
	err := verifyRuleBaseWithoutPath(&rule.InstBaseRule)
	if err != nil {
		return ex.Error(err)
	}
	if rule.StructType == "" {
		return ex.Errorf(nil, "empty struct type")
	}
	if rule.FieldName == "" || rule.FieldType == "" {
		return ex.Errorf(nil, "empty field name or type")
	}
	return nil
}
