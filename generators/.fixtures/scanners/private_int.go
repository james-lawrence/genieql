package example

import "database/sql"

// PrivateInt scanner interface.
type PrivateInt interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errPrivateInt struct {
	e error
}

func (t errPrivateInt) Scan(arg *int) error {
	return t.e
}

func (t errPrivateInt) Next() bool {
	return false
}

func (t errPrivateInt) Err() error {
	return t.e
}

func (t errPrivateInt) Close() error {
	return nil
}

const PrivateIntStaticColumns = "arg"

// NewPrivateIntStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewPrivateIntStatic(rows *sql.Rows, err error) PrivateInt {
	if err != nil {
		return errPrivateInt{e: err}
	}

	return privateIntStatic{
		Rows: rows,
	}
}

type privateIntStatic struct {
	Rows *sql.Rows
}

func (t privateIntStatic) Scan(arg *int) error {
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

func (t privateIntStatic) Err() error {
	return t.Rows.Err()
}

func (t privateIntStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t privateIntStatic) Next() bool {
	return t.Rows.Next()
}

// NewPrivateIntStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewPrivateIntStaticRow(row *sql.Row) PrivateIntStaticRow {
	return PrivateIntStaticRow{
		row: row,
	}
}

type PrivateIntStaticRow struct {
	row *sql.Row
}

func (t PrivateIntStaticRow) Scan(arg *int) error {
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

// NewPrivateIntDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewPrivateIntDynamic(rows *sql.Rows, err error) PrivateInt {
	if err != nil {
		return errPrivateInt{e: err}
	}

	return privateIntDynamic{
		Rows: rows,
	}
}

type privateIntDynamic struct {
	Rows *sql.Rows
}

func (t privateIntDynamic) Scan(arg *int) error {
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

func (t privateIntDynamic) Err() error {
	return t.Rows.Err()
}

func (t privateIntDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t privateIntDynamic) Next() bool {
	return t.Rows.Next()
}
