package example

import (
	"bitbucket.org/jatone/genieql/internal/sqlx"
)

func NewExample1BatchInsertFunction(q sqlx.Queryer, p ...StructA) Example1Scanner {
	return &batchInsertExample1{
		q:         q,
		remaining: p,
	}
}

func (t *batchInsertExample1) Scan(dst *StructA) error {
	return t.scanner.Scan(dst)
}

func (t *batchInsertExample1) Err() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Err()
}

func (t *batchInsertExample1) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *batchInsertExample1) Next() bool {
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

func (t *batchInsertExample1) advance(a ...StructA) (ExampleScanner, []StructA, bool) {
	switch len(a) {
	case 0:
		return nil, []StructA(nil), false
	case 1:
	}
	return false
}
