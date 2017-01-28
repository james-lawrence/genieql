package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func queryFunction8(q sqlx.Queryer, query string, arg1 StructA) ExampleScanner {
	return StaticExampleScanner(q.Query(query, arg1.A, arg1.B, arg1.C, arg1.D, arg1.E, arg1.F))
}
