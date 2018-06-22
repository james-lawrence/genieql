package example

import "database/sql"

const IntNoInterfaceStaticColumns = `arg`

// NewIntNoInterfaceStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewIntNoInterfaceStatic(rows *sql.Rows, err error) IntNoInterface {
	if err != nil {
		return errIntNoInterface{e: err}
	}

	return intNoInterfaceStatic{
		Rows: rows,
	}
}

type intNoInterfaceStatic struct {
	Rows *sql.Rows
}

func (t intNoInterfaceStatic) Scan(arg *int) error {
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

func (t intNoInterfaceStatic) Err() error {
	return t.Rows.Err()
}

func (t intNoInterfaceStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t intNoInterfaceStatic) Next() bool {
	return t.Rows.Next()
}

// NewIntNoInterfaceStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewIntNoInterfaceStaticRow(row *sql.Row) IntNoInterfaceStaticRow {
	return IntNoInterfaceStaticRow{
		row: row,
	}
}

type IntNoInterfaceStaticRow struct {
	row *sql.Row
}

func (t IntNoInterfaceStaticRow) Scan(arg *int) error {
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

// NewIntNoInterfaceDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewIntNoInterfaceDynamic(rows *sql.Rows, err error) IntNoInterface {
	if err != nil {
		return errIntNoInterface{e: err}
	}

	return intNoInterfaceDynamic{
		Rows: rows,
	}
}

type intNoInterfaceDynamic struct {
	Rows *sql.Rows
}

func (t intNoInterfaceDynamic) Scan(arg *int) error {
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

func (t intNoInterfaceDynamic) Err() error {
	return t.Rows.Err()
}

func (t intNoInterfaceDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t intNoInterfaceDynamic) Next() bool {
	return t.Rows.Next()
}
