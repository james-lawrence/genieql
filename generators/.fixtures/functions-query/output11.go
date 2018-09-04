package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func queryFunction6(q sqlx.Queryer, query string, _type int, _func int) ExampleRowScanner {
	return StaticExampleRowScanner(q.QueryRow(query, _type, _func))
}
