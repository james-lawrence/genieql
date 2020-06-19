package example

import (
	"database/sql"

	"bitbucket.org/jatone/genieql/internal/sqlx"
)

// NewBatchFunction4 creates a scanner that inserts a batch of
// records into the database.
func NewBatchFunction4(q sqlx.Queryer, p ...StructA) ExampleScanner {
	return &batchFunction4{
		q:         q,
		remaining: p,
	}
}

type batchFunction4 struct {
	q         sqlx.Queryer
	remaining []StructA
	scanner   ExampleScanner
}

func (t *batchFunction4) Scan(dst *StructA) error {
	return t.scanner.Scan(dst)
}

func (t *batchFunction4) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *batchFunction4) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *batchFunction4) Next() bool {
	var (
		advanced bool
	)

	if t.scanner != nil && t.scanner.Next() {
		return true
	}

	// advance to the next check
	if len(t.remaining) > 0 && t.Close() == nil {
		t.scanner, t.remaining, advanced = t.advance(t.q, t.remaining...)
		return advanced && t.scanner.Next()
	}

	return false
}

func (t *batchFunction4) advance(q sqlx.Queryer, p ...StructA) (ExampleScanner, []StructA, bool) {
	switch len(p) {
	case 0:
		return nil, []StructA(nil), false
	case 1:
		const query = `QUERY 1`
		exploder := func(p ...StructA) (r [8]interface{}) {
			for idx, v := range p[:1] {
				var (
					c0 sql.NullInt64
					c1 sql.NullInt64
					c2 sql.NullInt64
					c3 sql.NullBool
					c4 sql.NullBool
					c5 sql.NullBool
					c6 sql.NullInt64
					c7 sql.NullBool
				)
				c0.Valid = true
				c0.Int64 = int64(v.A)
				c1.Valid = true
				c1.Int64 = int64(v.B)
				c2.Valid = true
				c2.Int64 = int64(v.C)
				c3.Valid = true
				c3.Bool = v.D
				c4.Valid = true
				c4.Bool = v.E
				c5.Valid = true
				c5.Bool = v.F
				c6.Valid = true
				c6.Int64 = int64(*v.G)
				c7.Valid = true
				c7.Bool = *v.H
				r[idx*8+0], r[idx*8+1], r[idx*8+2], r[idx*8+3], r[idx*8+4], r[idx*8+5], r[idx*8+6], r[idx*8+7] = c0, c1, c2, c3, c4, c5, c6, c7
			}
			return r
		}
		tmp := exploder(p...)
		return StaticExampleScanner(q.Query(query, tmp[:]...)), []StructA(nil), true
	case 2:
		const query = `QUERY 2`
		exploder := func(p ...StructA) (r [16]interface{}) {
			for idx, v := range p[:2] {
				var (
					c0 sql.NullInt64
					c1 sql.NullInt64
					c2 sql.NullInt64
					c3 sql.NullBool
					c4 sql.NullBool
					c5 sql.NullBool
					c6 sql.NullInt64
					c7 sql.NullBool
				)
				c0.Valid = true
				c0.Int64 = int64(v.A)
				c1.Valid = true
				c1.Int64 = int64(v.B)
				c2.Valid = true
				c2.Int64 = int64(v.C)
				c3.Valid = true
				c3.Bool = v.D
				c4.Valid = true
				c4.Bool = v.E
				c5.Valid = true
				c5.Bool = v.F
				c6.Valid = true
				c6.Int64 = int64(*v.G)
				c7.Valid = true
				c7.Bool = *v.H
				r[idx*8+0], r[idx*8+1], r[idx*8+2], r[idx*8+3], r[idx*8+4], r[idx*8+5], r[idx*8+6], r[idx*8+7] = c0, c1, c2, c3, c4, c5, c6, c7
			}
			return r
		}
		tmp := exploder(p...)
		return StaticExampleScanner(q.Query(query, tmp[:]...)), []StructA(nil), true
	case 3:
		const query = `QUERY 3`
		exploder := func(p ...StructA) (r [24]interface{}) {
			for idx, v := range p[:3] {
				var (
					c0 sql.NullInt64
					c1 sql.NullInt64
					c2 sql.NullInt64
					c3 sql.NullBool
					c4 sql.NullBool
					c5 sql.NullBool
					c6 sql.NullInt64
					c7 sql.NullBool
				)
				c0.Valid = true
				c0.Int64 = int64(v.A)
				c1.Valid = true
				c1.Int64 = int64(v.B)
				c2.Valid = true
				c2.Int64 = int64(v.C)
				c3.Valid = true
				c3.Bool = v.D
				c4.Valid = true
				c4.Bool = v.E
				c5.Valid = true
				c5.Bool = v.F
				c6.Valid = true
				c6.Int64 = int64(*v.G)
				c7.Valid = true
				c7.Bool = *v.H
				r[idx*8+0], r[idx*8+1], r[idx*8+2], r[idx*8+3], r[idx*8+4], r[idx*8+5], r[idx*8+6], r[idx*8+7] = c0, c1, c2, c3, c4, c5, c6, c7
			}
			return r
		}
		tmp := exploder(p...)
		return StaticExampleScanner(q.Query(query, tmp[:]...)), []StructA(nil), true
	case 4:
		const query = `QUERY 4`
		exploder := func(p ...StructA) (r [32]interface{}) {
			for idx, v := range p[:4] {
				var (
					c0 sql.NullInt64
					c1 sql.NullInt64
					c2 sql.NullInt64
					c3 sql.NullBool
					c4 sql.NullBool
					c5 sql.NullBool
					c6 sql.NullInt64
					c7 sql.NullBool
				)
				c0.Valid = true
				c0.Int64 = int64(v.A)
				c1.Valid = true
				c1.Int64 = int64(v.B)
				c2.Valid = true
				c2.Int64 = int64(v.C)
				c3.Valid = true
				c3.Bool = v.D
				c4.Valid = true
				c4.Bool = v.E
				c5.Valid = true
				c5.Bool = v.F
				c6.Valid = true
				c6.Int64 = int64(*v.G)
				c7.Valid = true
				c7.Bool = *v.H
				r[idx*8+0], r[idx*8+1], r[idx*8+2], r[idx*8+3], r[idx*8+4], r[idx*8+5], r[idx*8+6], r[idx*8+7] = c0, c1, c2, c3, c4, c5, c6, c7
			}
			return r
		}
		tmp := exploder(p...)
		return StaticExampleScanner(q.Query(query, tmp[:]...)), []StructA(nil), true
	default:
		const query = `QUERY 5`
		exploder := func(p ...StructA) (r [40]interface{}) {
			for idx, v := range p[:5] {
				var (
					c0 sql.NullInt64
					c1 sql.NullInt64
					c2 sql.NullInt64
					c3 sql.NullBool
					c4 sql.NullBool
					c5 sql.NullBool
					c6 sql.NullInt64
					c7 sql.NullBool
				)
				c0.Valid = true
				c0.Int64 = int64(v.A)
				c1.Valid = true
				c1.Int64 = int64(v.B)
				c2.Valid = true
				c2.Int64 = int64(v.C)
				c3.Valid = true
				c3.Bool = v.D
				c4.Valid = true
				c4.Bool = v.E
				c5.Valid = true
				c5.Bool = v.F
				c6.Valid = true
				c6.Int64 = int64(*v.G)
				c7.Valid = true
				c7.Bool = *v.H
				r[idx*8+0], r[idx*8+1], r[idx*8+2], r[idx*8+3], r[idx*8+4], r[idx*8+5], r[idx*8+6], r[idx*8+7] = c0, c1, c2, c3, c4, c5, c6, c7
			}
			return r
		}
		tmp := exploder(p[:5]...)
		return StaticExampleScanner(q.Query(query, tmp[:]...)), p[5:], true
	}
}
