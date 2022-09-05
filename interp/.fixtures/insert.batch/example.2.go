package example

import "bitbucket.org/jatone/genieql/internal/sqlx"

func NewExample1BatchInsertFunction(q sqlx.Queryer, p ...Example1) Example1Scanner {
	return &example1BatchInsertFunction{
		q:         q,
		remaining: p,
	}
}

type example1BatchInsertFunction struct {
	q         sqlx.Queryer
	remaining []Example1
	scanner   Example1Scanner
}


// NewExample1BatchInsertFunction creates a scanner that inserts a batch of
// records into the database.
func NewExample1BatchInsertFunction(ctx context.Context, q sqlx.Queryer, p ...Example1) Example1Scanner {
	return &example1BatchInsertFunction{
		ctx: ctx,
		q:         q,
		remaining: p,
	}
}

type example1BatchInsertFunction struct {
	q         sqlx.Queryer
	remaining []Example1
	scanner   Example1Scanner
}

func (t *example1BatchInsertFunction) Scan(dst *Example1) error {
	return t.scanner.Scan(dst)
}

func (t *example1BatchInsertFunction) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *example1BatchInsertFunction) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *example1BatchInsertFunction) Next() bool {
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

func (t *example1BatchInsertFunction) advance(q sqlx.Queryer, p ...Example1) (Example1Scanner, []Example1, bool) {