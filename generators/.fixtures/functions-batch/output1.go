package example

import (
	"errors"

	"bitbucket.org/jatone/genieql/sqlx"
)

func batchFunction1(q sqlx.Queryer, i ...int) (ExampleScanner, []int) {
	switch len(i) {
	case 0:
		return StaticExampleScanner(nil, errors.New("need at least 1 value to execute a batch query")), i
	default:
		const query = `QUERY 1`

		return StaticExampleScanner(q.Query(query, i[:1]...)), i[1:]
	}
}
