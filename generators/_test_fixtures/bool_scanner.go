package example

import "database/sql"

// ExampleBool scanner interface.
type ExampleBool interface {
	Scan(arg *bool) error
	Next() bool
	Close() error
	Err() error
}

type errExampleBool struct {
	e error
}

func (t errExampleBool) Scan(arg *bool) error {
	return t.e
}

func (t errExampleBool) Next() bool {
	return false
}

func (t errExampleBool) Err() error {
	return t.e
}

func (t errExampleBool) Close() error {
	return nil
}

// StaticExampleBool creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleBool(rows *sql.Rows, err error) ExampleBool {
	if err != nil {
		return errExampleBool{e: err}
	}

	return staticExampleBool{
		Rows: rows,
	}
}

type staticExampleBool struct {
	Rows *sql.Rows
}

func (t staticExampleBool) Scan(arg *bool) error {
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

func (t staticExampleBool) Err() error {
	return t.Rows.Err()
}

func (t staticExampleBool) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleBool) Next() bool {
	return t.Rows.Next()
}

// DynamicExampleBool creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func DynamicExampleBool(rows *sql.Rows, err error) ExampleBool {
	if err != nil {
		return errExampleBool{e: err}
	}

	return dynamicExampleBool{
		Rows: rows,
	}
}

type dynamicExampleBool struct {
	Rows *sql.Rows
}

func (t dynamicExampleBool) Scan(arg *bool) error {
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
				tmp := c0.Bool
				*arg = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t dynamicExampleBool) Err() error {
	return t.Rows.Err()
}

func (t dynamicExampleBool) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamicExampleBool) Next() bool {
	return t.Rows.Next()
}
