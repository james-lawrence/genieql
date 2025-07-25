//go:build !genieql.ignore
// +build !genieql.ignore

package example2

import (
	"context"
	"database/sql"
	"time"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql/internal/sqlx"
)

// DO NOT EDIT: This File was auto generated by the following command:
// genieql

// Example1 generated by genieql
type Example1 struct {
	BigintField          int64
	BoolField            bool
	ByteArrayField       []byte
	DoublePrecisionField float64
	Int2Field            int16
	IntField             int32
	RealField            float32
	SmallintField        int16
	TextField            string
	TimestampField       time.Time
	UbigintField         uint64
	UintegerField        uint32
	UUIDField            string
}

// Example2 generated by genieql
type Example2 struct {
	BoolField      bool
	TextField      string
	TimestampField time.Time
	UUIDField      string
}

// Example1Scanner scanner interface.
type Example1Scanner interface {
	Scan(i *Example1) error
	Next() bool
	Close() error
	Err() error
}

type errExample1Scanner struct {
	e error
}

func (t errExample1Scanner) Scan(i *Example1) error {
	return t.e
}

func (t errExample1Scanner) Next() bool {
	return false
}

func (t errExample1Scanner) Err() error {
	return t.e
}

func (t errExample1Scanner) Close() error {
	return nil
}

// Example1ScannerStaticColumns generated by genieql
const Example1ScannerStaticColumns = `"bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`

// NewExample1ScannerStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewExample1ScannerStatic(rows *sql.Rows, err error) Example1Scanner {
	if err != nil {
		return errExample1Scanner{e: err}
	}

	return example1ScannerStatic{
		Rows: rows,
	}
}

// example1ScannerStatic generated by genieql
type example1ScannerStatic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
func (t example1ScannerStatic) Scan(i *Example1) error {
	var (
		c0  sql.NullInt64
		c1  sql.NullBool
		c2  []byte
		c3  sql.NullFloat64
		c4  sql.NullInt16
		c5  sql.NullInt32
		c6  sql.NullFloat64
		c7  sql.NullInt16
		c8  sql.NullString
		c9  sql.NullTime
		c10 sql.Null[uint64]
		c11 sql.NullInt64
		c12 sql.NullString
	)

	if err := t.Rows.Scan(&c0, &c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10, &c11, &c12); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int64(c0.Int64)
		i.BigintField = tmp
	}

	if c1.Valid {
		tmp := c1.Bool
		i.BoolField = tmp
	}

	i.ByteArrayField = c2

	if c3.Valid {
		tmp := float64(c3.Float64)
		i.DoublePrecisionField = tmp
	}

	if c4.Valid {
		tmp := int16(c4.Int16)
		i.Int2Field = tmp
	}

	if c5.Valid {
		tmp := int32(c5.Int32)
		i.IntField = tmp
	}

	if c6.Valid {
		tmp := float32(c6.Float64)
		i.RealField = tmp
	}

	if c7.Valid {
		tmp := int16(c7.Int16)
		i.SmallintField = tmp
	}

	if c8.Valid {
		tmp := string(c8.String)
		i.TextField = tmp
	}

	if c9.Valid {
		tmp := c9.Time
		i.TimestampField = tmp
	}

	if c10.Valid {
		tmp := c10.V
		i.UbigintField = tmp
	}

	if c11.Valid {
		tmp := uint32(c11.Int64)
		i.UintegerField = tmp
	}

	if c12.Valid {
		if uid, err := uuid.FromBytes([]byte(c12.String)); err != nil {
			return err
		} else {
			i.UUIDField = uid.String()
		}
	}

	return t.Rows.Err()
}

// Err generated by genieql
func (t example1ScannerStatic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t example1ScannerStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
func (t example1ScannerStatic) Next() bool {
	return t.Rows.Next()
}

// NewExample1ScannerStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewExample1ScannerStaticRow(row *sql.Row) Example1ScannerStaticRow {
	return Example1ScannerStaticRow{
		row: row,
	}
}

// Example1ScannerStaticRow generated by genieql
type Example1ScannerStaticRow struct {
	err error
	row *sql.Row
}

