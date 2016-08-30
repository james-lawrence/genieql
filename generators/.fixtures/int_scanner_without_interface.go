package example

import "database/sql"

// StaticExampleIntNoInterface creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleIntNoInterface(rows *sql.Rows, err error) ExampleIntNoInterface {
	if err != nil {
		return errExampleIntNoInterface{e: err}
	}

	return staticExampleIntNoInterface{
		Rows: rows,
	}
}

type staticExampleIntNoInterface struct {
	Rows *sql.Rows
}

func (t staticExampleIntNoInterface) Scan(arg *int) error {
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

func (t staticExampleIntNoInterface) Err() error {
	return t.Rows.Err()
}

func (t staticExampleIntNoInterface) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleIntNoInterface) Next() bool {
	return t.Rows.Next()
}

// NewStaticRowExampleIntNoInterface creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStaticRowExampleIntNoInterface(row *sql.Row) StaticRowExampleIntNoInterface {
	return StaticRowExampleIntNoInterface{
		row: row,
	}
}

type StaticRowExampleIntNoInterface struct {
	row *sql.Row
}

func (t StaticRowExampleIntNoInterface) Scan(arg *int) error {
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

// DynamicExampleIntNoInterface creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleIntNoInterface(rows *sql.Rows, err error) ExampleIntNoInterface {
	if err != nil {
		return errExampleIntNoInterface{e: err}
	}

	return dynamicExampleIntNoInterface{
		Rows: rows,
	}
}

type dynamicExampleIntNoInterface struct {
	Rows *sql.Rows
}

func (t dynamicExampleIntNoInterface) Scan(arg *int) error {
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

func (t dynamicExampleIntNoInterface) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleIntNoInterface) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleIntNoInterface) Next() bool {
	return t.Rows.Next()
}
