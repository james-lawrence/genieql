package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

// NewBatchFunction1 creates a scanner that inserts a batch of
// records into the database.
func NewBatchFunction1(q sqlx.Queryer, i ...int) ExampleScanner {
	return batchFunction1{
		q:         q,
		remaining: i,
	}
}

type batchFunction1 struct {
	q         sqlx.Queryer
	remaining []int
	scanner   ExampleScanner
}

func (t *batchFunction1) Scan(dst *int) error {
	return t.scanner.Scan(dst)
}

func (t *batchFunction1) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *batchFunction1) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *batchFunction1) Next() bool {
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

func (t *batchFunction1) advance(q sqlx.Queryer, i ...int) (ExampleScanner, []int, bool) {
	switch len(i) {
	case 0:
		return nil, []int(nil), false
	default:
		const query = `QUERY 1`

		return StaticExampleScanner(q.Query(query, i[:1]...)), i[1:], true
	}
}
