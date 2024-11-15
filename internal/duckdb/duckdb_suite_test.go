package duckdb_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/james-lawrence/genieql/internal/goosex"
	"github.com/james-lawrence/genieql/internal/sqlxtest"
	"github.com/james-lawrence/genieql/internal/testx"
	_ "github.com/marcboeker/go-duckdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pressly/goose/v3"
)

func TestDuckdb(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Duckdb Suite")
}

var (
	TX *sql.Tx
	DB *sql.DB
)

var _ = BeforeSuite(func(ctx context.Context) {
	DB = sqlxtest.NewDuckDB()
	sqlxtest.Migrate(ctx, DB, os.DirFS("../../.migrations/duckdb"), goose.WithStore(goosex.DuckdbStore{}))
})

var _ = AfterSuite(func() {
	if DB == nil {
		return
	}
	testx.MaybePanic(DB.Close())
})

var _ = BeforeEach(func() {
	TX = testx.Must(DB.Begin())
})

var _ = AfterEach(func() {
	testx.MaybePanic(TX.Rollback())
})
