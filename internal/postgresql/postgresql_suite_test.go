package postgresql_test

import (
	"database/sql"

	"bitbucket.org/jatone/genieql/xsqltest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPostgresql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgresql Suite")
}

var (
	TX     *sql.Tx
	DB     *sql.DB
	dbname string
)

var _ = BeforeSuite(func() {
	dbname, DB = xsqltest.NewPostgresql(xsqltest.TemplateDatabaseName)
})

var _ = AfterSuite(func() {
	Expect(DB.Close()).ToNot(HaveOccurred())
	xsqltest.DestroyPostgresql(xsqltest.TemplateDatabaseName, dbname)
})

var _ = BeforeEach(func() {
	var err error
	TX, err = DB.Begin()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(TX.Rollback()).ToNot(HaveOccurred())
})
