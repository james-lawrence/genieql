package sqlite3

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"golang.org/x/text/transform"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/columninfo"
	"github.com/james-lawrence/genieql/dialects"
	"github.com/james-lawrence/genieql/internal/debugx"
	"github.com/james-lawrence/genieql/internal/errorsx"
)

// Dialect constant representing the dialect name.
const Dialect = "sqlite3"

// NewDialect creates a sqlite Dialect from the queryer
func NewDialect(q *sql.DB) genieql.Dialect {
	return dialectImplementation{db: q}
}

func init() {
	errorsx.MaybePanic(dialects.Register(Dialect, dialectFactory{}))
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

func (t dialectImplementation) Insert(n int, offset int, table, conflict string, columns, projection, defaults []string) string {
	return Insert(n, offset, table, conflict, columns, defaults)
}

func (t dialectImplementation) Select(table string, columns, predicates []string) string {
	return Select(table, columns, predicates)
}

func (t dialectImplementation) Update(table string, columns, predicates, returning []string) string {
	return Update(table, columns, predicates)
}

func (t dialectImplementation) Delete(table string, columns, predicates []string) string {
	return Delete(table, columns, predicates)
}

func (t dialectImplementation) ColumnValueTransformer() genieql.ColumnTransformer {
	// TODO
	return columninfo.NewNameTransformer(transform.Nop)
}

func (t dialectImplementation) ColumnNameTransformer(opts ...transform.Transformer) genieql.ColumnTransformer {
	// TODO
	return columninfo.NewNameTransformer(transform.Nop, transform.Chain(opts...))
}

func (t dialectImplementation) ColumnInformationForTable(d genieql.Driver, table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `PRAGMA table_info('%s')`
	return columnInformation(d, t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(d genieql.Driver, query string) ([]genieql.ColumnInfo, error) {
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

	return columnInformation(d, tx, columnInformationQuery, table)
}

func (t dialectImplementation) QuotedString(s string) string {
	return s
}

func columnInformation(d genieql.Driver, q queryer, query, table string) ([]genieql.ColumnInfo, error) {
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
			columndef  genieql.ColumnDefinition
			id         int         // ignored.
			name       string      // column name.
			expr       string      // column type.
			nullable   int         // nullable.
			defaultVal interface{} // ignored.
			primary    int         // part of the primary key.
		)

		if err = rows.Scan(&id, &name, &expr, &nullable, &defaultVal, &primary); err != nil {
			return nil, errors.Wrapf(err, "error scanning column information for table (%s): %s", table, query)
		}

		if columndef, err = d.LookupType(expr); err != nil {
			log.Println("skipping column", name, "driver missing type", expr, "please open an issue")
			continue
		}

		columndef.Nullable = isNullable(nullable)
		columndef.PrimaryKey = isPrimary(primary)

		debugx.Println("found column", name, expr, spew.Sdump(columndef))

		columns = append(columns, genieql.ColumnInfo{
			Name:       name,
			Definition: columndef,
		})
	}

	columns = genieql.SortColumnInfo(columns)(genieql.ByName)

	return columns, errors.Wrap(rows.Err(), "error retrieving column information")
}

func isNullable(i int) bool {
	return i == 0
}

func isPrimary(i int) bool {
	return i == 0
}
