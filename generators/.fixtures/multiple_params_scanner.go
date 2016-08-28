package example

import "database/sql"

// ExampleMultipleParam scanner interface.
type ExampleMultipleParam interface {
	Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error
	Next() bool
	Close() error
	Err() error
}

type errExampleMultipleParam struct {
	e error
}

func (t errExampleMultipleParam) Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error {
	return t.e
}

func (t errExampleMultipleParam) Next() bool {
	return false
}

func (t errExampleMultipleParam) Err() error {
	return t.e
}

func (t errExampleMultipleParam) Close() error {
	return nil
}

// StaticExampleMultipleParam creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleMultipleParam(rows *sql.Rows, err error) ExampleMultipleParam {
	if err != nil {
		return errExampleMultipleParam{e: err}
	}

	return staticExampleMultipleParam{
		Rows: rows,
	}
}

type staticExampleMultipleParam struct {
	Rows *sql.Rows
}

func (t staticExampleMultipleParam) Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullBool
		c3 sql.NullString
	)

	if err := t.Rows.Scan(&c0, &c1, &c2, &c3); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*arg1 = tmp
	}

	if c1.Valid {
		tmp := int(c1.Int64)
		*arg2 = tmp
	}

	if c2.Valid {
		tmp := c2.Bool
		*arg3 = tmp
	}

	if c3.Valid {
		tmp := c3.String
		*arg4 = tmp
	}

	return t.Rows.Err()
}

func (t staticExampleMultipleParam) Err() error {
	return t.Rows.Err()
}

func (t staticExampleMultipleParam) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleMultipleParam) Next() bool {
	return t.Rows.Next()
}

// NewStaticRowExampleMultipleParam creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStaticRowExampleMultipleParam(row *sql.Row) StaticRowExampleMultipleParam {
	return StaticRowExampleMultipleParam{
		row: row,
	}
}

type StaticRowExampleMultipleParam struct {
	row *sql.Row
}

func (t StaticRowExampleMultipleParam) Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullBool
		c3 sql.NullString
	)

	if err := t.row.Scan(&c0, &c1, &c2, &c3); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*arg1 = tmp
	}

	if c1.Valid {
		tmp := int(c1.Int64)
		*arg2 = tmp
	}

	if c2.Valid {
		tmp := c2.Bool
		*arg3 = tmp
	}

	if c3.Valid {
		tmp := c3.String
		*arg4 = tmp
	}

	return nil
}

// DynamicExampleMultipleParam creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleMultipleParam(rows *sql.Rows, err error) ExampleMultipleParam {
	if err != nil {
		return errExampleMultipleParam{e: err}
	}

	return dynamicExampleMultipleParam{
		Rows: rows,
	}
}

type dynamicExampleMultipleParam struct {
	Rows *sql.Rows
}

func (t dynamicExampleMultipleParam) Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error {
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      sql.NullInt64
		c1      sql.NullInt64
		c2      sql.NullBool
		c3      sql.NullString
	)

	if columns, err = t.Rows.Columns(); err != nil {
		return err
	}

	dst = make([]interface{}, 0, len(columns))

	for _, column := range columns {
		switch column {
		case "arg1":
			dst = append(dst, &c0)
		case "arg2":
			dst = append(dst, &c1)
		case "arg3":
			dst = append(dst, &c2)
		case "arg4":
			dst = append(dst, &c3)
		default:
			dst = append(dst, &ignored)
		}
	}

	if err := t.Rows.Scan(dst...); err != nil {
		return err
	}

	for _, column := range columns {
		switch column {
		case "arg1":
			if c0.Valid {
				tmp := int(c0.Int64)
				*arg1 = tmp
			}
		case "arg2":
			if c1.Valid {
				tmp := int(c1.Int64)
				*arg2 = tmp
			}
		case "arg3":
			if c2.Valid {
				tmp := c2.Bool
				*arg3 = tmp
			}
		case "arg4":
			if c3.Valid {
				tmp := c3.String
				*arg4 = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t dynamicExampleMultipleParam) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleMultipleParam) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleMultipleParam) Next() bool {
	return t.Rows.Next()
}
