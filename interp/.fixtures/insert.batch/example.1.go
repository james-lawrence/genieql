package example

import (
	"bitbucket.org/jatone/genieql/internal/sqlx"
)

func NewExample1BatchInsertFunction(q sqlx.Queryer, p ...StructA) Example1Scanner {
	return &example1BatchInsertFunction{
		q:         q,
		remaining: p,
	}
}
