package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func queryFunction3(q sqlx.Queryer, query string, arg1 int) ExampleScanner {
	return StaticExampleScanner(q.QueryRow(query, arg1))
}
