package ducktype_test

import (
	"database/sql"
	"flag"
	"os"
	"testing"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/james-lawrence/genieql/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	flag.Parse()
	testx.Logging()
	os.Exit(m.Run())
}

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("duckdb", "")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}
