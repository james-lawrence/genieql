package example

import "database/sql"

// IntNoDynamic scanner interface.
type IntNoDynamic interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errIntNoDynamic struct {
	e error
}

func (t errIntNoDynamic) Scan(arg *int) error {
	return t.e
}

func (t errIntNoDynamic) Next() bool {
	return false
}

func (t errIntNoDynamic) Err() error {
	return t.e
}

func (t errIntNoDynamic) Close() error {
	return nil
}

const IntNoDynamicStaticColumns = `arg`

// NewIntNoDynamicStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewIntNoDynamicStatic(rows *sql.Rows, err error) IntNoDynamic {
	if err != nil {
		return errIntNoDynamic{e: err}
	}

	return intNoDynamicStatic{
		Rows: rows,
	}
}

type intNoDynamicStatic struct {
	Rows *sql.Rows
}

func (t intNoDynamicStatic) Scan(arg *int) error {
	var (
		c0 sql.NullInt64
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*arg = tmp
	}

	return t.Rows.Err()
}

func (t intNoDynamicStatic) Err() error {
	return t.Rows.Err()
}

func (t intNoDynamicStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t intNoDynamicStatic) Next() bool {
	return t.Rows.Next()
}

// NewIntNoDynamicStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewIntNoDynamicStaticRow(row *sql.Row) IntNoDynamicStaticRow {
	return IntNoDynamicStaticRow{
		row: row,
	}
}

type IntNoDynamicStaticRow struct {
	row *sql.Row
}

func (t IntNoDynamicStaticRow) Scan(arg *int) error {
	var (
		c0 sql.NullInt64
	)

	if err := t.row.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*arg = tmp
	}

	return nil
}
