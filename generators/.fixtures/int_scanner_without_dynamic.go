package example

import "database/sql"

const ExampleIntNoDynamicStaticColumns = "arg"

// ExampleIntNoDynamic scanner interface.
type ExampleIntNoDynamic interface {
	Scan(arg *int) error
	Next() bool
	Close() error
	Err() error
}

type errExampleIntNoDynamic struct {
	e error
}

func (t errExampleIntNoDynamic) Scan(arg *int) error {
	return t.e
}

func (t errExampleIntNoDynamic) Next() bool {
	return false
}

func (t errExampleIntNoDynamic) Err() error {
	return t.e
}

func (t errExampleIntNoDynamic) Close() error {
	return nil
}

// StaticExampleIntNoDynamic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func StaticExampleIntNoDynamic(rows *sql.Rows, err error) ExampleIntNoDynamic {
	if err != nil {
		return errExampleIntNoDynamic{e: err}
	}

	return staticExampleIntNoDynamic{
		Rows: rows,
	}
}

type staticExampleIntNoDynamic struct {
	Rows *sql.Rows
}

func (t staticExampleIntNoDynamic) Scan(arg *int) error {
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

func (t staticExampleIntNoDynamic) Err() error {
	return t.Rows.Err()
}

func (t staticExampleIntNoDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t staticExampleIntNoDynamic) Next() bool {
	return t.Rows.Next()
}

// NewStaticRowExampleIntNoDynamic creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStaticRowExampleIntNoDynamic(row *sql.Row) StaticRowExampleIntNoDynamic {
	return StaticRowExampleIntNoDynamic{
		row: row,
	}
}

type StaticRowExampleIntNoDynamic struct {
	row *sql.Row
}

func (t StaticRowExampleIntNoDynamic) Scan(arg *int) error {
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
