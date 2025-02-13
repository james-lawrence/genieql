package sqlx

import (
	"context"
	"database/sql"
	"log"
)

// Queryer interface for executing queries.
type Queryer interface {
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
}

// Debug creates a DebuggingQueryer
func Debug(q Queryer) DebuggingQueryer {
	return DebuggingQueryer{
		Delegate: q,
	}
}

// DebuggingQueryer queryer that prints out the queries being executed.
type DebuggingQueryer struct {
	Delegate Queryer
}

// Query execute a query
func (t DebuggingQueryer) Query(q string, args ...interface{}) (*sql.Rows, error) {
	return t.QueryContext(context.Background(), q, args...)
}

// QueryRow executes a query that returns a single row.
func (t DebuggingQueryer) QueryRow(q string, args ...interface{}) *sql.Row {
	return t.QueryRowContext(context.Background(), q, args...)
}

// Exec executes a statement.
func (t DebuggingQueryer) Exec(q string, args ...interface{}) (sql.Result, error) {
	return t.ExecContext(context.Background(), q, args...)
}

// QueryContext ...
func (t DebuggingQueryer) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	log.Printf("%s:\n%#v\n", q, args)
	return t.Delegate.QueryContext(ctx, q, args...)
}

// QueryRowContext ...
func (t DebuggingQueryer) QueryRowContext(ctx context.Context, q string, args ...interface{}) *sql.Row {
	log.Printf("%s:\n%#v\n", q, args)
	return t.Delegate.QueryRowContext(ctx, q, args...)
}

// ExecContext ...
func (t DebuggingQueryer) ExecContext(ctx context.Context, q string, args ...interface{}) (sql.Result, error) {
	log.Printf("%s:\n%#v\n", q, args)
	return t.Delegate.ExecContext(ctx, q, args...)
}
