package example

import "github.com/james-lawrence/genieql/internal/sqlx"

// NewBatchFunction2 creates a scanner that inserts a batch of
// records into the database.
func NewBatchFunction2(q sqlx.Queryer, i ...int) ExampleScanner {
	return &batchFunction2{
		q:         q,
		remaining: i,
	}
}

type batchFunction2 struct {
	q         sqlx.Queryer
	remaining []int
	scanner   ExampleScanner
}

func (t *batchFunction2) Scan(dst *int) error {
	return t.scanner.Scan(dst)
}

func (t *batchFunction2) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *batchFunction2) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *batchFunction2) Next() bool {
	var (
		advanced bool
	)

	if t.scanner != nil && t.scanner.Next() {
		return true
	}

	// advance to the next check
	if len(t.remaining) > 0 && t.Close() == nil {
		t.scanner, t.remaining, advanced = t.advance(t.q, t.remaining...)
		return advanced && t.scanner.Next()
	}

	return false
}

func (t *batchFunction2) advance(q sqlx.Queryer, i ...int) (ExampleScanner, []int, bool) {
	switch len(i) {
	case 0:
		return nil, []int(nil), false
	case 1:
		const query = `QUERY 1`

		return StaticExampleScanner(q.Query(query, i...)), []int(nil), true
	default:
		const query = `QUERY 2`

		return StaticExampleScanner(q.Query(query, i[:2]...)), i[2:], true
	}
}
