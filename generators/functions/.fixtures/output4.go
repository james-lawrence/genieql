package example

import (
	"context"

	"bitbucket.org/jatone/genieql/internal/sqlx"
)

func example4(ctx context.Context, q sqlx.Queryer, arg1 int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStaticRow(q.QueryRowContext(ctx, query, arg1))
}
