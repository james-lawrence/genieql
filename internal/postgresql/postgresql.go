package postgresql

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/types"
	"log"

	"github.com/davecgh/go-spew/spew"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/internal/debugx"
	"bitbucket.org/jatone/genieql/internal/postgresql/internal"
)

// Dialect constant representing the dialect name.
const Dialect = "postgres"

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

func (t dialectFactory) Connect(config genieql.Configuration) (_ genieql.Dialect, err error) {
	pcfg, err := pgx.ParseConfig(config.ConnectionURL)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to parse postgresql connection string: %s", config.ConnectionURL)
	}

	return dialectImplementation{db: stdlib.OpenDB(*pcfg)}, nil
}

type dialectImplementation struct {
	db *sql.DB
}

func (t dialectImplementation) Insert(n int, table string, columns, defaults []string) string {
	return Insert(n, table, columns, defaults)
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

func (t dialectImplementation) ColumnNameTransformer() genieql.ColumnTransformer {
	return genieql.NewColumnInfoNameTransformer(`"`)
}

func (t dialectImplementation) ColumnInformationForTable(d genieql.Driver, table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT a.attname, a.atttypid, NOT a.attnotnull AS nullable, COALESCE(a.attnum = ANY(i.indkey), 'f') AND COALESCE(i.indisprimary, 'f') AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) AND i.indisprimary = 't' WHERE a.attrelid = ($1)::regclass AND a.attnum > 0 AND a.attisdropped = 'f'`
	return columnInformation(d, t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(d genieql.Driver, query string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT a.attname, a.atttypid, 'f' AS nullable, 'f' AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid WHERE a.attrelid = ($1)::regclass AND a.attnum > 0`
	const table = "genieql_query_columns_table"

	tx, err := t.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failure to start transaction")
	}
	defer tx.Rollback()

	q := fmt.Sprintf("CREATE TABLE %s AS (%s)", table, query)
	if _, err = tx.Exec(q); err != nil {
		return nil, errors.Wrapf(err, "failure to execute %s", q)
	}

	return columnInformation(d, tx, columnInformationQuery, table)
}

func columnInformation(d genieql.Driver, q queryer, query, table string) ([]genieql.ColumnInfo, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []genieql.ColumnInfo
	)

	if rows, err = q.Query(query, table); err != nil {
		return nil, errors.Wrapf(err, "failed to query column information: %s, %s", query, table)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			columndef genieql.ColumnDefinition
			oid       int
			expr      ast.Expr
			primary   bool
			nullable  bool
			name      string
		)

		if err = rows.Scan(&name, &oid, &nullable, &primary); err != nil {
			return nil, errors.Wrapf(err, "error scanning column information for table (%s): %s", table, query)
		}

		if expr = internal.OIDToType(oid); expr == nil {
			log.Println("skipping column", name, "unknown type identifier", oid, "please open an issue")
			continue
		}

		if columndef, err = d.LookupType(types.ExprString(expr)); err != nil {
			log.Println("skipping column", name, "driver missing type", types.ExprString(expr), "please open an issue")
			continue
		}

		switch columndef.Native {
		case "[]byte":
			columndef.Nullable = false
		default:
			columndef.Nullable = nullable
		}

		columndef.PrimaryKey = primary

		debugx.Println("found column", name, types.ExprString(expr), spew.Sdump(columndef))

		columns = append(columns, genieql.ColumnInfo{
			Name:       name,
			Definition: columndef,
		})
	}

	columns = genieql.SortColumnInfo(columns)(genieql.ByName)

	return columns, errors.Wrap(rows.Err(), "error retrieving column information")
}
