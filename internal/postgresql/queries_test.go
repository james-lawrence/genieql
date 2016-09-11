package postgresql_test

import (
	. "bitbucket.org/jatone/genieql/internal/postgresql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("queries", func() {
	DescribeTable("Insert",
		func(table string, columns, defaults []string, query string) {
			Expect(Insert(table, columns, defaults)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{}, "INSERT INTO MyTable1 (col1,col2,col3) VALUES ($1,$2,$3) RETURNING col1,col2,col3"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col4"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES ($1,$2,$3,DEFAULT) RETURNING col1,col2,col3,col4"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col3"}, "INSERT INTO MyTable2 (col1,col2,col3,col4) VALUES (DEFAULT,$1,DEFAULT,$2) RETURNING col1,col2,col3,col4"),
	)

	DescribeTable("Select",
		func(table string, columns, predicates []string, query string) {
			Expect(Select(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "SELECT col1,col2,col3 FROM MyTable1 WHERE col1 = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "SELECT col1,col2,col3,col4 FROM MyTable2 WHERE col1 = $1 AND col2 = $2"),
	)

	DescribeTable("Update",
		func(table string, columns, predicates []string, query string) {
			Expect(Update(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "UPDATE MyTable1 SET col1 = $1, col2 = $2, col3 = $3 WHERE col1 = $4 RETURNING col1,col2,col3"),
		Entry("example 1", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE col1 = $5 AND col2 = $6 RETURNING col1,col2,col3,col4"),
		Entry("example 1", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "UPDATE MyTable2 SET col1 = $1, col2 = $2, col3 = $3, col4 = $4 WHERE 't' RETURNING col1,col2,col3,col4"),
	)

	DescribeTable("Delete",
		func(table string, columns, predicates []string, query string) {
			Expect(Delete(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"}, "DELETE FROM MyTable1 WHERE col1 = $1 RETURNING col1,col2,col3"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"}, "DELETE FROM MyTable2 WHERE col1 = $1 AND col2 = $2 RETURNING col1,col2,col3,col4"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{}, "DELETE FROM MyTable2 WHERE 't' RETURNING col1,col2,col3,col4"),
	)
})
