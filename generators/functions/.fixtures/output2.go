package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func example2(q sqlx.Queryer, _default int, _genieql_q int, _genieql_query int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStatic(q.Query(query, _default, _genieql_q, _genieql_query))
}
