package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func example2(q sqlx.Queryer, _default int, _genieqlQ int, _genieqlQuery int) ExampleScanner {
	const query = `SELECT * FROM example WHERE id = $1`
	return NewExampleScannerStatic(q.Query(query, _default, _genieqlQ, _genieqlQuery))
}
