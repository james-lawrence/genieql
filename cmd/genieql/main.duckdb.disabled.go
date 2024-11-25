//go:build !genieql.duckdb

package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

func (t *duckdb) execute(*kingpin.ParseContext) (err error) {
	return errorsx.String("genieql was not compiled with duckdb please add -tags genieql.duckdb,no_duckdb_arrow to your build")
}
