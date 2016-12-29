package genieql

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dialect", func() {
	Describe("dialectRegistry", func() {
		Describe("RegisterDialect", func() {
			It("should err if the dialect is already registered", func() {
				dialect := testDialectFactory{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect("testDialect", dialect)).ToNot(HaveOccurred())
				Expect(reg.RegisterDialect("testDialect", dialect)).To(MatchError(ErrDuplicateDialect))
			})

			It("should register a dialect", func() {
				dialect := testDialectFactory{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect("testDialect", dialect)).ToNot(HaveOccurred())
			})
		})

		Describe("LookupDialect", func() {
			It("should err if the dialect is not registered", func() {
				reg := dialectRegistry{}
				dialect, err := reg.LookupDialect("testDialect")
				Expect(dialect).To(BeNil())
				Expect(err).To(MatchError("dialect (testDialect) is not registered"))
			})

			It("should return the dialect if its been registered", func() {
				dialectName := "testDialect"
				dialect := testDialectFactory{}
				reg := dialectRegistry{}
				Expect(reg.RegisterDialect(dialectName, dialect)).ToNot(HaveOccurred())
				foundDialect, err := reg.LookupDialect(dialectName)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundDialect).To(Equal(dialect))
			})
		})
	})
})

type testDialectFactory testDialect

func (t testDialectFactory) Connect(Configuration) (Dialect, error) {
	return testDialect(t), nil
}

type testDialect struct {
	insertq string
	selectq string
	updateq string
	deleteq string
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

func (t testDialect) ColumnValueTransformer() ColumnTransformer {
	return NewColumnInfoNameTransformer()
}

func (t testDialect) ColumnInformationForQuery(query string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}

func (t testDialect) ColumnInformationForTable(table string) ([]ColumnInfo, error) {
	return []ColumnInfo{}, nil
}