// Scan generated by genieql
func (t Example1ScannerStaticRow) Scan(i *Example1) error {
	var (
		c0  sql.NullInt64
		c1  sql.NullBool
		c2  []byte
		c3  sql.NullFloat64
		c4  sql.NullInt16
		c5  sql.NullInt32
		c6  sql.NullFloat64
		c7  sql.NullInt16
		c8  sql.NullString
		c9  sql.NullTime
		c10 sql.Null[uint64]
		c11 sql.NullInt64
		c12 sql.NullString
	)

	if t.err != nil {
		return t.err
	}

	if err := t.row.Scan(&c0, &c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10, &c11, &c12); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int64(c0.Int64)
		i.BigintField = tmp
	}

	if c1.Valid {
		tmp := c1.Bool
		i.BoolField = tmp
	}

	i.ByteArrayField = c2

	if c3.Valid {
		tmp := float64(c3.Float64)
		i.DoublePrecisionField = tmp
	}

	if c4.Valid {
		tmp := int16(c4.Int16)
		i.Int2Field = tmp
	}

	if c5.Valid {
		tmp := int32(c5.Int32)
		i.IntField = tmp
	}

	if c6.Valid {
		tmp := float32(c6.Float64)
		i.RealField = tmp
	}

	if c7.Valid {
		tmp := int16(c7.Int16)
		i.SmallintField = tmp
	}

	if c8.Valid {
		tmp := string(c8.String)
		i.TextField = tmp
	}

	if c9.Valid {
		tmp := c9.Time
		i.TimestampField = tmp
	}

	if c10.Valid {
		tmp := c10.V
		i.UbigintField = tmp
	}

	if c11.Valid {
		tmp := uint32(c11.Int64)
		i.UintegerField = tmp
	}

	if c12.Valid {
		if uid, err := uuid.FromBytes([]byte(c12.String)); err != nil {
			return err
		} else {
			i.UUIDField = uid.String()
		}
	}

	return nil
}

// Err set an error to return by scan
func (t Example1ScannerStaticRow) Err(err error) Example1ScannerStaticRow {
	t.err = err
	return t
}

// NewExample1ScannerDynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func NewExample1ScannerDynamic(rows *sql.Rows, err error) Example1Scanner {
	if err != nil {
		return errExample1Scanner{e: err}
	}

	return example1ScannerDynamic{
		Rows: rows,
	}
}

// example1ScannerDynamic generated by genieql
type example1ScannerDynamic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
func (t example1ScannerDynamic) Scan(i *Example1) error {
	const (
		cn0  = "bigint_field"
		cn1  = "bool_field"
		cn2  = "byte_array_field"
		cn3  = "double_precision_field"
		cn4  = "int2_field"
		cn5  = "int_field"
		cn6  = "real_field"
		cn7  = "smallint_field"
		cn8  = "text_field"
		cn9  = "timestamp_field"
		cn10 = "ubigint_field"
		cn11 = "uinteger_field"
		cn12 = "uuid_field"
	)
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		c0      sql.NullInt64
		c1      sql.NullBool
		c2      []byte
		c3      sql.NullFloat64
		c4      sql.NullInt16
		c5      sql.NullInt32
		c6      sql.NullFloat64
		c7      sql.NullInt16
		c8      sql.NullString
		c9      sql.NullTime
		c10     sql.Null[uint64]
		c11     sql.NullInt64
		c12     sql.NullString
	)

	if columns, err = t.Rows.Columns(); err != nil {
		return err
	}

	dst = make([]interface{}, 0, len(columns))

	for _, column := range columns {
		switch column {
		case cn0:
			dst = append(dst, &c0)
		case cn1:
			dst = append(dst, &c1)
		case cn2:
			dst = append(dst, &c2)
		case cn3:
			dst = append(dst, &c3)
		case cn4:
			dst = append(dst, &c4)
		case cn5:
			dst = append(dst, &c5)
		case cn6:
			dst = append(dst, &c6)
		case cn7:
			dst = append(dst, &c7)
		case cn8:
			dst = append(dst, &c8)
		case cn9:
			dst = append(dst, &c9)
		case cn10:
			dst = append(dst, &c10)
		case cn11:
			dst = append(dst, &c11)
		case cn12:
			dst = append(dst, &c12)
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
				tmp := int64(c0.Int64)
				i.BigintField = tmp
			}

		case cn1:
			if c1.Valid {
				tmp := c1.Bool
				i.BoolField = tmp
			}

		case cn2:
			i.ByteArrayField = c2

		case cn3:
			if c3.Valid {
				tmp := float64(c3.Float64)
				i.DoublePrecisionField = tmp
			}

		case cn4:
			if c4.Valid {
				tmp := int16(c4.Int16)
				i.Int2Field = tmp
			}

		case cn5:
			if c5.Valid {
				tmp := int32(c5.Int32)
				i.IntField = tmp
			}

		case cn6:
			if c6.Valid {
				tmp := float32(c6.Float64)
				i.RealField = tmp
			}

		case cn7:
			if c7.Valid {
				tmp := int16(c7.Int16)
				i.SmallintField = tmp
			}

		case cn8:
			if c8.Valid {
				tmp := string(c8.String)
				i.TextField = tmp
			}

		case cn9:
			if c9.Valid {
				tmp := c9.Time
				i.TimestampField = tmp
			}

		case cn10:
			if c10.Valid {
				tmp := c10.V
				i.UbigintField = tmp
			}

		case cn11:
			if c11.Valid {
				tmp := uint32(c11.Int64)
				i.UintegerField = tmp
			}

		case cn12:
			if c12.Valid {
				if uid, err := uuid.FromBytes([]byte(c12.String)); err != nil {
					return err
				} else {
					i.UUIDField = uid.String()
				}
			}

		}
	}

	return t.Rows.Err()
}

