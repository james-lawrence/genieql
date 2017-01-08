package example

import (
	"errors"

	"bitbucket.org/jatone/genieql/sqlx"
)

func batchFunction3(q sqlx.Queryer, i ...int) (ExampleScanner, []int) {
	switch len(i) {
	case 0:
		return StaticExampleScanner(nil, errors.New("need at least 1 value to execute a batch query")), i
	case 1:
		const query = `QUERY 1`
		exploder := func(i ...int) (r [3]interface{}) {
			for idx, v := range i[:1] {
				r[idx*3+0], r[idx*3+1], r[idx*3+2] = v.A, v.B, v.C
			}
			return
		}
		return StaticExampleScanner(q.Query(query, exploder(i...)[:]...)), i[len(i)-1:]
	case 2:
		const query = `QUERY 2`
		exploder := func(i ...int) (r [6]interface{}) {
			for idx, v := range i[:2] {
				r[idx*3+0], r[idx*3+1], r[idx*3+2] = v.A, v.B, v.C
			}
			return
		}
		return StaticExampleScanner(q.Query(query, exploder(i...)[:]...)), i[len(i)-1:]
	default:
		const query = `QUERY 3`
		exploder := func(i ...int) (r [9]interface{}) {
			for idx, v := range i[:3] {
				r[idx*3+0], r[idx*3+1], r[idx*3+2] = v.A, v.B, v.C
			}
			return
		}
		return StaticExampleScanner(q.Query(query, exploder(i[:3]...)[:]...)), i[3:]
	}
}
