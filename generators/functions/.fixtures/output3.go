package example

import "github.com/james-lawrence/genieql/internal/sqlx"

func example3(q sqlx.Queryer, arg1 int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStaticRow(q.QueryRow(query, arg1))
}
