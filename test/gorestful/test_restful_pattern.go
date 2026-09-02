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

package main

import (
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"net/http"
	"strconv"
	"time"

	"github.com/alibaba/loongsuite-go/test/verifier"
	restful "github.com/emicklei/go-restful/v3"
)

type userResource struct{}

func (u userResource) WebService() *restful.WebService {
	ws := &restful.WebService{}

	ws.Path("/users").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.Route(ws.GET("/{user-id}").To(u.getUser).
		Param(ws.PathParameter("user-id", "identifier of the user").DataType("integer").DefaultValue("1")).
		Writes(user{}). // on the response
		Returns(http.StatusOK, "OK", user{}).
		Returns(http.StatusNotFound, "Not Found", nil))
	return ws
}

func (u userResource) getUser(req *restful.Request, resp *restful.Response) {
	uid := req.PathParameter("user-id")
	id, err := strconv.Atoi(uid)
	if err == nil && id >= 100 {
		_ = resp.WriteEntity(user{id})
		return
	}
	_ = resp.WriteErrorString(http.StatusNotFound, "User could not be found.")
}

type user struct {
	ID int `json:"id" description:"identifier of the user"`
}

func setupHttp() {
	u := userResource{}
	restful.DefaultContainer.Add(u.WebService())

	_ = http.ListenAndServe(":8080", nil)
}

func main() {
	// starter server
	go setupHttp()
	time.Sleep(3 * time.Second)
	// use a http client to request to the server
	client := http.Client{}
	client.Get("http://127.0.0.1:8080/users/123")
	// verify trace
	verifier.WaitAndAssertTraces(func(stubs []tracetest.SpanStubs) {
		verifier.VerifyHttpClientAttributes(stubs[0][0], "GET", "GET", "http://127.0.0.1:8080/users/123", "http", "1.1", "tcp", "ipv4", "", "127.0.0.1:8080", 200, 0, 8080)
		verifier.VerifyHttpServerAttributes(stubs[0][1], "GET /users/{user-id}", "GET", "http", "tcp", "ipv4", "", "127.0.0.1:8080", "Go-http-client/1.1", "http", "/users/123", "", "/users/{user-id}", 200)
	}, 1)
}
