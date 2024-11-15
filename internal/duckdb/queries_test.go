package duckdb_test

import (
	. "github.com/james-lawrence/genieql/internal/duckdb"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("queries", func() {
	DescribeTable("Insert",
		func(n int, table string, conflict string, columns, defaults []string, query string) {
			Expect(Insert(n, 0, table, conflict, columns, columns, defaults)).To(Equal(query))
		},
		Entry("example 1", 1, "MyTable1", "", []string{"col1", "col2", "col3"}, []string{},
			"INSERT INTO `MyTable1` (`col1`,`col2`,`col3`) VALUES ($1,$2,$3) RETURNING `col1`,`col2`,`col3`"),
		Entry("example 2", 1, "MyTable2", "", []string{"col1", "col2", "col3", "col4"}, []string{"col4"},
			"INSERT INTO `MyTable2` (`col1`,`col2`,`col3`,`col4`) VALUES ($1,$2,$3,DEFAULT) RETURNING `col1`,`col2`,`col3`,`col4`"),
		Entry("example 3", 1, "MyTable2", "", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col3"},
			"INSERT INTO `MyTable2` (`col1`,`col2`,`col3`,`col4`) VALUES (DEFAULT,$1,DEFAULT,$2) RETURNING `col1`,`col2`,`col3`,`col4`"),
		Entry("example 4", 3, "MyTable2", "", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col3"},
			"INSERT INTO `MyTable2` (`col1`,`col2`,`col3`,`col4`) VALUES (DEFAULT,$1,DEFAULT,$2),(DEFAULT,$3,DEFAULT,$4),(DEFAULT,$5,DEFAULT,$6) RETURNING `col1`,`col2`,`col3`,`col4`"),
	)

	DescribeTable("Select",
		func(table string, columns, predicates []string, query string) {
			Expect(Select(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"},
			"SELECT `col1`,`col2`,`col3` FROM `MyTable1` WHERE `col1` = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"},
			"SELECT `col1`,`col2`,`col3`,`col4` FROM `MyTable2` WHERE `col1` = $1 AND `col2` = $2"),
	)

	DescribeTable("Update",
		func(table string, columns, predicates []string, query string) {
			Expect(Update(table, columns, predicates, columns)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"},
			"UPDATE `MyTable1` SET `col1` = $1, `col2` = $2, `col3` = $3 WHERE `col1` = $4 RETURNING `col1`,`col2`,`col3`"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"},
			"UPDATE `MyTable2` SET `col1` = $1, `col2` = $2, `col3` = $3, `col4` = $4 WHERE `col1` = $5 AND `col2` = $6 RETURNING `col1`,`col2`,`col3`,`col4`"),
		Entry("example 3", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{},
			"UPDATE `MyTable2` SET `col1` = $1, `col2` = $2, `col3` = $3, `col4` = $4 WHERE TRUE RETURNING `col1`,`col2`,`col3`,`col4`"),
	)

	DescribeTable("Delete",
		func(table string, columns, predicates []string, query string) {
			Expect(Delete(table, columns, predicates)).To(Equal(query))
		},
		Entry("example 1", "MyTable1", []string{"col1", "col2", "col3"}, []string{"col1"},
			"DELETE FROM `MyTable1` WHERE `col1` = $1"),
		Entry("example 2", "MyTable2", []string{"col1", "col2", "col3", "col4"}, []string{"col1", "col2"},
			"DELETE FROM `MyTable2` WHERE `col1` = $1 AND `col2` = $2"),
	)
})
