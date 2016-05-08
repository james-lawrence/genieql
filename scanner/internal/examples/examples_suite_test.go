package examples_test

import (
	. "bitbucket.org/jatone/genieql/internal/sqlxtest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"
	"testing"
)

func TestIntegrationTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Exampels Suite")
}

var TX *sql.Tx
var DB *sql.DB
var dbname string

var _ = BeforeSuite(func() {
	dbname, DB = NewPostgresql(TemplateDatabaseName)
})

var _ = AfterSuite(func() {
	Expect(DB.Close()).ToNot(HaveOccurred())
	DestroyPostgresql(TemplateDatabaseName, dbname)
})

var _ = BeforeEach(func() {
	var err error
	TX, err = DB.Begin()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(TX.Rollback()).ToNot(HaveOccurred())
})
