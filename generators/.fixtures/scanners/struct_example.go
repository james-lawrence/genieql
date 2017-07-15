package example

import "database/sql"

// StructExample scanner interface.
type StructExample interface {
	Scan(arg *StructA) error
	Next() bool
	Close() error
	Err() error
}

type errStructExample struct {
	e error
}

func (t errStructExample) Scan(arg *StructA) error {
	return t.e
}

func (t errStructExample) Next() bool {
	return false
}

func (t errStructExample) Err() error {
	return t.e
}

func (t errStructExample) Close() error {
	return nil
}

const StructExampleStaticColumns = "a,b,c,d,e,f"

// NewStructExampleStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewStructExampleStatic(rows *sql.Rows, err error) StructExample {
	if err != nil {
		return errStructExample{e: err}
	}

	return structExampleStatic{
		Rows: rows,
	}
}

type structExampleStatic struct {
	Rows *sql.Rows
}

func (t structExampleStatic) Scan(arg *StructA) error {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullInt64
		c3 sql.NullBool
		c4 sql.NullBool
		c5 sql.NullBool
	)

	if err := t.Rows.Scan(&c0, &c1, &c2, &c3, &c4, &c5); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		arg.A = tmp
	}

	if c1.Valid {
		tmp := int(c1.Int64)
		arg.B = tmp
	}

	if c2.Valid {
		tmp := int(c2.Int64)
		arg.C = tmp
	}

	if c3.Valid {
		tmp := c3.Bool
		arg.D = tmp
	}

	if c4.Valid {
		tmp := c4.Bool
		arg.E = tmp
	}

	if c5.Valid {
		tmp := c5.Bool
		arg.F = tmp
	}

	return t.Rows.Err()
}

func (t structExampleStatic) Err() error {
	return t.Rows.Err()
}

func (t structExampleStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t structExampleStatic) Next() bool {
	return t.Rows.Next()
}

// NewStructExampleStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewStructExampleStaticRow(row *sql.Row) StructExampleStaticRow {
	return StructExampleStaticRow{
		row: row,
	}
}

type StructExampleStaticRow struct {
	row *sql.Row
}

func (t StructExampleStaticRow) Scan(arg *StructA) error {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullInt64
		c3 sql.NullBool
		c4 sql.NullBool
		c5 sql.NullBool
	)

	if err := t.row.Scan(&c0, &c1, &c2, &c3, &c4, &c5); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		arg.A = tmp
	}

	if c1.Valid {
		tmp := int(c1.Int64)
		arg.B = tmp
	}

	if c2.Valid {
		tmp := int(c2.Int64)
		arg.C = tmp
	}

	if c3.Valid {
		tmp := c3.Bool
		arg.D = tmp
	}

	if c4.Valid {
		tmp := c4.Bool
		arg.E = tmp
	}

	if c5.Valid {
		tmp := c5.Bool
		arg.F = tmp
	}

	return nil
}

// NewStructExampleDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewStructExampleDynamic(rows *sql.Rows, err error) StructExample {
	if err != nil {
		return errStructExample{e: err}
	}

	return structExampleDynamic{
		Rows: rows,
	}
}

type structExampleDynamic struct {
	Rows *sql.Rows
}

func (t structExampleDynamic) Scan(arg *StructA) error {
	const (
		a0 = "a"
		b1 = "b"
		c2 = "c"
		d3 = "d"
		e4 = "e"
		f5 = "f"
	)
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      sql.NullInt64
		c1      sql.NullInt64
		c2      sql.NullInt64
		c3      sql.NullBool
		c4      sql.NullBool
		c5      sql.NullBool
	)

	if columns, err = t.Rows.Columns(); err != nil {
		return err
	}

	dst = make([]interface{}, 0, len(columns))

	for _, column := range columns {
		switch column {
		case a0:
			dst = append(dst, &c0)
		case b1:
			dst = append(dst, &c1)
		case c2:
			dst = append(dst, &c2)
		case d3:
			dst = append(dst, &c3)
		case e4:
			dst = append(dst, &c4)
		case f5:
			dst = append(dst, &c5)
		default:
			dst = append(dst, &ignored)
		}
	}

	if err := t.Rows.Scan(dst...); err != nil {
		return err
	}

	for _, column := range columns {
		switch column {
		case a0:
			if c0.Valid {
				tmp := int(c0.Int64)
				arg.A = tmp
			}
		case b1:
			if c1.Valid {
				tmp := int(c1.Int64)
				arg.B = tmp
			}
		case c2:
			if c2.Valid {
				tmp := int(c2.Int64)
				arg.C = tmp
			}
		case d3:
			if c3.Valid {
				tmp := c3.Bool
				arg.D = tmp
			}
		case e4:
			if c4.Valid {
				tmp := c4.Bool
				arg.E = tmp
			}
		case f5:
			if c5.Valid {
				tmp := c5.Bool
				arg.F = tmp
			}
		}
	}

	return t.Rows.Err()
}

func (t structExampleDynamic) Err() error {
	return t.Rows.Err()
}

func (t structExampleDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t structExampleDynamic) Next() bool {
	return t.Rows.Next()
}
