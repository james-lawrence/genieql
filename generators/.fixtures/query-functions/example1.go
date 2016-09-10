package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction1(q sqlx.Queryer, arg1 int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return StaticExampleScanner(q.Query(query, arg1))
}
