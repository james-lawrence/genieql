package postgresql

import (
	"fmt"

	"bitbucket.org/jatone/genieql"
)

// Dialect constant representing the dialect name.
const Dialect = "postgres"

func init() {
	maybePanic := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	maybePanic(genieql.RegisterDialect(Dialect, dialectImplementation{}))
}

const primaryKeyQuery = `SELECT a.attname	FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid
AND a.attnum = ANY(i.indkey)
WHERE  i.indrelid = '%s'::regclass
AND    i.indisprimary`

const columnQuery = `SELECT * FROM %s LIMIT 1`

type dialectImplementation struct {
}

func (t dialectImplementation) Insert(table string, columns, defaults []string) string {
	return Insert(table, columns, defaults)
}

func (t dialectImplementation) Select(table string, columns, predicates []string) string {
	return Select(table, columns, predicates)
}

func (t dialectImplementation) Update(table string, columns, predicates []string) string {
	return Update(table, columns, predicates)
}

func (t dialectImplementation) Delete(table string, columns, predicates []string) string {
	return Delete(table, columns, predicates)
}

func (t dialectImplementation) ColumnQuery(table string) string {
	return fmt.Sprintf(columnQuery, table)
}

func (t dialectImplementation) PrimaryKeyQuery(table string) string {
	return fmt.Sprintf(primaryKeyQuery, table)
}
