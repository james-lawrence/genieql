package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func queryFunction10(q sqlx.Queryer, query int) ExampleScanner {
	const __gqlquery__ = `SELECT * FROM example WHERE id = $1`
	return StaticExampleScanner(q.Query(__gqlquery__, query))
}
