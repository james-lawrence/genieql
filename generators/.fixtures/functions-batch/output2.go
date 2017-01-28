package example

import (
	"errors"

	"bitbucket.org/jatone/genieql/internal/sqlx"
)

func batchFunction2(q sqlx.Queryer, i ...int) (ExampleScanner, []int) {
	switch len(i) {
	case 0:
		return StaticExampleScanner(nil, errors.New("need at least 1 value to execute a batch query")), i
	case 1:
		const query = `QUERY 1`

		return StaticExampleScanner(q.Query(query, i...)), i[len(i)-1:]
	default:
		const query = `QUERY 2`

		return StaticExampleScanner(q.Query(query, i[:2]...)), i[2:]
	}
}
