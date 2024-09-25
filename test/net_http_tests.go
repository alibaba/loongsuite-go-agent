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

package test

import "testing"

func init() {
	TestCases = append(TestCases,
		NewGeneralTestCase("nethttp-basic-test", "nethttp", "", "", "1.18", "", TestBasicNetHttp),
		NewGeneralTestCase("nethttp-http-2-test", "nethttp", "", "", "1.18", "", TestHttp2),
		NewGeneralTestCase("nethttp-https-test", "nethttp", "", "", "1.18", "", TestHttps),
	)
}

func TestBasicNetHttp(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunInstrument(t, "-debuglog", "--", "test_http.go", "http_server.go")
	RunApp(t, "test_http", env...)
}

func TestHttp2(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunInstrument(t, "-debuglog", "--", "test_http_2.go", "http_server.go")
	RunApp(t, "test_http_2", env...)
}

func TestHttps(t *testing.T, env ...string) {
	UseApp("nethttp")
	RunInstrument(t, "-debuglog", "--", "test_https.go", "http_server.go")
	RunApp(t, "test_https", env...)
}