// Err generated by genieql
func (t example1ScannerDynamic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t example1ScannerDynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
func (t example1ScannerDynamic) Next() bool {
	return t.Rows.Next()
}

// ExampleComboScanner scanner interface.
type ExampleComboScanner interface {
	Scan(i *int, ts *time.Time, e1 *Example1, e2 *Example2) error
	Next() bool
	Close() error
	Err() error
}

type errExampleComboScanner struct {
	e error
}

func (t errExampleComboScanner) Scan(i *int, ts *time.Time, e1 *Example1, e2 *Example2) error {
	return t.e
}

func (t errExampleComboScanner) Next() bool {
	return false
}

func (t errExampleComboScanner) Err() error {
	return t.e
}

func (t errExampleComboScanner) Close() error {
	return nil
}

// NewExampleComboScannerStatic creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func NewExampleComboScannerStatic(rows *sql.Rows, err error) ExampleComboScanner {
	if err != nil {
		return errExampleComboScanner{e: err}
	}

	return exampleComboScannerStatic{
		Rows: rows,
	}
}

// exampleComboScannerStatic generated by genieql
type exampleComboScannerStatic struct {
	Rows *sql.Rows
}

// Scan generated by genieql
func (t exampleComboScannerStatic) Scan(i *int, ts *time.Time, e1 *Example1, e2 *Example2) error {
	var (
		c0  sql.NullInt64
		c1  sql.NullTime
		c2  sql.NullInt64
		c3  sql.NullBool
		c4  []byte
		c5  sql.NullFloat64
		c6  sql.NullInt16
		c7  sql.NullInt32
		c8  sql.NullFloat64
		c9  sql.NullInt16
		c10 sql.NullString
		c11 sql.NullTime
		c12 sql.Null[uint64]
		c13 sql.NullInt64
		c14 sql.NullString
		c15 sql.NullBool
		c16 sql.NullString
		c17 sql.NullTime
		c18 sql.NullString
	)

	if err := t.Rows.Scan(&c0, &c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10, &c11, &c12, &c13, &c14, &c15, &c16, &c17, &c18); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*i = tmp
	}

	if c1.Valid {
		tmp := c1.Time
		*ts = tmp
	}

	if c2.Valid {
		tmp := int64(c2.Int64)
		e1.BigintField = tmp
	}

	if c3.Valid {
		tmp := c3.Bool
		e1.BoolField = tmp
	}

	e1.ByteArrayField = c4

	if c5.Valid {
		tmp := float64(c5.Float64)
		e1.DoublePrecisionField = tmp
	}

	if c6.Valid {
		tmp := int16(c6.Int16)
		e1.Int2Field = tmp
	}

	if c7.Valid {
		tmp := int32(c7.Int32)
		e1.IntField = tmp
	}

	if c8.Valid {
		tmp := float32(c8.Float64)
		e1.RealField = tmp
	}

	if c9.Valid {
		tmp := int16(c9.Int16)
		e1.SmallintField = tmp
	}

	if c10.Valid {
		tmp := string(c10.String)
		e1.TextField = tmp
	}

	if c11.Valid {
		tmp := c11.Time
		e1.TimestampField = tmp
	}

	if c12.Valid {
		tmp := c12.V
		e1.UbigintField = tmp
	}

	if c13.Valid {
		tmp := uint32(c13.Int64)
		e1.UintegerField = tmp
	}

	if c14.Valid {
		if uid, err := uuid.FromBytes([]byte(c14.String)); err != nil {
			return err
		} else {
			e1.UUIDField = uid.String()
		}
	}

	if c15.Valid {
		tmp := c15.Bool
		e2.BoolField = tmp
	}

	if c16.Valid {
		tmp := string(c16.String)
		e2.TextField = tmp
	}

	if c17.Valid {
		tmp := c17.Time
		e2.TimestampField = tmp
	}

	if c18.Valid {
		if uid, err := uuid.FromBytes([]byte(c18.String)); err != nil {
			return err
		} else {
			e2.UUIDField = uid.String()
		}
	}

	return t.Rows.Err()
}

