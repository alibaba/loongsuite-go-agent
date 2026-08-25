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
	"net/http"
	"net/url"
	"strconv"
)

type netHttpRequest struct {
	method         string
	url            *url.URL
	host           string
	isTls          bool
	header         http.Header
	version        string
	requestHeaders string
	requestBody    string
}

type netHttpResponse struct {
	statusCode   int
	header       http.Header
	hasResponse  bool
	responseBody string
}

func getProtocolVersion(majorVersion, minorVersion int) string {
	if majorVersion == 1 && minorVersion == 0 {
		return "1.0"
	} else if majorVersion == 1 && minorVersion == 1 {
		return "1.1"
	} else if majorVersion == 1 && minorVersion == 2 {
		return "1.2"
	} else if majorVersion == 2 {
		return "2"
	} else if majorVersion == 3 {
		return "3"
	}
	return strconv.Itoa(majorVersion) + "." + strconv.Itoa(minorVersion)
}
