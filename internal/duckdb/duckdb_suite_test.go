package duckdb_test

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/goosex"
	"github.com/james-lawrence/genieql/internal/sqlxtest"
	"github.com/james-lawrence/genieql/internal/testx"
	"github.com/pressly/goose/v3"
)

var (
	TX *sql.Tx
	DB *sql.DB
)

func TestMain(m *testing.M) {
	flag.Parse()
	testx.Logging()
	DB = sqlxtest.NewDuckDB()
	errorsx.MaybePanic(sqlxtest.Migrate(context.Background(), DB, os.DirFS("../../.migrations/duckdb"), goose.WithStore(goosex.DuckdbStore{})))
	code := m.Run()
	errorsx.MaybePanic(DB.Close())
	os.Exit(code)
}
