// Copyright (c) 2025 Alibaba Group Holding Ltd.
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

package sqlx

import (
	"context"
	"database/sql"
	"github.com/alibaba/loongsuite-go/pkg/api"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	_ "unsafe"
)

type sqlxInnerEnabler struct {
	enabled bool
}

func (s sqlxInnerEnabler) Enable() bool {
	return s.enabled
}

var sqlxEnabler = sqlxInnerEnabler{
	enabled: os.Getenv("OTEL_INSTRUMENTATION_SQLX_ENABLED") != "false",
}

var sqlInstrumenter = BuildSqlxInstrumenter()

//go:linkname beforeConnect github.com/jmoiron/sqlx.beforeConnect
func beforeConnect(ctx api.CallContext, driverName, dataSourceName string) {
	if !sqlxEnabler.Enable() {
		return
	}
	dbConfig, err := parseDSN(driverName, dataSourceName)
	if err != nil {
		log.Printf("fail to parse endpoint/dbName, err: %v.\n", err)
	}
	ctx.SetData(map[string]string{
		"endpoint": dbConfig.Endpoint,
		"dbName":   dbConfig.DBName,
	})
}

//go:linkname afterConnect github.com/jmoiron/sqlx.afterConnect
func afterConnect(ctx api.CallContext, db *sqlx.DB, err error) {
	if !sqlxEnabler.Enable() {
		return
	}
	if err != nil || db == nil {
		return
	}
	callData, ok := ctx.GetData().(map[string]string)
	if !ok {
		return
	}
	db.DbName = callData["dbName"]
	db.Endpoint = callData["endpoint"]
}

//go:linkname beforeQueryx github.com/jmoiron/sqlx.beforeQueryx
func beforeQueryx(ctx api.CallContext, db *sqlx.DB, query string, args ...interface{}) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := sqlxRequest{
		opType:     extractOpType(query),
		statement:  query,
		endpoint:   db.Endpoint,
		driverName: db.DriverName(),
		dbName:     db.DbName,
		params:     args,
	}
	sqlInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQueryx github.com/jmoiron/sqlx.afterQueryx
func afterQueryx(ctx api.CallContext, _ *sqlx.Rows, err error) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := ctx.GetData().(sqlxRequest)
	sqlInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeQueryRowx github.com/jmoiron/sqlx.beforeQueryRowx
func beforeQueryRowx(ctx api.CallContext, db *sqlx.DB, query string, args ...interface{}) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := sqlxRequest{
		opType:     extractOpType(query),
		statement:  query,
		endpoint:   db.Endpoint,
		driverName: db.DriverName(),
		dbName:     db.DbName,
		params:     args,
	}
	sqlInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQueryRowx github.com/jmoiron/sqlx.afterQueryRowx
func afterQueryRowx(ctx api.CallContext, row *sqlx.Row) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := ctx.GetData().(sqlxRequest)
	sqlInstrumenter.End(context.Background(), request, nil, row.Err())
}

//go:linkname beforeNamedExec github.com/jmoiron/sqlx.beforeNamedExec
func beforeNamedExec(ctx api.CallContext, db *sqlx.DB, query string, arg interface{}) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := sqlxRequest{
		opType:     extractOpType(query),
		statement:  query,
		endpoint:   db.Endpoint,
		driverName: db.DriverName(),
		dbName:     db.DbName,
		params:     []any{arg},
	}
	sqlInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterNamedExec github.com/jmoiron/sqlx.afterNamedExec
func afterNamedExec(ctx api.CallContext, _ sql.Result, err error) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := ctx.GetData().(sqlxRequest)
	sqlInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeQueryxContext github.com/jmoiron/sqlx.beforeQueryxContext
func beforeQueryxContext(ctx api.CallContext, db *sqlx.DB, _ context.Context, query string, args ...interface{}) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := sqlxRequest{
		opType:     extractOpType(query),
		statement:  query,
		endpoint:   db.Endpoint,
		driverName: db.DriverName(),
		dbName:     db.DbName,
		params:     args,
	}
	sqlInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQueryxContext github.com/jmoiron/sqlx.afterQueryxContext
func afterQueryxContext(ctx api.CallContext, _ *sqlx.Rows, err error) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := ctx.GetData().(sqlxRequest)
	sqlInstrumenter.End(context.Background(), request, nil, err)
}

//go:linkname beforeQueryRowxContext github.com/jmoiron/sqlx.beforeQueryRowxContext
func beforeQueryRowxContext(ctx api.CallContext, db *sqlx.DB, _ context.Context, query string, args ...interface{}) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := sqlxRequest{
		opType:     extractOpType(query),
		statement:  query,
		endpoint:   db.Endpoint,
		driverName: db.DriverName(),
		dbName:     db.DbName,
		params:     args,
	}
	sqlInstrumenter.Start(context.Background(), request)
	ctx.SetData(request)
}

//go:linkname afterQueryRowxContext github.com/jmoiron/sqlx.afterQueryRowxContext
func afterQueryRowxContext(ctx api.CallContext, row *sqlx.Row) {
	if !sqlxEnabler.Enable() {
		return
	}
	request := ctx.GetData().(sqlxRequest)
	sqlInstrumenter.End(context.Background(), request, nil, row.Err())
}
