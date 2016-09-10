package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction4(q sqlx.Queryer, query string, params ...interface{}) ExampleScanner {
	return StaticExampleScanner(q.QueryRow(query, params...))
}
