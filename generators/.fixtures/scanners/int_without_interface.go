package example

import "database/sql"

// IntNoInterfaceStaticColumns generated by genieql
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

// intNoInterfaceStatic generated by genieql
type intNoInterfaceStatic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
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

// Err generated by genieql
func (t intNoInterfaceStatic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t intNoInterfaceStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
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

// IntNoInterfaceStaticRow generated by genieql
type IntNoInterfaceStaticRow struct {
	err error
	row *sql.Row
}

// Scan generated by genieql
func (t IntNoInterfaceStaticRow) Scan(arg *int) error {
	var (
		c0 sql.NullInt64
	)

	if t.err != nil {
		return t.err
	}

	if err := t.row.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*arg = tmp
	}

	return nil
}

// Err set an error to return by scan
func (t IntNoInterfaceStaticRow) Err(err error) IntNoInterfaceStaticRow {
	t.err = err
	return t
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

// intNoInterfaceDynamic generated by genieql
type intNoInterfaceDynamic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
func (t intNoInterfaceDynamic) Scan(arg *int) error {
	const (
		cn0 = "arg"
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
		case cn0:
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
		case cn0:
			if c0.Valid {
				tmp := int(c0.Int64)
				*arg = tmp
			}

		}
	}

	return t.Rows.Err()
}

// Err generated by genieql
func (t intNoInterfaceDynamic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t intNoInterfaceDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
func (t intNoInterfaceDynamic) Next() bool {
	return t.Rows.Next()
}
