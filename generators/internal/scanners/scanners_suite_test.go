package scanners_test

import (
	"log"

	. "bitbucket.org/jatone/genieql/internal/sqlxtest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"database/sql"
	"testing"
)

func TestIntegrationTests(t *testing.T) {
	log.SetFlags(log.Flags() | log.Lshortfile)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scanners Suite")
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