// Err generated by genieql
func (t exampleComboScannerStatic) Err() error {
	return t.Rows.Err()
}

// Close generated by genieql
func (t exampleComboScannerStatic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

// Next generated by genieql
func (t exampleComboScannerStatic) Next() bool {
	return t.Rows.Next()
}

// NewExampleComboScannerStaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func NewExampleComboScannerStaticRow(row *sql.Row) ExampleComboScannerStaticRow {
	return ExampleComboScannerStaticRow{
		row: row,
	}
}

// ExampleComboScannerStaticRow generated by genieql
type ExampleComboScannerStaticRow struct {
	err error
	row *sql.Row
}

// Scan generated by genieql
func (t ExampleComboScannerStaticRow) Scan(i *int, ts *time.Time, e1 *Example1, e2 *Example2) error {
	var (
		c0  sql.NullInt64
		c1  sql.NullTime
		c2  sql.NullInt64
		c3  sql.NullBool
		c4  []byte
		c5  sql.NullFloat64
		c6  sql.NullInt16
		c7  sql.NullInt32
		c8  sql.NullFloat64
		c9  sql.NullInt16
		c10 sql.NullString
		c11 sql.NullTime
		c12 sql.Null[uint64]
		c13 sql.NullInt64
		c14 sql.NullString
		c15 sql.NullBool
		c16 sql.NullString
		c17 sql.NullTime
		c18 sql.NullString
	)

	if t.err != nil {
		return t.err
	}

	if err := t.row.Scan(&c0, &c1, &c2, &c3, &c4, &c5, &c6, &c7, &c8, &c9, &c10, &c11, &c12, &c13, &c14, &c15, &c16, &c17, &c18); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		*i = tmp
	}

	if c1.Valid {
		tmp := c1.Time
		*ts = tmp
	}

	if c2.Valid {
		tmp := int64(c2.Int64)
		e1.BigintField = tmp
	}

	if c3.Valid {
		tmp := c3.Bool
		e1.BoolField = tmp
	}

	e1.ByteArrayField = c4

	if c5.Valid {
		tmp := float64(c5.Float64)
		e1.DoublePrecisionField = tmp
	}

	if c6.Valid {
		tmp := int16(c6.Int16)
		e1.Int2Field = tmp
	}

	if c7.Valid {
		tmp := int32(c7.Int32)
		e1.IntField = tmp
	}

	if c8.Valid {
		tmp := float32(c8.Float64)
		e1.RealField = tmp
	}

	if c9.Valid {
		tmp := int16(c9.Int16)
		e1.SmallintField = tmp
	}

	if c10.Valid {
		tmp := string(c10.String)
		e1.TextField = tmp
	}

	if c11.Valid {
		tmp := c11.Time
		e1.TimestampField = tmp
	}

	if c12.Valid {
		tmp := c12.V
		e1.UbigintField = tmp
	}

	if c13.Valid {
		tmp := uint32(c13.Int64)
		e1.UintegerField = tmp
	}

	if c14.Valid {
		if uid, err := uuid.FromBytes([]byte(c14.String)); err != nil {
			return err
		} else {
			e1.UUIDField = uid.String()
		}
	}

	if c15.Valid {
		tmp := c15.Bool
		e2.BoolField = tmp
	}

	if c16.Valid {
		tmp := string(c16.String)
		e2.TextField = tmp
	}

	if c17.Valid {
		tmp := c17.Time
		e2.TimestampField = tmp
	}

	if c18.Valid {
		if uid, err := uuid.FromBytes([]byte(c18.String)); err != nil {
			return err
		} else {
			e2.UUIDField = uid.String()
		}
	}

	return nil
}

// Err set an error to return by scan
func (t ExampleComboScannerStaticRow) Err(err error) ExampleComboScannerStaticRow {
	t.err = err
	return t
}

// Example1Insert1StaticColumns generated by genieql
const Example1Insert1StaticColumns = `$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,DEFAULT`

