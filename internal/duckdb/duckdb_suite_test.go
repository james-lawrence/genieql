package duckdb_test

import (
	"database/sql"
	"testing"

	"github.com/james-lawrence/genieql/internal/sqlxtest"
	_ "github.com/marcboeker/go-duckdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDuckdb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Duckdb Suite")
}

var (
	TX     *sql.Tx
	DB     *sql.DB
	dbname string
)

var _ = BeforeSuite(func() {
	dbname, DB = sqlxtest.NewDuckDB(sqlxtest.TemplateDatabaseName)
})

var _ = AfterSuite(func() {
	Expect(DB.Close()).ToNot(HaveOccurred())
	sqlxtest.DestroyDuckDB(sqlxtest.TemplateDatabaseName, dbname)
})

var _ = BeforeEach(func() {
	var err error
	TX, err = DB.Begin()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(TX.Rollback()).ToNot(HaveOccurred())
})
