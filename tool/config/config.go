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

package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"

	"github.com/alibaba/loongsuite-go-agent/tool/ex"
	"github.com/alibaba/loongsuite-go-agent/tool/util"
)

const (
	EnvPrefix     = "OTELTOOL_"
	BuildConfFile = "conf.json"
)

type BuildConfig struct {
	// RuleJsonFiles is the name of the rule file. It is used to tell instrument
	// tool where to find the instrument rules. Multiple rules are separated by
	// comma. e.g. -rule=rule1.json,rule2.json. By default, new rules are appended
	// to default rules, i.e. -rule=rule1.json,rule2.json is exactly equivalent to
	// -rule=default.json,rule1.json,rule2.json. But if you do want to disable
	// default rules, you can configure -disable flag in advance.
	RuleJsonFiles string

	// Verbose true means print verbose log.
	Verbose bool

	// Debug true means debug mode.
	Debug bool

	// DisableRules specifies which rules to disable. It can be:
	// - "all" to disable all default rules
	// - comma-separated list of rule file names to disable specific rules
	//   e.g. "gorm.json,redis.json"
	// - empty string to enable all default rules
	// Note that base.json is inevitable to be enabled, even if it is explicitly
	// disabled.
	DisableRules string
}

var conf *BuildConfig

func GetConf() *BuildConfig {
	util.Assert(conf != nil, "build config is not initialized")
	return conf
}

func (bc *BuildConfig) IsDisableAll() bool {
	return bc.DisableRules == "all"
}

// GetDisabledRules returns a set of rule file names that should be disabled
func (bc *BuildConfig) GetDisabledRules() string {
	return bc.DisableRules
}

func (bc *BuildConfig) makeRuleAbs(file string) (string, error) {
	if util.PathNotExists(file) {
		return "", ex.Errorf(nil, "file %s not exists", file)
	}
	file, err := filepath.Abs(file)
	if err != nil {
		return "", ex.Error(err)
	}
	return file, nil
}

func (bc *BuildConfig) parseRuleFiles() error {
	if util.InInstrument() {
		return nil
	}
	// Get absolute path of rule file, otherwise instrument will not
	// be able to find the rule file because it is running in different
	// working directory.
	if bc.RuleJsonFiles == "" {
		return nil
	}
	if strings.Contains(bc.RuleJsonFiles, ",") {
		files := strings.Split(bc.RuleJsonFiles, ",")
		for i, file := range files {
			f, err := bc.makeRuleAbs(file)
			if err != nil {
				return ex.Error(err)
			}
			files[i] = f
		}
		bc.RuleJsonFiles = strings.Join(files, ",")
	} else {
		f, err := bc.makeRuleAbs(bc.RuleJsonFiles)
		if err != nil {
			return ex.Error(err)
		}
		bc.RuleJsonFiles = f
	}
	return nil
}

func getConfPath(name string) string {
	return util.GetTempBuildDirWith(name)
}

func storeConfig(bc *BuildConfig) error {
	util.Assert(bc != nil, "build config is not initialized")

	file := getConfPath(BuildConfFile)
	bs, err := json.Marshal(bc)
	if err != nil {
		return ex.Error(err)
	}
	_, err = util.WriteFile(file, string(bs))
	if err != nil {
		return ex.Error(err)
	}
	return nil
}

func loadConfig() (*BuildConfig, error) {
	util.Assert(conf == nil, "build config is already initialized")
	// If the build config file does not exist, return a default build config
	confFile := getConfPath(BuildConfFile)
	if util.PathNotExists(confFile) {
		return &BuildConfig{}, nil
	}
	// Load build config from json file
	file := getConfPath(BuildConfFile)
	data, err := util.ReadFile(file)
	if err != nil {
		return &BuildConfig{}, ex.Error(err)
	}
	bc := &BuildConfig{}
	err = json.Unmarshal([]byte(data), bc)
	if err != nil {
		return nil, ex.Error(err)
	}
	return bc, nil
}

func toUpperSnakeCase(input string) string {
	var result []rune

	for i, char := range input {
		if unicode.IsUpper(char) {
			if i != 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToUpper(char))
		} else {
			result = append(result, unicode.ToUpper(char))
		}
	}

	return string(result)
}
func loadConfigFromEnv(conf *BuildConfig) {
	// Environment variables are able to overwrite the config items even if the
	// config file sets them. The environment variable name is the upper snake
	// case of the config item name, prefixed with "OTELTOOL_". For example, the
	// environment variable for "Log" is "OTELTOOL_LOG".
	typ := reflect.TypeOf(*conf)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		envKey := fmt.Sprintf("%s%s", EnvPrefix, toUpperSnakeCase(field.Name))
		envVal := os.Getenv(envKey)
		if envVal != "" {
			if util.InPreprocess() {
				util.Log("Overwrite config %s with environment variable %s",
					field.Name, envKey)
			}
			v := reflect.ValueOf(conf).Elem()
			f := v.FieldByName(field.Name)
			switch f.Kind() {
			case reflect.Bool:
				f.SetBool(envVal == "true")
			case reflect.String:
				f.SetString(envVal)
			default:
				util.ShouldNotReachHere()
			}
		}
	}
}

func InitConfig() (err error) {
	// Load build config from json file
	conf, err = loadConfig()
	if err != nil {
		return ex.Error(err)
	}
	loadConfigFromEnv(conf)

	err = conf.parseRuleFiles()
	if err != nil {
		return ex.Error(err)
	}

	mode := os.O_WRONLY | os.O_APPEND
	if util.InPreprocess() {
		// We always create log file in preprocess phase, but in further
		// instrument phase, we append log content to the existing file.
		mode = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	// Always redirect log to debug log file
	debugLogPath := util.GetTempBuildDirWith(util.DebugLogFile)
	debugLog, _ := os.OpenFile(debugLogPath, mode, 0777)
	if debugLog != nil {
		util.SetLogger(debugLog)
	}
	return nil
}

func Configure() error {
	// Parse command line flags to get build config
	bc, err := loadConfig()
	if err != nil {
		bc = &BuildConfig{}
	}
	flag.BoolVar(&bc.Verbose, "verbose", bc.Verbose,
		"Print verbose log")
	flag.BoolVar(&bc.Debug, "debug", bc.Debug,
		"Enable debug mode, leave temporary files for debugging")
	flag.StringVar(&bc.RuleJsonFiles, "rule", bc.RuleJsonFiles,
		"Use custom.json rules. Multiple rules are separated by comma.")
	flag.StringVar(&bc.DisableRules, "disable", bc.DisableRules,
		"Disable specific rules. Use 'all' to disable all default rules, or comma-separated list of rule file names to disable specific rules")
	flag.CommandLine.Parse(os.Args[2:])

	util.Log("Configured in %s", getConfPath(BuildConfFile))

	// Store build config for future phases
	err = storeConfig(bc)
	if err != nil {
		return ex.Error(err)
	}
	return nil
}
