package example

import (
	"net"

	"github.com/jackc/pgtype"
	"github.com/james-lawrence/genieql/internal/sqlx"
)

// queryFunction2 generated by genieql
func queryFunction2(q sqlx.Queryer, query string, a net.IPNet) ExampleRowScanner {
	var (
		c0 pgtype.CIDR
	)

	if err := c0.Set(a); err != nil {
		return StaticExampleRowScanner(nil).Err(err)
	}

	return StaticExampleRowScanner(q.QueryRow(query, c0))
}
