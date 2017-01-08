package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction3(q sqlx.Queryer, query string, arg1 int) ExampleRowScanner {
	return StaticExampleRowScanner(q.QueryRow(query, arg1))
}
