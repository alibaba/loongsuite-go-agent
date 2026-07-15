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

package redigo

import (
	"container/list"
	"context"
	"fmt"
	"github.com/gomodule/redigo/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

const max_queue_length = 2048

var configuredQueueLength int

var commandQueue = list.New()
var transactionQueue = list.New() // transaction queue
var redigoInstrumenter = BuildRedigoInstrumenter()

type armsConn struct {
	redis.Conn
	endpoint string
	ctx      context.Context
}

func (a *armsConn) Close() error {
	return a.Conn.Close()
}

func (a *armsConn) Err() error {
	return a.Conn.Err()
}

func (a *armsConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	req := &redigoRequest{
		args:     args,
		endpoint: a.endpoint,
		cmd:      commandName,
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	startTime := time.Now()
	reply, err = a.Conn.Do(commandName, args...)
	endTime := time.Now()
	switch strings.ToLower(commandName) {
	case "multi":
		// Start of a transaction
		push(req, transactionQueue)
	case "exec":
		// End of a transaction, we need to build the command string for all commands in the transaction
		string_builder := strings.Builder{}
		for transactionQueue.Len() > 0 {
			r := pop(transactionQueue)
			if r != nil && strings.ToLower(r.cmd) != "multi" {
				string_builder.WriteString(r.cmd + " ")
				for i, arg := range r.args {
					string_builder.WriteString(fmt.Sprintf("%v", arg))
					if i < len(r.args)-1 {
						string_builder.WriteString(" ")
					}
				}
				string_builder.WriteString("; ")
			}
		}
		req.cmd = "EXEC"
		req.args = []interface{}{string_builder.String()}
		redigoInstrumenter.StartAndEnd(ctx, req, nil, redigoSpanEndErr(err), startTime, endTime)
		transactionQueue.Init() // purge the transaction queue after EXEC
	case "discard":
		// Purge the transaction queue
		transactionQueue.Init()
	default:
		// Inspect if we are in a transaction, otherwise just push to the command queue
		if transactionQueue.Len() > 0 {
			push(req, transactionQueue)
			return
		}
		redigoInstrumenter.StartAndEnd(ctx, req, nil, redigoSpanEndErr(err), startTime, endTime)
	}
	return
}

func (a *armsConn) Send(commandName string, args ...interface{}) error {
	now := time.Now()
	req := &redigoRequest{
		args:      args,
		endpoint:  a.endpoint,
		cmd:       commandName,
		startTime: now,
	}
	ctx := a.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	req.ctx = ctx
	switch strings.ToLower(commandName) {
	case "multi":
		// Start of a transaction
		push(req, transactionQueue)
	case "discard":
		// Purge the transaction queue
		transactionQueue.Init()
	default:
		// Inspect if we are in a transaction, otherwise just push to the command queue
		if transactionQueue.Len() > 0 {
			push(req, transactionQueue)
			return a.Conn.Send(commandName, args...)
		}
		push(req, commandQueue)
	}
	return a.Conn.Send(commandName, args...)
}

func (a *armsConn) Flush() error {
	return a.Conn.Flush()
}

func (a *armsConn) Receive() (reply interface{}, err error) {
	reply, err = a.Conn.Receive()
	req := pop(commandQueue)
	if req != nil {
		now := time.Now()
		redigoInstrumenter.StartAndEnd(req.ctx, req, nil, redigoSpanEndErr(err), req.startTime, now)
	}
	return
}

func push(request *redigoRequest, queue *list.List) {
	if queue != nil && queue.Len() > getMaxQueueLength() {
		return
	}
	queue.PushBack(request)
}

func pop(queue *list.List) *redigoRequest {
	front := queue.Front()
	queue.Remove(front)
	p, ok := front.Value.(*redigoRequest)
	if ok {
		return p
	}
	return nil
}

func getMaxQueueLength() int {
	if configuredQueueLength == 0 {
		var e = os.Getenv("MAX_REDIGO_QUEUE_LENGTH")
		if e != "" {
			configuredQueueLength, _ = strconv.Atoi(os.Getenv(e))
		} else {
			configuredQueueLength = max_queue_length
		}
	}
	return configuredQueueLength
}
