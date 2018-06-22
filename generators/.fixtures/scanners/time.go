package example

import (
	"database/sql"
	"time"
)

// Time scanner interface.
type Time interface {
	Scan(arg *time.Time) error
	Next() bool
	Close() error
	Err() error
}

type errTime struct {
	e error
}

func (t errTime) Scan(arg *time.Time) error {
	return t.e
}

func (t errTime) Next() bool {
	return false
}

func (t errTime) Err() error {
	return t.e
}

func (t errTime) Close() error {
	return nil
}

const TimeStaticColumns = `arg`

// NewTimeStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewTimeStatic(rows *sql.Rows, err error) Time {
	if err != nil {
		return errTime{e: err}
	}

	return timeStatic{
		Rows: rows,
	}
}

type timeStatic struct {
	Rows *sql.Rows
}

func (t timeStatic) Scan(arg *time.Time) error {
	var (
		c0 time.Time
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	*arg = c0

	return t.Rows.Err()
}

func (t timeStatic) Err() error {
	return t.Rows.Err()
}

func (t timeStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t timeStatic) Next() bool {
	return t.Rows.Next()
}

// NewTimeStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewTimeStaticRow(row *sql.Row) TimeStaticRow {
	return TimeStaticRow{
		row: row,
	}
}

type TimeStaticRow struct {
	row *sql.Row
}

func (t TimeStaticRow) Scan(arg *time.Time) error {
	var (
		c0 time.Time
	)

	if err := t.row.Scan(&c0); err != nil {
		return err
	}

	*arg = c0

	return nil
}

// NewTimeDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewTimeDynamic(rows *sql.Rows, err error) Time {
	if err != nil {
		return errTime{e: err}
	}

	return timeDynamic{
		Rows: rows,
	}
}

type timeDynamic struct {
	Rows *sql.Rows
}

func (t timeDynamic) Scan(arg *time.Time) error {
	const (
		arg0 = "arg"
	)
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
		case arg0:
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
		case arg0:
			*arg = c0
		}
	}

	return t.Rows.Err()
}

func (t timeDynamic) Err() error {
	return t.Rows.Err()
}

func (t timeDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t timeDynamic) Next() bool {
	return t.Rows.Next()
}
