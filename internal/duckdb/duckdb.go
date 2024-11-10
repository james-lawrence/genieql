package duckdb

import (
	"database/sql"
	"fmt"
	"go/types"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"golang.org/x/text/transform"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/columninfo"
	"github.com/james-lawrence/genieql/dialects"
	"github.com/james-lawrence/genieql/internal/debugx"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/md5x"
	"github.com/james-lawrence/genieql/internal/transformx"
)

// Dialect constant representing the dialect name.
const Dialect = "duckdb"

// NewDialect creates a duckdb Dialect from the queryer
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

func (t dialectFactory) Connect(config genieql.Configuration) (_ genieql.Dialect, err error) {
	db, err := sql.Open(Dialect, config.ConnectionURL)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to DuckDB: %s", config.ConnectionURL)
	}
	return dialectImplementation{db: db}, nil
}

type dialectImplementation struct {
	db *sql.DB
}

func (t dialectImplementation) Insert(n int, offset int, table, conflict string, columns, projection, defaults []string) string {
	return Insert(n, offset, table, conflict, columns, projection, defaults)
}

func (t dialectImplementation) Select(table string, columns, predicates []string) string {
	return Select(table, columns, predicates)
}

func (t dialectImplementation) Update(table string, columns, predicates, returning []string) string {
	return Update(table, columns, predicates, returning)
}

func (t dialectImplementation) Delete(table string, columns, predicates []string) string {
	return Delete(table, columns, predicates)
}

func (t dialectImplementation) ColumnValueTransformer() genieql.ColumnTransformer {
	return &columnValueTransformer{}
}

func (t dialectImplementation) ColumnNameTransformer(transforms ...transform.Transformer) genieql.ColumnTransformer {
	return columninfo.NewNameTransformer(
		transformx.Wrap("\""),
		transform.Chain(transforms...),
	)
}

func (t dialectImplementation) ColumnInformationForTable(d genieql.Driver, table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `DESCRIBE %s`
	return columnInformation(d, t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(d genieql.Driver, query string) ([]genieql.ColumnInfo, error) {
	var (
		tx  *sql.Tx
		err error
	)
	const columnInformationQuery = `DESCRIBE %s`

	uid := md5x.String(query)
	table := fmt.Sprintf("gql_%s", uid)

	tx, err = t.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failure to start transaction")
	}
	defer tx.Rollback()

	q := fmt.Sprintf("CREATE TEMPORARY TABLE %s AS (%s LIMIT 1)", table, query)
	if _, err = tx.Exec(q); err != nil {
		return nil, errors.Wrapf(err, "failure to execute %s", q)
	}

	return columnInformation(d, tx, columnInformationQuery, table)
}

func (t dialectImplementation) QuotedString(s string) string {
	return quotedString(s)
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
	defer rows.Close()

	for rows.Next() {
		var (
			columndef genieql.ColumnDefinition
			dataType  string
			nullable  string
			name      string
		)

		if err = rows.Scan(&name, &dataType, &nullable); err != nil {
			return nil, errors.Wrapf(err, "error scanning column information for table (%s): %s", table, query)
		}

		expr := astutil.Expr(dataType)
		if columndef, err = d.LookupType(types.ExprString(expr)); err != nil {
			log.Println("skipping column", name, "driver missing type", types.ExprString(expr), "please open an issue")
			continue
		}

		columndef.Nullable = (nullable == "YES")
		debugx.Println("found column", name, types.ExprString(expr), spew.Sdump(columndef))

		columns = append(columns, genieql.ColumnInfo{
			Name:       name,
			Definition: columndef,
		})
	}

	columns = genieql.SortColumnInfo(columns)(genieql.ByName)

	return columns, errors.Wrap(rows.Err(), "error retrieving column information")
}
