package example

import (
	"database/sql"
	"time"
)

// ExampleTime scanner interface.
type ExampleTime interface {
	Scan(arg *time.Time) error
	Next() bool
	Close() error
	Err() error
}

type errExampleTime struct {
	e error
}

func (t errExampleTime) Scan(arg *time.Time) error {
	return t.e
}

func (t errExampleTime) Next() bool {
	return false
}

func (t errExampleTime) Err() error {
	return t.e
}

func (t errExampleTime) Close() error {
	return nil
}

// StaticExampleTime creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleTime(rows *sql.Rows, err error) ExampleTime {
	if err != nil {
		return errExampleTime{e: err}
	}

	return staticExampleTime{
		Rows: rows,
	}
}

type staticExampleTime struct {
	Rows *sql.Rows
}

func (t staticExampleTime) Scan(arg *time.Time) error {
	var (
		c0 time.Time
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	*arg = c0

	return t.Rows.Err()
}

func (t staticExampleTime) Err() error {
	return t.Rows.Err()
}

func (t staticExampleTime) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleTime) Next() bool {
	return t.Rows.Next()
}

// NewStaticRowExampleTime creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStaticRowExampleTime(row *sql.Row) StaticRowExampleTime {
	return StaticRowExampleTime{
		row: row,
	}
}

type StaticRowExampleTime struct {
	row *sql.Row
}

func (t StaticRowExampleTime) Scan(arg *time.Time) error {
	var (
		c0 time.Time
	)

	if err := t.row.Scan(&c0); err != nil {
		return err
	}

	*arg = c0

	return nil
}

// DynamicExampleTime creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleTime(rows *sql.Rows, err error) ExampleTime {
	if err != nil {
		return errExampleTime{e: err}
	}

	return dynamicExampleTime{
		Rows: rows,
	}
}

type dynamicExampleTime struct {
	Rows *sql.Rows
}

func (t dynamicExampleTime) Scan(arg *time.Time) error {
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      time.Time
	)

	if columns, err = t.Rows.Columns(); err != nil {
		return err
	}

	dst = make([]interface{}, 0, len(columns))

	for _, column := range columns {
		switch column {
		case "arg":
			dst = append(dst, &c0)
		default:
			dst = append(dst, &ignored)
		}
	}

	if err := t.Rows.Scan(dst...); err != nil {
		return err
	}

	for _, column := range columns {
		switch column {
		case "arg":
			*arg = c0
		}
	}

	return t.Rows.Err()
}

func (t dynamicExampleTime) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleTime) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleTime) Next() bool {
	return t.Rows.Next()
}
