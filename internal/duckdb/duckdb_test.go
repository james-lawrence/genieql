package duckdb_test

import (
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/drivers"
	. "github.com/james-lawrence/genieql/internal/duckdb"
	"github.com/james-lawrence/genieql/internal/testx"
	"github.com/stretchr/testify/require"
)

func TestDialect(t *testing.T) {
	driver := testx.MustT(genieql.LookupDriver(drivers.DuckDB))(t)

	t.Run("should return the columns in the query in sorted order", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		info, err := NewDialect(DB).ColumnInformationForQuery(
			driver,
			"SELECT database_name, schema_oid, is_nullable FROM duckdb_columns",
		)
		require.NoError(t, err)
		require.Equal(t, []string{"database_name", "is_nullable", "schema_oid"}, genieql.ColumnInfoSet(info).ColumnNames())
	})

	t.Run("should properly quote table names", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		info, err := NewDialect(DB).ColumnInformationForTable(driver, "example.foo.bar")
		require.NoError(t, err)
		require.Equal(t, []string{"id"}, genieql.ColumnInfoSet(info).ColumnNames())
	})

	t.Run("should return the columns in the table in the sorted order", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		info, err := NewDialect(DB).ColumnInformationForTable(driver, "duckdb_columns")
		require.NoError(t, err)
		require.Equal(t, []string{
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
		}, genieql.ColumnInfoSet(info).ColumnNames())
	})

	t.Run("should support insert queries", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		q := NewDialect(DB).Insert(1, 0, "table", "", []string{"c1", "c2", "c2"}, []string{"c1", "c2", "c2"}, []string{"c1"})
		require.Equal(t, "INSERT INTO \"table\" (\"c1\",\"c2\",\"c2\") VALUES (DEFAULT,$1,$2) RETURNING \"c1\",\"c2\",\"c2\"", q)
	})

	t.Run("should support select queries", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		q := NewDialect(DB).Select("table", []string{"c1", "c2", "c2"}, []string{"c1"})
		require.Equal(t, "SELECT \"c1\",\"c2\",\"c2\" FROM \"table\" WHERE \"c1\" = $1", q)
	})

	t.Run("should support update queries", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		q := NewDialect(DB).Update("table", []string{"c1", "c2", "c2"}, []string{"c1"}, []string{"c1", "c2", "c2"})
		require.Equal(t, "UPDATE \"table\" SET \"c1\" = $1, \"c2\" = $2, \"c2\" = $3 WHERE \"c1\" = $4 RETURNING \"c1\",\"c2\",\"c2\"", q)
	})

	t.Run("should support delete queries", func(t *testing.T) {
		TX = testx.MustT(DB.Begin())(t)
		t.Cleanup(func() { require.NoError(t, TX.Rollback()) })

		q := NewDialect(DB).Delete("table", []string{"c1", "c2", "c2"}, []string{"c1"})
		require.Equal(t, "DELETE FROM \"table\" WHERE \"c1\" = $1", q)
	})
}
