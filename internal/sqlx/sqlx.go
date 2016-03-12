package sqlx

import (
	"database/sql"
)

// Queryer interface for executing queries.
type Queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
}