// Example1Insert1Explode generated by genieql
func Example1Insert1Explode(a *Example1) ([]interface{}, error) {
	var (
		c0  sql.NullInt64    // bigint_field
		c1  sql.NullBool     // bool_field
		c2  []byte           // byte_array_field
		c3  sql.NullFloat64  // double_precision_field
		c4  sql.NullInt16    // int2_field
		c5  sql.NullInt32    // int_field
		c6  sql.NullFloat64  // real_field
		c7  sql.NullInt16    // smallint_field
		c8  sql.NullString   // text_field
		c9  sql.NullTime     // timestamp_field
		c10 sql.Null[uint64] // ubigint_field
		c11 sql.NullInt64    // uinteger_field
	)

	c0.Valid = true
	c0.Int64 = int64(a.BigintField)

	c1.Valid = true
	c1.Bool = a.BoolField

	c2 = a.ByteArrayField

	c3.Valid = true
	c3.Float64 = float64(a.DoublePrecisionField)

	c4.Valid = true
	c4.Int16 = int16(a.Int2Field)

	c5.Valid = true
	c5.Int32 = int32(a.IntField)

	c6.Valid = true
	c6.Float64 = float64(a.RealField)

	c7.Valid = true
	c7.Int16 = int16(a.SmallintField)

	c8.Valid = true
	c8.String = a.TextField

	c9.Valid = true
	c9.Time = a.TimestampField

	c10.Valid = true
	c10.V = a.UbigintField

	c11.Valid = true
	c11.Int64 = int64(a.UintegerField)

	return []interface{}{c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11}, nil
}

