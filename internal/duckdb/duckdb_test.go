package duckdb_test

import (
	"github.com/james-lawrence/genieql"

	"github.com/james-lawrence/genieql/internal/drivers"
	. "github.com/james-lawrence/genieql/internal/duckdb"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("duckdb", func() {
	Describe("dialect", func() {
		driver := testx.Must(genieql.LookupDriver(drivers.DuckDB))

		It("should return the columns in the query in sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForQuery(
				driver,
				"SELECT database_name, schema_oid, is_nullable FROM duckdb_columns",
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.ColumnInfoSet(info).ColumnNames()).To(Equal([]string{"database_name", "is_nullable", "schema_oid"}))
		})

		It("should return the columns in the table in the sorted order", func() {
			info, err := NewDialect(DB).ColumnInformationForTable(driver, "duckdb_columns")
			Expect(err).ToNot(HaveOccurred())
			Expect(
				genieql.ColumnInfoSet(info).ColumnNames(),
			).To(
				Equal([]string{
					"character_maximum_length",
					"column_default",
					"column_index",
					"column_name",
					"comment",
					"data_type",
					"data_type_id",
					"database_name",
					"database_oid",
					"internal",
					"is_nullable",
					"numeric_precision",
					"numeric_precision_radix",
					"numeric_scale",
					"schema_name",
					"schema_oid",
					"table_name",
					"table_oid",
				}),
			)
		})

		It("should support insert queries", func() {
			q := NewDialect(DB).Insert(1, 0, "table", "", []string{"c1", "c2", "c2"}, []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("INSERT INTO `table` (`c1`,`c2`,`c2`) VALUES (DEFAULT,$1,$2) RETURNING `c1`,`c2`,`c2`"))
		})

		It("should support select queries", func() {
			q := NewDialect(DB).Select("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("SELECT `c1`,`c2`,`c2` FROM `table` WHERE `c1` = $1"))
		})

		It("should support update queries", func() {
			q := NewDialect(DB).Update("table", []string{"c1", "c2", "c2"}, []string{"c1"}, []string{"c1", "c2", "c2"})
			Expect(q).To(Equal("UPDATE `table` SET `c1` = $1, `c2` = $2, `c2` = $3 WHERE `c1` = $4 RETURNING `c1`,`c2`,`c2`"))
		})

		It("should support delete queries", func() {
			q := NewDialect(DB).Delete("table", []string{"c1", "c2", "c2"}, []string{"c1"})
			Expect(q).To(Equal("DELETE FROM `table` WHERE `c1` = $1"))
		})
	})
})
