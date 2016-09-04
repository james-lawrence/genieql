package example

import "database/sql"

const ExampleIntNoStaticStaticColumns = "arg"

// ExampleIntNoStatic scanner interface.
type ExampleIntNoStatic interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errExampleIntNoStatic struct {
	e error
}

func (t errExampleIntNoStatic) Scan(arg *int) error {
	return t.e
}

func (t errExampleIntNoStatic) Next() bool {
	return false
}

func (t errExampleIntNoStatic) Err() error {
	return t.e
}

func (t errExampleIntNoStatic) Close() error {
	return nil
}

// DynamicExampleIntNoStatic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleIntNoStatic(rows *sql.Rows, err error) ExampleIntNoStatic {
	if err != nil {
		return errExampleIntNoStatic{e: err}
	}

	return dynamicExampleIntNoStatic{
		Rows: rows,
	}
}

type dynamicExampleIntNoStatic struct {
	Rows *sql.Rows
}

func (t dynamicExampleIntNoStatic) Scan(arg *int) error {
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

func (t dynamicExampleIntNoStatic) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleIntNoStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleIntNoStatic) Next() bool {
	return t.Rows.Next()
}
