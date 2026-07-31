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

package databasesql

import (
	"errors"
	"fmt"
	nurl "net/url"
)

func parseDSN(driverName, dsn string) (addr string, err error) {
	// TODO: need a more delegate DFA
	switch driverName {
	case "mysql":
		return parseMySQL(dsn)
	case "postgres":
		fallthrough
	case "postgresql":
		return parsePostgres(dsn)
	case "sqlite", "sqlite3":
		// SQLite is an embedded database whose DSN is a file path (or
		// ":memory:") rather than a network address, so there is no
		// endpoint to extract and nothing to report as an error.
		return "", nil
	}

	return "", errors.New("invalid DSN")
}

func parsePostgres(url string) (addr string, err error) {
	u, err := nurl.Parse(url)
	if err != nil {
		return "", err
	}

	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("invalid connection protocol: %s", u.Scheme)
	}

	return u.Host + ":" + u.Port(), nil
}

func parseMySQL(dsn string) (addr string, err error) {
	n := len(dsn)
	i, j := -1, -1
	for k := 0; k < n; k++ {
		if dsn[k] == '(' {
			i = k
		}
		if dsn[k] == ')' {
			j = k
			break
		}
	}
	if i >= 0 && j > i {
		return dsn[i+1 : j], nil
	}
	return "", errors.New("invalid MySQL DSN")
}
