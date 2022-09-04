package functions_test

import (
	"database/sql"

	. "bitbucket.org/jatone/genieql/internal/sqlxtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFunctions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internal Functions Suite")
}

var (
	TX     *sql.Tx
	DB     *sql.DB
	dbname string
)

var _ = BeforeSuite(func() {
	dbname, DB = NewPostgresql(TemplateDatabaseName)
})

var _ = AfterSuite(func() {
	if DB != nil {
		Expect(DB.Close()).ToNot(HaveOccurred())
	}
	if len(dbname) > 0 {
		DestroyPostgresql(TemplateDatabaseName, dbname)
	}
})

var _ = BeforeEach(func() {
	var err error
	TX, err = DB.Begin()
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(TX.Rollback()).ToNot(HaveOccurred())
})