// Example1Insert1 generated by genieql
func Example1Insert1(ctx context.Context, q sqlx.Queryer, a Example1) Example1ScannerStaticRow {
	const query = `INSERT INTO "example1" ("bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,DEFAULT) RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
	var (
		c0  sql.NullInt64    // bigint_field
		c1  sql.NullBool     // bool_field
		c2  []byte           // byte_array_field
		c3  sql.NullFloat64  // double_precision_field
		c4  sql.NullInt16    // int2_field
		c5  sql.NullInt32    // int_field
		c6  sql.NullFloat64  // real_field
		c7  sql.NullInt16    // smallint_field
		c8  sql.NullString   // text_field
		c9  sql.NullTime     // timestamp_field
		c10 sql.Null[uint64] // ubigint_field
		c11 sql.NullInt64
	)
	c0.Valid = true
	c0.Int64 = int64(a.BigintField)
	c1.Valid = true
	c1.Bool = a.BoolField
	c2 = a.ByteArrayField
	c3.Valid = true
	c3.Float64 = float64(a.DoublePrecisionField)
	c4.Valid = true
	c4.Int16 = int16(a.Int2Field)
	c5.Valid = true
	c5.Int32 = int32(a.IntField)
	c6.Valid = true
	c6.Float64 = float64(a.RealField)
	c7.Valid = true
	c7.Int16 = int16(a.SmallintField)
	c8.Valid = true
	c8.String = a.TextField
	c9.Valid = true
	c9.Time = a.TimestampField
	c10.Valid = true
	c10.V = a.UbigintField
	c11.Valid = true
	c11.Int64 = int64(a.UintegerField) // uinteger_field
	return NewExample1ScannerStaticRow(q.QueryRowContext(ctx, query, c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11))
}

// Example1Insert2StaticColumns generated by genieql
const Example1Insert2StaticColumns = `$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,DEFAULT`

// Example1Insert2Explode generated by genieql
func Example1Insert2Explode(a *Example1) ([]interface{}, error) {
	var (
		c0  sql.NullInt64    // bigint_field
		c1  sql.NullBool     // bool_field
		c2  []byte           // byte_array_field
		c3  sql.NullFloat64  // double_precision_field
		c4  sql.NullInt16    // int2_field
		c5  sql.NullInt32    // int_field
		c6  sql.NullFloat64  // real_field
		c7  sql.NullInt16    // smallint_field
		c8  sql.NullString   // text_field
		c9  sql.NullTime     // timestamp_field
		c10 sql.Null[uint64] // ubigint_field
		c11 sql.NullInt64    // uinteger_field
	)

	c0.Valid = true
	c0.Int64 = int64(a.BigintField)

	c1.Valid = true
	c1.Bool = a.BoolField

	c2 = a.ByteArrayField

	c3.Valid = true
	c3.Float64 = float64(a.DoublePrecisionField)

	c4.Valid = true
	c4.Int16 = int16(a.Int2Field)

	c5.Valid = true
	c5.Int32 = int32(a.IntField)

	c6.Valid = true
	c6.Float64 = float64(a.RealField)

	c7.Valid = true
	c7.Int16 = int16(a.SmallintField)

	c8.Valid = true
	c8.String = a.TextField

	c9.Valid = true
	c9.Time = a.TimestampField

	c10.Valid = true
	c10.V = a.UbigintField

	c11.Valid = true
	c11.Int64 = int64(a.UintegerField)

	return []interface{}{c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11}, nil
}

// Example1Insert2 generated by genieql
func Example1Insert2(ctx context.Context, q sqlx.Queryer, a Example1) Example1ScannerStaticRow {
	const query = `INSERT INTO "example1" ("bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,DEFAULT) RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field"`
	var (
		c0  sql.NullInt64    // bigint_field
		c1  sql.NullBool     // bool_field
		c2  []byte           // byte_array_field
		c3  sql.NullFloat64  // double_precision_field
		c4  sql.NullInt16    // int2_field
		c5  sql.NullInt32    // int_field
		c6  sql.NullFloat64  // real_field
		c7  sql.NullInt16    // smallint_field
		c8  sql.NullString   // text_field
		c9  sql.NullTime     // timestamp_field
		c10 sql.Null[uint64] // ubigint_field
		c11 sql.NullInt64
	)
	c0.Valid = true
	c0.Int64 = int64(a.BigintField)
	c1.Valid = true
	c1.Bool = a.BoolField
	c2 = a.ByteArrayField
	c3.Valid = true
	c3.Float64 = float64(a.DoublePrecisionField)
	c4.Valid = true
	c4.Int16 = int16(a.Int2Field)
	c5.Valid = true
	c5.Int32 = int32(a.IntField)
	c6.Valid = true
	c6.Float64 = float64(a.RealField)
	c7.Valid = true
	c7.Int16 = int16(a.SmallintField)
	c8.Valid = true
	c8.String = a.TextField
	c9.Valid = true
	c9.Time = a.TimestampField
	c10.Valid = true
	c10.V = a.UbigintField
	c11.Valid = true
	c11.Int64 = int64(a.UintegerField) // uinteger_field
	return NewExample1ScannerStaticRow(q.QueryRowContext(ctx, query, c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11))
}

// Example1Insert3StaticColumns generated by genieql
const Example1Insert3StaticColumns = `$1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,DEFAULT`

// Example1Insert3Explode generated by genieql
func Example1Insert3Explode(a *Example1) ([]interface{}, error) {
	var (
		c0  sql.NullInt64    // bigint_field
		c1  sql.NullBool     // bool_field
		c2  []byte           // byte_array_field
		c3  sql.NullFloat64  // double_precision_field
		c4  sql.NullInt16    // int2_field
		c5  sql.NullInt32    // int_field
		c6  sql.NullFloat64  // real_field
		c7  sql.NullInt16    // smallint_field
		c8  sql.NullString   // text_field
		c9  sql.NullTime     // timestamp_field
		c10 sql.Null[uint64] // ubigint_field
		c11 sql.NullInt64    // uinteger_field
	)

	c0.Valid = true
	c0.Int64 = int64(a.BigintField)

	c1.Valid = true
	c1.Bool = a.BoolField

	c2 = a.ByteArrayField

	c3.Valid = true
	c3.Float64 = float64(a.DoublePrecisionField)

	c4.Valid = true
	c4.Int16 = int16(a.Int2Field)

	c5.Valid = true
	c5.Int32 = int32(a.IntField)

	c6.Valid = true
	c6.Float64 = float64(a.RealField)

	c7.Valid = true
	c7.Int16 = int16(a.SmallintField)

	c8.Valid = true
	c8.String = a.TextField

	c9.Valid = true
	c9.Time = a.TimestampField

	c10.Valid = true
	c10.V = a.UbigintField

	c11.Valid = true
	c11.Int64 = int64(a.UintegerField)

	return []interface{}{c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11}, nil
}

// Example1Insert3 generated by genieql
func Example1Insert3(ctx context.Context, q sqlx.Queryer, id int, a Example1) Example1ScannerStaticRow {
	const query = `INSERT INTO "example1" ("bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field") VALUES ($2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,DEFAULT) ON CONFLICT id = $1 AND b = $2 WHERE id = $1 RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field"`
	var (
		c0  sql.NullInt64    // id
		c1  sql.NullInt64    // bigint_field
		c2  sql.NullBool     // bool_field
		c3  []byte           // byte_array_field
		c4  sql.NullFloat64  // double_precision_field
		c5  sql.NullInt16    // int2_field
		c6  sql.NullInt32    // int_field
		c7  sql.NullFloat64  // real_field
		c8  sql.NullInt16    // smallint_field
		c9  sql.NullString   // text_field
		c10 sql.NullTime     // timestamp_field
		c11 sql.Null[uint64] // ubigint_field
		c12 sql.NullInt64
	)
	c0.Valid = true
	c0.Int64 = int64(id)
	c1.Valid = true
	c1.Int64 = int64(a.BigintField)
	c2.Valid = true
	c2.Bool = a.BoolField
	c3 = a.ByteArrayField
	c4.Valid = true
	c4.Float64 = float64(a.DoublePrecisionField)
	c5.Valid = true
	c5.Int16 = int16(a.Int2Field)
	c6.Valid = true
	c6.Int32 = int32(a.IntField)
	c7.Valid = true
	c7.Float64 = float64(a.RealField)
	c8.Valid = true
	c8.Int16 = int16(a.SmallintField)
	c9.Valid = true
	c9.String = a.TextField
	c10.Valid = true
	c10.Time = a.TimestampField
	c11.Valid = true
	c11.V = a.UbigintField
	c12.Valid = true
	c12.Int64 = int64(a.UintegerField) // uinteger_field
	return NewExample1ScannerStaticRow(q.QueryRowContext(ctx, query, c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11, c12))
}

// Example1InsertBatch1 generated by genieql
func NewExample1InsertBatch1(ctx context.Context, q sqlx.Queryer, a ...Example1) Example1Scanner {
	return &example1InsertBatch1{ctx: ctx, q: q, remaining: a}
}

type example1InsertBatch1 struct {
	ctx       context.Context
	q         sqlx.Queryer
	remaining []Example1
	scanner   Example1Scanner
}

func (t *example1InsertBatch1) Scan(a *Example1) error {
	return t.scanner.Scan(a)
}

func (t *example1InsertBatch1) Err() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Err()
}

