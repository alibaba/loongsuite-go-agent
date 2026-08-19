// Copyright (c) 2026 Alibaba Group Holding Ltd.
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

package test

import "testing"

const aws_sdk_go_dependency_name = "github.com/aws/aws-sdk-go"
const aws_sdk_go_module_name = "aws-sdk-go"

func init() {
	tc1 := NewGeneralTestCase("aws-sdk-go-s3-retry-test", aws_sdk_go_module_name, "v1.55.5", "", "1.24", "", TestAwsSdkGoS3Retry)
	tc2 := NewMuzzleTestCase("aws-sdk-go-muzzle-test", aws_sdk_go_dependency_name, aws_sdk_go_module_name, "v1.55.5", "", "1.24", "", []string{"go", "build", "test_s3_retry.go"})
	// The rule reaches session internals through go:linkname and the version
	// range is left open, so the latest release has to be exercised, not just
	// compiled against.
	tc3 := NewLatestDepthTestCase("aws-sdk-go-latestdepth", aws_sdk_go_dependency_name, aws_sdk_go_module_name, "v1.55.5", "", "1.24", "", TestAwsSdkGoS3Retry)

	if tc1 != nil {
		TestCases = append(TestCases, tc1)
	}
	if tc2 != nil {
		TestCases = append(TestCases, tc2)
	}
	if tc3 != nil {
		TestCases = append(TestCases, tc3)
	}
}

func TestAwsSdkGoS3Retry(t *testing.T, env ...string) {
	UseApp("aws-sdk-go/v1.55.5")
	RunGoBuild(t, "go", "build", "test_s3_retry.go")
	RunApp(t, "./test_s3_retry", env...)
}
