package duckdb_test

import (
	"testing"

	. "github.com/james-lawrence/genieql/internal/duckdb"
	"github.com/stretchr/testify/require"
)

func TestQueries(t *testing.T) {
	t.Run("Insert", func(t *testing.T) {
		cases := []struct {
			name     string
			n        int
			table    string
			conflict string
			columns  []string
			defaults []string
			expected string
		}{
			{
				name:     "example 1",
				n:        1,
				table:    "MyTable1",
				columns:  []string{"col1", "col2", "col3"},
				defaults: []string{},
				expected: "INSERT INTO \"MyTable1\" (\"col1\",\"col2\",\"col3\") VALUES ($1,$2,$3) RETURNING \"col1\",\"col2\",\"col3\"",
			},
			{
				name:     "example 2",
				n:        1,
				table:    "MyTable2",
				columns:  []string{"col1", "col2", "col3", "col4"},
				defaults: []string{"col4"},
				expected: "INSERT INTO \"MyTable2\" (\"col1\",\"col2\",\"col3\",\"col4\") VALUES ($1,$2,$3,DEFAULT) RETURNING \"col1\",\"col2\",\"col3\",\"col4\"",
			},
			{
				name:     "example 3",
				n:        1,
				table:    "MyTable2",
				columns:  []string{"col1", "col2", "col3", "col4"},
				defaults: []string{"col1", "col3"},
				expected: "INSERT INTO \"MyTable2\" (\"col1\",\"col2\",\"col3\",\"col4\") VALUES (DEFAULT,$1,DEFAULT,$2) RETURNING \"col1\",\"col2\",\"col3\",\"col4\"",
			},
			{
				name:     "example 4",
				n:        3,
				table:    "MyTable2",
				columns:  []string{"col1", "col2", "col3", "col4"},
				defaults: []string{"col1", "col3"},
				expected: "INSERT INTO \"MyTable2\" (\"col1\",\"col2\",\"col3\",\"col4\") VALUES (DEFAULT,$1,DEFAULT,$2),(DEFAULT,$3,DEFAULT,$4),(DEFAULT,$5,DEFAULT,$6) RETURNING \"col1\",\"col2\",\"col3\",\"col4\"",
			},
			{
				name:     "example 5",
				n:        3,
				table:    "MyTable1",
				columns:  []string{"col1", "col2", "col3"},
				defaults: []string{},
				expected: "INSERT INTO \"MyTable1\" (\"col1\",\"col2\",\"col3\") VALUES ($1,$2,$3),($4,$5,$6),($7,$8,$9) RETURNING \"col1\",\"col2\",\"col3\"",
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				require.Equal(t, c.expected, Insert(c.n, 0, c.table, c.conflict, c.columns, c.columns, c.defaults))
			})
		}
	})

	t.Run("Select", func(t *testing.T) {
		cases := []struct {
			name       string
			table      string
			columns    []string
			predicates []string
			expected   string
		}{
			{
				name:       "example 1",
				table:      "MyTable1",
				columns:    []string{"col1", "col2", "col3"},
				predicates: []string{"col1"},
				expected:   "SELECT \"col1\",\"col2\",\"col3\" FROM \"MyTable1\" WHERE \"col1\" = $1",
			},
			{
				name:       "example 2",
				table:      "MyTable2",
				columns:    []string{"col1", "col2", "col3", "col4"},
				predicates: []string{"col1", "col2"},
				expected:   "SELECT \"col1\",\"col2\",\"col3\",\"col4\" FROM \"MyTable2\" WHERE \"col1\" = $1 AND \"col2\" = $2",
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				require.Equal(t, c.expected, Select(c.table, c.columns, c.predicates))
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		cases := []struct {
			name       string
			table      string
			columns    []string
			predicates []string
			expected   string
		}{
			{
				name:       "example 1",
				table:      "MyTable1",
				columns:    []string{"col1", "col2", "col3"},
				predicates: []string{"col1"},
				expected:   "UPDATE \"MyTable1\" SET \"col1\" = $1, \"col2\" = $2, \"col3\" = $3 WHERE \"col1\" = $4 RETURNING \"col1\",\"col2\",\"col3\"",
			},
			{
				name:       "example 2",
				table:      "MyTable2",
				columns:    []string{"col1", "col2", "col3", "col4"},
				predicates: []string{"col1", "col2"},
				expected:   "UPDATE \"MyTable2\" SET \"col1\" = $1, \"col2\" = $2, \"col3\" = $3, \"col4\" = $4 WHERE \"col1\" = $5 AND \"col2\" = $6 RETURNING \"col1\",\"col2\",\"col3\",\"col4\"",
			},
			{
				name:       "example 3",
				table:      "MyTable2",
				columns:    []string{"col1", "col2", "col3", "col4"},
				predicates: []string{},
				expected:   "UPDATE \"MyTable2\" SET \"col1\" = $1, \"col2\" = $2, \"col3\" = $3, \"col4\" = $4 WHERE TRUE RETURNING \"col1\",\"col2\",\"col3\",\"col4\"",
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				require.Equal(t, c.expected, Update(c.table, c.columns, c.predicates, c.columns))
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		cases := []struct {
			name       string
			table      string
			columns    []string
			predicates []string
			expected   string
		}{
			{
				name:       "example 1",
				table:      "MyTable1",
				columns:    []string{"col1", "col2", "col3"},
				predicates: []string{"col1"},
				expected:   "DELETE FROM \"MyTable1\" WHERE \"col1\" = $1",
			},
			{
				name:       "example 2",
				table:      "MyTable2",
				columns:    []string{"col1", "col2", "col3", "col4"},
				predicates: []string{"col1", "col2"},
				expected:   "DELETE FROM \"MyTable2\" WHERE \"col1\" = $1 AND \"col2\" = $2",
			},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				require.Equal(t, c.expected, Delete(c.table, c.columns, c.predicates))
			})
		}
	})
}
