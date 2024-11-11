package example

import (
	"context"

	"github.com/james-lawrence/genieql/internal/sqlx"
)

func example5(ctx context.Context, q sqlx.Queryer, arg1 int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStaticRow(q.QueryRowContext(ctx, query, arg1))
}
