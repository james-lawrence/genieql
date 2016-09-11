package sqlite3

import (
	"database/sql"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

// Dialect constant representing the dialect name.
const Dialect = "sqlite3"

// NewDialect creates a postgresql Dialect from the queryer
func NewDialect(q *sql.DB) genieql.Dialect {
	return dialectImplementation{db: q}
}

func init() {
	maybePanic := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	maybePanic(genieql.RegisterDialect(Dialect, dialectFactory{}))
}

type queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}

type dialectFactory struct{}

func (t dialectFactory) Connect(config genieql.Configuration) (genieql.Dialect, error) {
	var (
		err error
		db  *sql.DB
	)

	db, err = sql.Open(config.Dialect, config.ConnectionURL)
	return dialectImplementation{db: db}, errors.Wrap(err, "failure to open database connection")
}

type dialectImplementation struct {
	db *sql.DB
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

func (t dialectImplementation) ColumnInformationForTable(table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT a.attname, a.atttypid, NOT a.attnotnull AS nullable, COALESCE(a.attnum = ANY(i.indkey), 'f') AND COALESCE(i.indisprimary, 'f') AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) AND i.indisprimary = 't' WHERE a.attrelid = ($1)::regclass AND a.attnum > 0 AND a.attisdropped = 'f'`
	return columnInformation(t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(query string) ([]genieql.ColumnInfo, error) {
	return columnInformation(t.db, query, "table")
	// const columnInformationQuery = `SELECT a.attname, a.atttypid, 'f' AS nullable, 'f' AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid WHERE a.attrelid = ($1)::regclass AND a.attnum > 0`
	// const table = "genieql_query_columns_table"
	//
	// tx, err := t.db.Begin()
	// if err != nil {
	// 	return nil, errors.Wrap(err, "failure to start transaction")
	// }
	// defer tx.Rollback()
	//
	// q := fmt.Sprintf("CREATE TABLE %s AS (%s)", table, query)
	// if _, err = tx.Exec(q); err != nil {
	// 	return nil, errors.Wrapf(err, "failure to execute %s", q)
	// }
	//
	// return columnInformation(tx, columnInformationQuery, table)
}

func columnInformation(q queryer, query, table string) ([]genieql.ColumnInfo, error) {
	return []genieql.ColumnInfo(nil), nil
}