func (t *example1InsertBatch1) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *example1InsertBatch1) Next() bool {
	var advanced bool
	if t.scanner != nil && t.scanner.Next() {
		return true
	}
	if len(t.remaining) > 0 && t.Close() == nil {
		t.scanner, t.remaining, advanced = t.advance(t.remaining...)
		return advanced && t.scanner.Next()
	}
	return false
}

func (t *example1InsertBatch1) advance(a ...Example1) (Example1Scanner, []Example1, bool) {
	transform := func(a Example1) (c0 sql.NullInt64, c1 sql.NullBool, c2 []byte, c3 sql.NullFloat64, c4 sql.NullInt16, c5 sql.NullInt32, c6 sql.NullFloat64, c7 sql.NullInt16, c8 sql.NullString, c9 sql.NullTime, c10 sql.Null[uint64], c11 sql.NullInt64, c12 sql.NullString, err error) {
		c0.Valid = true
		c0.Int64 = int64(a.BigintField)
		c1.Valid = true
		c1.Bool = a.BoolField
		c2 = a.ByteArrayField
		c3.Valid = true
		c3.Float64 = float64(a.DoublePrecisionField)
		c4.Valid = true
		c4.Int16 = int16(a.Int2Field)
		c5.Valid = true
		c5.Int32 = int32(a.IntField)
		c6.Valid = true
		c6.Float64 = float64(a.RealField)
		c7.Valid = true
		c7.Int16 = int16(a.SmallintField)
		c8.Valid = true
		c8.String = a.TextField
		c9.Valid = true
		c9.Time = a.TimestampField
		c10.Valid = true
		c10.V = a.UbigintField
		c11.Valid = true
		c11.Int64 = int64(a.UintegerField)
		c12.Valid = true
		c12.String = a.UUIDField
		return c0, c1, c2, c3, c4, c5, c6, c7, c8, c9, c10, c11, c12, nil
	}
	switch len(a) {
	case 0:
		return nil, []Example1(nil), false
	case 1:
		const query = `INSERT INTO "example1" ("bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13) RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
		var (
			r0c0  sql.NullInt64
			r0c1  sql.NullBool
			r0c2  []byte
			r0c3  sql.NullFloat64
			r0c4  sql.NullInt16
			r0c5  sql.NullInt32
			r0c6  sql.NullFloat64
			r0c7  sql.NullInt16
			r0c8  sql.NullString
			r0c9  sql.NullTime
			r0c10 sql.Null[uint64]
			r0c11 sql.NullInt64
			r0c12 sql.NullString
			err   error
		)
		if r0c0, r0c1, r0c2, r0c3, r0c4, r0c5, r0c6, r0c7, r0c8, r0c9, r0c10, r0c11, r0c12, err = transform(a[0]); err != nil {
			return NewExample1ScannerStatic(nil, err), []Example1(nil), false
		}
		return NewExample1ScannerStatic(t.q.QueryContext(t.ctx, query, r0c0, r0c1, r0c2, r0c3, r0c4, r0c5, r0c6, r0c7, r0c8, r0c9, r0c10, r0c11, r0c12)), a[1:], true
	default:
		const query = `INSERT INTO "example1" ("bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13),($14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26) RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
		var (
			r0c0  sql.NullInt64
			r0c1  sql.NullBool
			r0c2  []byte
			r0c3  sql.NullFloat64
			r0c4  sql.NullInt16
			r0c5  sql.NullInt32
			r0c6  sql.NullFloat64
			r0c7  sql.NullInt16
			r0c8  sql.NullString
			r0c9  sql.NullTime
			r0c10 sql.Null[uint64]
			r0c11 sql.NullInt64
			r0c12 sql.NullString
			r1c0  sql.NullInt64
			r1c1  sql.NullBool
			r1c2  []byte
			r1c3  sql.NullFloat64
			r1c4  sql.NullInt16
			r1c5  sql.NullInt32
			r1c6  sql.NullFloat64
			r1c7  sql.NullInt16
			r1c8  sql.NullString
			r1c9  sql.NullTime
			r1c10 sql.Null[uint64]
			r1c11 sql.NullInt64
			r1c12 sql.NullString
			err   error
		)
		if r0c0, r0c1, r0c2, r0c3, r0c4, r0c5, r0c6, r0c7, r0c8, r0c9, r0c10, r0c11, r0c12, err = transform(a[0]); err != nil {
			return NewExample1ScannerStatic(nil, err), []Example1(nil), false
		}
		if r1c0, r1c1, r1c2, r1c3, r1c4, r1c5, r1c6, r1c7, r1c8, r1c9, r1c10, r1c11, r1c12, err = transform(a[1]); err != nil {
			return NewExample1ScannerStatic(nil, err), []Example1(nil), false
		}
		return NewExample1ScannerStatic(t.q.QueryContext(t.ctx, query, r0c0, r0c1, r0c2, r0c3, r0c4, r0c5, r0c6, r0c7, r0c8, r0c9, r0c10, r0c11, r0c12, r1c0, r1c1, r1c2, r1c3, r1c4, r1c5, r1c6, r1c7, r1c8, r1c9, r1c10, r1c11, r1c12)), []Example1(nil), false
	}
}

// Example1Update1 generated by genieql
func Example1Update1(ctx context.Context, q sqlx.Queryer, i int, camelCaseID int, snakeCase int, e1 Example1, e2 Example2) Example1ScannerStaticRow {
	const query = `UPDATE example1 SET WHERE bigint_field = $1 RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
	var c0 sql.NullInt64 // bigint_field
	c0.Valid = true
	c0.Int64 = int64(e1.BigintField)
	return NewExample1ScannerStaticRow(q.QueryRowContext(ctx, query, c0))
}

// Example1Update2 generated by genieql
func Example1Update2(ctx context.Context, q sqlx.Queryer, i int, camelCaseID int, snakeCase int, e1 Example1, e2 Example2) Example1Scanner {
	const query = `UPDATE example1 SET WHERE bigint_field = $1 RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
	var c0 sql.NullInt64 // bigint_field
	c0.Valid = true
	c0.Int64 = int64(e1.BigintField)
	return NewExample1ScannerStatic(q.QueryContext(ctx, query, c0))
}

// Example1Update3 generated by genieql
func Example1Update3(ctx context.Context, q sqlx.Queryer, i int, ts time.Time) Example1Scanner {
	const query = `UPDATE example2 SET WHERE id = $1 AND timestamp = $2 RETURNING "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field"`
	var (
		c0 sql.NullInt64 // i
		c1 sql.NullTime  // ts
	)
	c0.Valid = true
	c0.Int64 = int64(i)
	c1.Valid = true
	c1.Time = ts
	return NewExample1ScannerStatic(q.QueryContext(ctx, query, c0, c1))
}

// Example1FindByBigintField generated by genieql
// test simple function generation with field replacement
func Example1FindByBigintField(ctx context.Context, q sqlx.Queryer, p Example1) Example1Scanner {
	const query = `SELECT "bigint_field","bool_field","byte_array_field","double_precision_field","int2_field","int_field","real_field","smallint_field","text_field","timestamp_field","ubigint_field","uinteger_field","uuid_field" FROM example1 WHERE "id" = $2 AND "id" = $1`
	var (
		c0 sql.NullInt64 // bigint_field
		c1 sql.NullInt32 // int_field
	)
	c0.Valid = true
	c0.Int64 = int64(p.BigintField)
	c1.Valid = true
	c1.Int32 = int32(p.IntField)
	return NewExample1ScannerStatic(q.QueryContext(ctx, query, c0, c1))
}
