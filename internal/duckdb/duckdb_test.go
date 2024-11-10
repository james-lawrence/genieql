package duckdb_test

import (
	"github.com/james-lawrence/genieql"

	"github.com/james-lawrence/genieql/internal/drivers"
	. "github.com/james-lawrence/genieql/internal/duckdb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("duckdb", func() {
	Describe("dialect", func() {
		driver := genieql.MustLookupDriver(drivers.DuckDB)
		It("should return the columns in the query in sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForQuery(
				driver,
				"SELECT total_time, num_calls, avg_time, max_time FROM duckdb_functions",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.ColumnInfoSet(info).ColumnNames()).To(Equal([]string{"avg_time", "max_time", "num_calls", "total_time"}))
		})

		It("should return the columns in the table in the sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForTable(driver, "duckdb_functions")
			Expect(err).ToNot(HaveOccurred())
			Expect(
				genieql.ColumnInfoSet(info).ColumnNames(),
			).To(
				Equal([]string{
					"avg_time",
					"database_name",
					"function_name",
					"max_time",
					"min_time",
					"num_calls",
					"schema_name",
					"total_time",
				}),
			)
		})

		It("should support insert queries", func() {
			q := NewDialect(DB).Insert(1, 0, "table", "", []string{"c1", "c2", "c2"}, []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("INSERT INTO `table` (`c1`,`c2`,`c2`) VALUES (DEFAULT,?1,?2) RETURNING *"))
		})

		It("should support select queries", func() {
			q := NewDialect(DB).Select("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("SELECT `c1`,`c2`,`c2` FROM `table` WHERE `c1` = ?1"))
		})

		It("should support update queries", func() {
			q := NewDialect(DB).Update("table", []string{"c1", "c2", "c2"}, []string{"c1"}, []string{"c1", "c2", "c2"})
			Expect(q).To(Equal("UPDATE `table` SET `c1` = ?1, `c2` = ?2, `c2` = ?3 WHERE `c1` = ?4 RETURNING *"))
		})

		It("should support delete queries", func() {
			q := NewDialect(DB).Delete("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("DELETE FROM `table` WHERE `c1` = ?1"))
		})
	})
})
