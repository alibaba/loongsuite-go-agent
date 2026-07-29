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

package fiberv3

import (
	"net/url"

	"github.com/valyala/fasthttp"
)

type fiberv3Request struct {
	method         string
	url            *url.URL
	isTls          bool
	port           int
	header         *fasthttp.RequestHeader
	requestHeaders string
	requestBody    string
}

type fiberv3Response struct {
	statusCode   int
	header       *fasthttp.ResponseHeader
	responseBody string
}
