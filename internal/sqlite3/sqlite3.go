package sqlite3

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

// Dialect constant representing the dialect name.
const Dialect = "sqlite3"

// NewDialect creates a sqlite Dialect from the queryer
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

func (t dialectImplementation) ColumnValueTransformer() genieql.ColumnTransformer {
	// TODO
	return genieql.ColumnInfoNameTransformer{}
}

func (t dialectImplementation) ColumnInformationForTable(table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `PRAGMA table_info('%s')`
	return columnInformation(t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(query string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `PRAGMA table_info('%s')`
	const table = "genieql_query_columns_table"

	tx, err := t.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failure to start transaction")
	}
	defer tx.Rollback()

	q := fmt.Sprintf("CREATE TABLE %s AS %s", table, query)
	if _, err = tx.Exec(q); err != nil {
		return nil, errors.Wrapf(err, "failure to execute %s", q)
	}

	return columnInformation(tx, columnInformationQuery, table)
}

func columnInformation(q queryer, query, table string) ([]genieql.ColumnInfo, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []genieql.ColumnInfo
	)

	if rows, err = q.Query(fmt.Sprintf(query, table)); err != nil {
		return nil, errors.Wrapf(err, "failed to query column information: %s, %s", query, table)
	}

	for rows.Next() {
		var (
			id         int         // ignored.
			cName      string      // column name.
			cType      string      // column type.
			nullable   int         // nullable.
			defaultVal interface{} // ignored.
			primary    int         // part of the primary key.
		)

		if err = rows.Scan(&id, &cName, &cType, &nullable, &defaultVal, &primary); err != nil {
			return nil, errors.Wrapf(err, "error scanning column information for table (%s): %s", table, query)
		}

		columns = append(columns, genieql.ColumnInfo{
			Name:       cName,
			Nullable:   isNullable(nullable),
			PrimaryKey: isPrimary(primary),
			Type:       cType, // TODO mapping function from sqlite to golang.
		})
	}

	columns = genieql.SortColumnInfo(columns)(genieql.ByName)

	return columns, errors.Wrap(rows.Err(), "error retrieving column information")
}

func isNullable(i int) bool {
	if i == 0 {
		return true
	}
	return false
}

func isPrimary(i int) bool {
	if i == 0 {
		return false
	}
	return true
}
