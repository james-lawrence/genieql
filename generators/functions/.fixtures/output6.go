package example

import (
	"context"

	"bitbucket.org/jatone/genieql/internal/sqlx"
)

func example6(ctx context.Context, q sqlx.Queryer, arg1 int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStatic(q.QueryContext(ctx, query, arg1))
}
