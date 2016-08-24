package postgresql

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/types"
	"log"

	"github.com/jackc/pgx"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// Dialect constant representing the dialect name.
const Dialect = "postgres"

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

	log.Printf("connection %s\n", config.ConnectionURL)

	db, err = sql.Open(config.Dialect, config.ConnectionURL)
	return dialectImplementation{db: db}, err
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

func (t dialectImplementation) ColumnInformation(table string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT a.attname, a.atttypid, NOT a.attnotnull AS nullable, COALESCE(a.attnum = ANY(i.indkey), 'f') AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid WHERE a.attrelid = ($1)::regclass AND a.attnum > 0`
	return t.columnInformation(t.db, columnInformationQuery, table)
}

func (t dialectImplementation) ColumnInformationForQuery(query string) ([]genieql.ColumnInfo, error) {
	const columnInformationQuery = `SELECT a.attname, a.atttypid, 'f' AS nullable, 'f' AS isprimary FROM pg_index i RIGHT OUTER JOIN pg_attribute a ON a.attrelid = i.indrelid WHERE a.attrelid = ($1)::regclass AND a.attnum > 0`
	const table = "genieql_query_columns"

	tx, err := t.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err = tx.Exec(fmt.Sprintf("CREATE TABLE %s AS (%s)", table, query)); err != nil {
		return nil, err
	}

	return t.columnInformation(tx, columnInformationQuery, table)
}

func (t dialectImplementation) columnInformation(q queryer, query, table string) ([]genieql.ColumnInfo, error) {
	var (
		err     error
		rows    *sql.Rows
		columns []genieql.ColumnInfo
	)

	if rows, err = q.Query(query, table); err != nil {
		return nil, err
	}

	for rows.Next() {
		var (
			info genieql.ColumnInfo
			oid  int
			expr ast.Expr
		)

		if err = rows.Scan(&info.Name, &oid, &info.Nullable, &info.PrimaryKey); err != nil {
			return nil, err
		}

		if expr = oidToType(oid); expr == nil {
			log.Println("skipping column", info.Name, "unknown type identifier", oid, "please open and issue")
			continue
		}

		info.Type = types.ExprString(expr)

		columns = append(columns, info)
	}

	return columns, rows.Err()
}

// This is driver dependent, will have to abstract away.
func oidToType(oid int) ast.Expr {
	switch oid {
	case pgx.BoolOid:
		return astutil.Expr("bool")
	case pgx.UuidOid:
		return astutil.Expr("string")
	case pgx.TimestampTzOid, pgx.TimestampOid, pgx.DateOid:
		return astutil.Expr("time.Time")
	case pgx.Int2Oid, pgx.Int4Oid, pgx.Int8Oid:
		return astutil.Expr("int")
	case pgx.TextOid, pgx.VarcharOid:
		return astutil.Expr("string")
	case pgx.ByteaOid:
		return astutil.Expr("[]byte")
	case pgx.Float4Oid:
		return astutil.Expr("float32")
	case pgx.Float8Oid:
		return astutil.Expr("float64")
	case pgx.InetOid:
		return astutil.Expr("string")
	default:
		return nil
	}
}
