package example

import "database/sql"

// Int scanner interface.
type Int interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errInt struct {
	e error
}

func (t errInt) Scan(arg *int) error {
	return t.e
}

func (t errInt) Next() bool {
	return false
}

func (t errInt) Err() error {
	return t.e
}

func (t errInt) Close() error {
	return nil
}

const IntStaticColumns = "arg"

// NewIntStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewIntStatic(rows *sql.Rows, err error) Int {
	if err != nil {
		return errInt{e: err}
	}

	return intStatic{
		Rows: rows,
	}
}

type intStatic struct {
	Rows *sql.Rows
}

func (t intStatic) Scan(arg *int) error {
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

func (t intStatic) Err() error {
	return t.Rows.Err()
}

func (t intStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t intStatic) Next() bool {
	return t.Rows.Next()
}

// NewIntStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewIntStaticRow(row *sql.Row) IntStaticRow {
	return IntStaticRow{
		row: row,
	}
}

type IntStaticRow struct {
	row *sql.Row
}

func (t IntStaticRow) Scan(arg *int) error {
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

// NewIntDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewIntDynamic(rows *sql.Rows, err error) Int {
	if err != nil {
		return errInt{e: err}
	}

	return intDynamic{
		Rows: rows,
	}
}

type intDynamic struct {
	Rows *sql.Rows
}

func (t intDynamic) Scan(arg *int) error {
	const (
		arg0 = "arg"
	)
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      sql.NullInt64
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
				tmp := int(c0.Int64)
				*arg = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t intDynamic) Err() error {
	return t.Rows.Err()
}

func (t intDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t intDynamic) Next() bool {
	return t.Rows.Next()
}
