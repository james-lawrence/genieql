package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func queryFunction1(q sqlx.Queryer, arg1 int) ExampleScanner {
	var query = mypkg.HelloWorld
	return StaticExampleScanner(q.Query(query, arg1))
}
