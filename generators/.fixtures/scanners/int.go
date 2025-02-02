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

// IntStaticColumns generated by genieql
const IntStaticColumns = `arg`

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

// intStatic generated by genieql
type intStatic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
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

// Err generated by genieql
func (t intStatic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t intStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
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

// IntStaticRow generated by genieql
type IntStaticRow struct {
	err error
	row *sql.Row
}

// Scan generated by genieql
func (t IntStaticRow) Scan(arg *int) error {
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
func (t IntStaticRow) Err(err error) IntStaticRow {
	t.err = err
	return t
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

// intDynamic generated by genieql
type intDynamic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
func (t intDynamic) Scan(arg *int) error {
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
func (t intDynamic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t intDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
func (t intDynamic) Next() bool {
	return t.Rows.Next()
}
