package postgresql

import (
	"database/sql"
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

type queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}

type dialectImplementation struct{}

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

func (t dialectImplementation) ColumnInformation(db *sql.DB, table string) ([]genieql.ColumnInfo, error) {
	return t.columnInformation(db, table)
}

func (t dialectImplementation) ColumnInformationForQuery(db *sql.DB, query string) ([]genieql.ColumnInfo, error) {
	const table = "genieql_query_columns"

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(fmt.Sprintf("CREATE TABLE %s AS %s", table, query)); err != nil {
		return nil, err
	}

	return t.columnInformation(tx, table)
}

func (t dialectImplementation) columnInformation(q queryer, table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT column_name, data_type, CASE WHEN (is_nullable = 'NO') THEN 'f' ELSE 't' END AS is_nullable FROM information_schema.columns WHERE table_name = $1`
	const primaryKeyQuery = `SELECT a.attname FROM pg_index i
	JOIN pg_attribute a ON a.attrelid = i.indrelid
	AND a.attnum = ANY(i.indkey)
	WHERE  i.indrelid = ($1)::regclass
	AND    i.indisprimary`

	var (
		err     error
		rows    *sql.Rows
		columns []genieql.ColumnInfo
	)

	if rows, err = q.Query(columnInformationQuery, table); err != nil {
		return nil, err
	}

	for rows.Next() {
		var info genieql.ColumnInfo

		if err = rows.Scan(&info.Name, &info.Type, &info.Nullable); err != nil {
			return nil, err
		}

		columns = append(columns, info)
	}

	if rows.Err() != nil {
		return nil, err
	}

	if rows, err = q.Query(primaryKeyQuery, table); err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			name string
		)
		columnIdx := -1

		if err = rows.Scan(&name); err != nil {
			return nil, err
		}

		for idx, info := range columns {
			if info.Name == name {
				columnIdx = idx
			}
		}

		if columnIdx != -1 {
			info := columns[columnIdx]
			info.PrimaryKey = true
			columns[columnIdx] = info
		}
	}

	return columns, nil
}
