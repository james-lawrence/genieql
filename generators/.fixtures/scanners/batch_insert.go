package example

func newBatchInsertScanner(q sqlx.Queryer, p ...int) Int {
	return batchInsertIntScanner{
		q:         q,
		remaining: p,
	}
}

type batchInsertIntScanner struct {
	q         sqlx.Queryer
	remaining []int
	scanner   Int
}

func (t *batchInsertIntScanner) Scan(arg *int) error {
	return t.scanner.Scan(&dst)
}

func (t *batchInsertIntScanner) Next() bool {
	var (
		advanced bool
	)

	if t.scanner != nil && t.scanner.Next() {
		return true
	}

	// advance to the next check
	if len(remaining) > 0 && t.Close() == nil {
		t.scanner, t.remaining, advanced = t.advance(t.q, t.remaining...)
		return advanced && t.scanner.Next()
	}

	return false
}

func (t *batchInsertIntScanner) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *batchInsertIntScanner) Close() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Close()
}

func (t *batchInsertIntScanner) advance(q sqlx.Queryer, p ...int) (scanner, []int, bool) {
	switch len(p) {
	case 0:
		return nil, []int(nil), false
	case 1:
		const query = `QUERY 1`
		exploder := func(p ...Example1) (r [5]interface{}) {
			for idx, v := range p[:1] {
				r[idx*5+0], r[idx*5+1], r[idx*5+2], r[idx*5+3], r[idx*5+4] = v.CreatedAt, v.ID, v.TextField, v.UpdatedAt, v.UUIDField
			}
			return
		}
		tmp := exploder(p...)
		return NewIntScannerStatic(q.Query(query, tmp[:]...)), []int(nil), true
	case 2:
		const query = `QUERY 2`
		exploder := func(p ...Example1) (r [10]interface{}) {
			for idx, v := range p[:2] {
				r[idx*5+0], r[idx*5+1], r[idx*5+2], r[idx*5+3], r[idx*5+4] = v.CreatedAt, v.ID, v.TextField, v.UpdatedAt, v.UUIDField
			}
			return
		}
		tmp := exploder(p...)
		return NewIntScannerStatic(q.Query(query, tmp[:]...)), []int(nil), true
	case 3:
		const query = `QUERY 3`
		exploder := func(p ...Example1) (r [15]interface{}) {
			for idx, v := range p[:3] {
				r[idx*5+0], r[idx*5+1], r[idx*5+2], r[idx*5+3], r[idx*5+4] = v.CreatedAt, v.ID, v.TextField, v.UpdatedAt, v.UUIDField
			}
			return
		}
		tmp := exploder(p...)
		return NewIntScannerStatic(q.Query(query, tmp[:]...)), []int(nil), true
	case 4:
		const query = `QUERY 4`
		exploder := func(p ...Example1) (r [20]interface{}) {
			for idx, v := range p[:4] {
				r[idx*5+0], r[idx*5+1], r[idx*5+2], r[idx*5+3], r[idx*5+4] = v.CreatedAt, v.ID, v.TextField, v.UpdatedAt, v.UUIDField
			}
			return
		}
		tmp := exploder(p...)
		return NewIntScannerStatic(q.Query(query, tmp[:]...)), []int(nil), true
	default:
		const query = `QUERY 5`
		exploder := func(p ...Example1) (r [25]interface{}) {
			for idx, v := range p[:5] {
				r[idx*5+0], r[idx*5+1], r[idx*5+2], r[idx*5+3], r[idx*5+4] = v.CreatedAt, v.ID, v.TextField, v.UpdatedAt, v.UUIDField
			}
			return
		}
		tmp := exploder(p[:5]...)
		return NewIntScannerStatic(q.Query(query, tmp[:]...)), p[5:], true
	}
}
