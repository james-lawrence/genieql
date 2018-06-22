package example

import "database/sql"

// Bool scanner interface.
type Bool interface {
	Scan(arg *bool) error
	Next() bool
	Close() error
	Err() error
}

type errBool struct {
	e error
}

func (t errBool) Scan(arg *bool) error {
	return t.e
}

func (t errBool) Next() bool {
	return false
}

func (t errBool) Err() error {
	return t.e
}

func (t errBool) Close() error {
	return nil
}

const BoolStaticColumns = `arg`

// NewBoolStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewBoolStatic(rows *sql.Rows, err error) Bool {
	if err != nil {
		return errBool{e: err}
	}

	return boolStatic{
		Rows: rows,
	}
}

type boolStatic struct {
	Rows *sql.Rows
}

func (t boolStatic) Scan(arg *bool) error {
	var (
		c0 sql.NullBool
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := c0.Bool
		*arg = tmp
	}

	return t.Rows.Err()
}

func (t boolStatic) Err() error {
	return t.Rows.Err()
}

func (t boolStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t boolStatic) Next() bool {
	return t.Rows.Next()
}

// NewBoolStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewBoolStaticRow(row *sql.Row) BoolStaticRow {
	return BoolStaticRow{
		row: row,
	}
}

type BoolStaticRow struct {
	row *sql.Row
}

func (t BoolStaticRow) Scan(arg *bool) error {
	var (
		c0 sql.NullBool
	)

	if err := t.row.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := c0.Bool
		*arg = tmp
	}

	return nil
}

// NewBoolDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewBoolDynamic(rows *sql.Rows, err error) Bool {
	if err != nil {
		return errBool{e: err}
	}

	return boolDynamic{
		Rows: rows,
	}
}

type boolDynamic struct {
	Rows *sql.Rows
}

func (t boolDynamic) Scan(arg *bool) error {
	const (
		arg0 = "arg"
	)
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      sql.NullBool
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
			if c0.Valid {
				tmp := c0.Bool
				*arg = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t boolDynamic) Err() error {
	return t.Rows.Err()
}

func (t boolDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t boolDynamic) Next() bool {
	return t.Rows.Next()
}
