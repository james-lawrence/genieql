package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction4(q sqlx.Queryer, query string, params ...interface{}) ExampleRowScanner {
	return StaticExampleRowScanner(q.QueryRow(query, params...))
}