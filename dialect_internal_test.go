package genieql

import (
	"database/sql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dialect", func() {
	Describe("dialectRegistry", func() {
		Describe("RegisterDialect", func() {
			It("should err if the dialect is already registered", func() {
				dialect := testDialect{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect("testDialect", dialect)).ToNot(HaveOccurred())
				Expect(reg.RegisterDialect("testDialect", dialect)).To(MatchError(ErrDuplicateDialect))
			})

			It("should register a dialect", func() {
				dialect := testDialect{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect("testDialect", dialect)).ToNot(HaveOccurred())
			})
		})

		Describe("LookupDialect", func() {
			It("should err if the dialect is not registered", func() {
				reg := dialectRegistry{}
				dialect, err := reg.LookupDialect("testDialect")
				Expect(dialect).To(BeNil())
				Expect(err).To(MatchError(ErrMissingDialect))
			})

			It("should return the dialect if its been registered", func() {
				dialectName := "testDialect"
				dialect := testDialect{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect(dialectName, dialect)).ToNot(HaveOccurred())
				foundDialect, err := reg.LookupDialect(dialectName)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundDialect).To(Equal(dialect))
			})
		})
	})
})

type testDialect struct {
	insertq     string
	selectq     string
	updateq     string
	deleteq     string
	columnq     string
	primarykeyq string
}

func (t testDialect) Insert(table string, columns, defaults []string) string {
	return t.insertq
}

func (t testDialect) Select(table string, columns, predicates []string) string {
	return t.selectq
}

func (t testDialect) Update(table string, columns, predicates []string) string {
	return t.updateq
}

func (t testDialect) Delete(table string, columns, predicates []string) string {
	return t.deleteq
}

func (t testDialect) ColumnQuery(table string) string {
	return t.columnq
}

func (t testDialect) PrimaryKeyQuery(table string) string {
	return t.primarykeyq
}

func (t testDialect) ColumnInformation(db *sql.DB, table string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}

func (t testDialect) ColumnInformationForQuery(db *sql.DB, query string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}
