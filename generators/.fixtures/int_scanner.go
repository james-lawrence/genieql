package example

import "database/sql"

// ExampleInt scanner interface.
type ExampleInt interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errExampleInt struct {
	e error
}

func (t errExampleInt) Scan(arg *int) error {
	return t.e
}

func (t errExampleInt) Next() bool {
	return false
}

func (t errExampleInt) Err() error {
	return t.e
}

func (t errExampleInt) Close() error {
	return nil
}

// StaticExampleInt creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleInt(rows *sql.Rows, err error) ExampleInt {
	if err != nil {
		return errExampleInt{e: err}
	}

	return staticExampleInt{
		Rows: rows,
	}
}

type staticExampleInt struct {
	Rows *sql.Rows
}

func (t staticExampleInt) Scan(arg *int) error {
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

func (t staticExampleInt) Err() error {
	return t.Rows.Err()
}

func (t staticExampleInt) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleInt) Next() bool {
	return t.Rows.Next()
}

// NewStaticRowExampleInt creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStaticRowExampleInt(row *sql.Row) StaticRowExampleInt {
	return StaticRowExampleInt{
		row: row,
	}
}

type StaticRowExampleInt struct {
	row *sql.Row
}

func (t StaticRowExampleInt) Scan(arg *int) error {
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

// DynamicExampleInt creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleInt(rows *sql.Rows, err error) ExampleInt {
	if err != nil {
		return errExampleInt{e: err}
	}

	return dynamicExampleInt{
		Rows: rows,
	}
}

type dynamicExampleInt struct {
	Rows *sql.Rows
}

func (t dynamicExampleInt) Scan(arg *int) error {
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
			if c0.Valid {
				tmp := int(c0.Int64)
				*arg = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t dynamicExampleInt) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleInt) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleInt) Next() bool {
	return t.Rows.Next()
}
