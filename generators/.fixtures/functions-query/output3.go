package example

import (
	"database/sql"

	"github.com/james-lawrence/genieql/internal/sqlx"
)

// queryFunction3 generated by genieql
func queryFunction3(q sqlx.Queryer, query string, arg1 int) ExampleRowScanner {
	var (
		c0 sql.NullInt64
	)

	c0.Valid = true
	c0.Int64 = int64(arg1)

	return StaticExampleRowScanner(q.QueryRow(query, c0))
}
