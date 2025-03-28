package example

import (
	"database/sql"

	"github.com/james-lawrence/genieql/internal/sqlx"
)

// queryFunction6 generated by genieql
func queryFunction6(q sqlx.Queryer, arg1 int) ExampleScanner {
	var query = HelloWorld
	var (
		c0 sql.NullInt64
	)

	c0.Valid = true
	c0.Int64 = int64(arg1)

	return StaticExampleScanner(q.Query(query, c0))
}
