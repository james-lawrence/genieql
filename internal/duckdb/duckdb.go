package duckdb

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/types"
	"log"

	"github.com/davecgh/go-spew/spew"
	"golang.org/x/text/transform"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/astutil"
	"github.com/james-lawrence/genieql/columninfo"
	"github.com/james-lawrence/genieql/dialects"
	"github.com/james-lawrence/genieql/internal/debugx"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/langx"
	"github.com/james-lawrence/genieql/internal/md5x"
	"github.com/james-lawrence/genieql/internal/transformx"
)

// Dialect constant representing the dialect name.
const Dialect = "duckdb"

// NewDialect creates a duckdb Dialect from the queryer
func NewDialect(q *sql.DB) genieql.Dialect {
	return DialectFn{db: q}
}

func init() {
	errorsx.MaybePanic(dialects.Register(Dialect, dialectFactory{}))
}

type queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
}

type dialectFactory struct{}

func (t dialectFactory) Connect(config genieql.Configuration) (_ genieql.Dialect, err error) {
	db, err := sql.Open(Dialect, config.Database)
	if err != nil {
		return nil, errorsx.Wrapf(err, "unable to connect to DuckDB: %s", config.Database)
	}
	return DialectFn{db: db}, nil
}

type DialectFn struct {
	db *sql.DB
}

func (t DialectFn) Insert(n int, offset int, table, conflict string, columns, projection, defaults []string) string {
	return Insert(n, offset, table, conflict, columns, projection, defaults)
}

func (t DialectFn) Select(table string, columns, predicates []string) string {
	return Select(table, columns, predicates)
}

func (t DialectFn) Update(table string, columns, predicates, returning []string) string {
	return Update(table, columns, predicates, returning)
}

func (t DialectFn) Delete(table string, columns, predicates []string) string {
	return Delete(table, columns, predicates)
}

func (t DialectFn) ColumnValueTransformer() genieql.ColumnTransformer {
	return &columnValueTransformer{}
}

func (t DialectFn) ColumnNameTransformer(transforms ...transform.Transformer) genieql.ColumnTransformer {
	return columninfo.NewNameTransformer(
		transformx.Wrap("\""),
		transform.Chain(transforms...),
	)
}

func (t DialectFn) ColumnInformationForTable(d genieql.Driver, table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `DESCRIBE %s`
	return columnInformation(d, t.db, columnInformationQuery, table)
}

func (t DialectFn) ColumnInformationForQuery(d genieql.Driver, query string) (_ []genieql.ColumnInfo, err error) {
	const columnInformationQuery = `DESCRIBE %s`
	var (
		tx *sql.Tx
	)
	table := fmt.Sprintf("gql_%s", md5x.Hex(query))

	tx, err = t.db.Begin()
	if err != nil {
		return nil, errorsx.Wrap(err, "failure to start transaction")
	}
	defer func() {
		err = errorsx.Compact(err, tx.Rollback())
	}()

	q := fmt.Sprintf("CREATE TEMPORARY TABLE %s AS (%s LIMIT 1)", table, query)
	if _, err = tx.Exec(q); err != nil {
		return nil, errorsx.Wrapf(err, "failure to execute %s", q)
	}

	return columnInformation(d, tx, columnInformationQuery, table)
}

func (t DialectFn) QuotedString(s string) string {
	return quotedString(s)
}

func (t DialectFn) SQLDB(cb func(db *sql.DB)) {
	cb(t.db)
}

func columnInformation(d genieql.Driver, q queryer, query, table string) ([]genieql.ColumnInfo, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []genieql.ColumnInfo
	)

	if rows, err = q.Query(fmt.Sprintf(query, table)); err != nil {
		return nil, errorsx.Wrapf(err, "failed to query column information: %s, %s", query, table)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			columndef genieql.ColumnDefinition
			dataType  string
			nullable  string
			name      string
			key       *string = new(string)
			defaulted *string = new(string)
			extra     *string = new(string)
		)

		if err = rows.Scan(&name, &dataType, &nullable, &key, &defaulted, &extra); err != nil {
			return nil, errorsx.Wrapf(err, "error scanning column information for table (%s): %s", table, query)
		}

		expr := totypeexpr(dataType)
		if expr == nil {
			log.Println("skipping column", name, "driver missing type", dataType, "please open an issue")
			continue
		}

		if columndef, err = d.LookupType(types.ExprString(expr)); err != nil {
			log.Println(err)
			log.Println("skipping column", name, err, "please open an issue")
			continue
		}

		columndef.Nullable = (nullable == "YES")
		columndef.PrimaryKey = (langx.Autoderef(key) == "PRI")
		debugx.Println("found column", name, types.ExprString(expr), spew.Sdump(columndef))

		columns = append(columns, genieql.ColumnInfo{
			Name:       name,
			Definition: columndef,
		})
	}

	columns = genieql.SortColumnInfo(columns)(genieql.ByName)

	return columns, errorsx.Wrap(rows.Err(), "error retrieving column information")
}

// OIDToType maps object id to golang types.
func totypeexpr(id string) ast.Expr {
	// if strings.HasPrefix(id, "DECIMAL") {
	// 	return astutil.Expr("sql.NullFloat64")
	// }

	switch id {
	case "FLOAT":
		return astutil.Expr("FLOAT")
	case "DOUBLE":
		return astutil.Expr("DOUBLE")
	case "VARCHAR":
		return astutil.Expr("VARCHAR")
	// case "VARCHAR[]":
	// 	return astutil.Expr("VARCHARARRAY")
	case "BOOLEAN":
		return astutil.Expr("BOOLEAN")
	case "BIGINT":
		return astutil.Expr("BIGINT")
	case "UINTEGER":
		return astutil.Expr("UINTEGER")
	case "UBIGINT":
		return astutil.Expr("UBIGINT")
	case "USMALLINT":
		return astutil.Expr("USMALLINT")
	case "INTEGER":
		return astutil.Expr("INTEGER")
	case "SMALLINT":
		return astutil.Expr("SMALLINT")
	// case "SMALLINT[]":
	// 	return astutil.Expr("SMALLINTARRAY")
	case "TIMESTAMPZ", "TIMESTAMP WITH TIME ZONE":
		return astutil.Expr("TIMESTAMPZ")
	// case "INTERVAL":
	// 	return astutil.Expr("INTERVAL")
	case "BINARY":
		return astutil.Expr("BINARY")
	case "BLOB":
		return astutil.Expr("BLOB")
	case "INET":
		return astutil.Expr("INET")
	case "UUID":
		return astutil.Expr("UUID")
	default:
		return nil
	}
}
