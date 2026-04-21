package duckdb

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math"
	"net/netip"
	"path/filepath"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/timex"
)

func ExampleExample1Insert() {
	var (
		res Example1
	)
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	db := errorsx.Must(sql.Open("duckdb", filepath.Join("..", "..", genieql.RelDir(), ".duckdb", "duck.db")))
	defer db.Close()

	uid := uuid.Must(uuid.NewV7()).String()
	ex := Example1{
		BigintField:    1,
		BoolField:      true,
		UUIDField:      uid,
		IntField:       2,
		RealField:      3.1,
		SmallintField:  4,
		TextField:      "hello world",
		UintegerField:  2,
		UbigintField:   math.MaxUint64,
		ByteArrayField: []byte{0x2},
		// IntervalField:  time.Minute,
		// Int2Array:      []int{9},
		InetField: netip.IPv4Unspecified(),
	}

	errorsx.MaybePanic(Example1Insert(ctx, db, ex).Scan(&res))

	fmt.Println(
		"uid", res.UUIDField == ex.UUIDField,
		"bigint", res.BigintField == ex.BigintField,
		"int", res.IntField == ex.IntField,
		"smallint", res.SmallintField == ex.SmallintField,
		"float", res.RealField == ex.RealField,
		"bool", res.BoolField == ex.BoolField,
		"text", res.TextField == ex.TextField,
		"uinteger", res.UintegerField == ex.UintegerField,
		"binary", bytes.Compare(res.ByteArrayField, ex.ByteArrayField),
		"ubigint", res.UbigintField == ex.UbigintField,
		// "int array", slices.Compare(res.Int2Array, ex.Int2Array),
		"ip", res.InetField,
	)
	// Output: uid true bigint true int true smallint true float true bool true text true uinteger true binary 0 ubigint true ip 0.0.0.0
}

func ExampleNewExample1BatchInsertWithDefaults() {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	db := errorsx.Must(sql.Open("duckdb", filepath.Join("..", "..", genieql.RelDir(), ".duckdb", "duck.db")))
	defer db.Close()

	records := []Example1{
		{BigintField: 10, BoolField: true, UUIDField: uuid.Must(uuid.NewV7()).String(), IntField: 1, RealField: 1.1, SmallintField: 1, TextField: "batch record 1", UintegerField: 1, UbigintField: 1, ByteArrayField: []byte{0x1}, InetField: netip.IPv4Unspecified()},
		{BigintField: 20, BoolField: false, UUIDField: uuid.Must(uuid.NewV7()).String(), IntField: 2, RealField: 2.2, SmallintField: 2, TextField: "batch record 2", UintegerField: 2, UbigintField: 2, ByteArrayField: []byte{0x2}, InetField: netip.IPv4Unspecified()},
		{BigintField: 30, BoolField: true, UUIDField: uuid.Must(uuid.NewV7()).String(), IntField: 3, RealField: 3.3, SmallintField: 3, TextField: "batch record 3", UintegerField: 3, UbigintField: 3, ByteArrayField: []byte{0x3}, InetField: netip.IPv4Unspecified()},
	}

	scanner := NewExample1BatchInsertWithDefaults(ctx, db, records...)
	defer scanner.Close()

	var count int
	for scanner.Next() {
		var res Example1
		errorsx.MaybePanic(scanner.Scan(&res))
		fmt.Println("text", res.TextField, "bigint", res.BigintField)
		count++
	}
	errorsx.MaybePanic(scanner.Err())
	fmt.Println("count", count)
	// Output:
	// text batch record 1 bigint 10
	// text batch record 2 bigint 20
	// text batch record 3 bigint 30
	// count 3
}

func ExampleNewExample1BatchInsertWithDefaults_multipleAdvances() {
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	db := errorsx.Must(sql.Open("duckdb", filepath.Join("..", "..", genieql.RelDir(), ".duckdb", "duck.db")))
	defer db.Close()

	// 35 records spans two advance calls: first batch of 32, then a batch of 3.
	records := make([]Example1, 35)
	for i := range records {
		records[i] = Example1{
			BigintField:    int64(i),
			UUIDField:      uuid.Must(uuid.NewV7()).String(),
			TextField:      fmt.Sprintf("multi %d", i),
			ByteArrayField: []byte{byte(i)},
			InetField:      netip.IPv4Unspecified(),
		}
	}

	scanner := NewExample1BatchInsertWithDefaults(ctx, db, records...)
	defer scanner.Close()

	var count int
	for scanner.Next() {
		var res Example1
		errorsx.MaybePanic(scanner.Scan(&res))
		count++
	}
	errorsx.MaybePanic(scanner.Err())
	fmt.Println("inserted", count)
	// Output: inserted 35
}

func ExampleExample1UpdateTime() {
	var (
		res Example1
	)
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	db := errorsx.Must(sql.Open("duckdb", filepath.Join("..", "..", genieql.RelDir(), ".duckdb", "duck.db")))
	defer db.Close()

	uid := uuid.Must(uuid.NewV7()).String()
	ex := Example1{
		BigintField:    1,
		BoolField:      true,
		UUIDField:      uid,
		IntField:       2,
		RealField:      3.1,
		SmallintField:  4,
		TextField:      "hello world",
		UintegerField:  2,
		UbigintField:   math.MaxUint64,
		ByteArrayField: []byte{0x2},
		TimestampField: time.Date(2025, time.September, 10, 0, 0, 0, 0, time.UTC),
		InetField:      netip.IPv6LinkLocalAllNodes(),
	}

	errorsx.MaybePanic(Example1Insert(ctx, db, ex).Scan(&res))
	fmt.Println(
		"timestamp", res.TimestampField,
	)

	ex.TimestampField = timex.Inf()
	errorsx.MaybePanic(Example1UpdateTime(ctx, db, ex).Scan(&res))

	fmt.Println(
		"timestamp", res.TimestampField.UTC(),
	)

	ex.TimestampField = timex.NegInf()
	errorsx.MaybePanic(Example1UpdateTime(ctx, db, ex).Scan(&res))

	fmt.Println(
		"timestamp", res.TimestampField.UTC(),
	)

	// Output:
	// timestamp 2025-09-10 00:00:00 +0000 UTC
	// timestamp 292277024627-12-06 15:30:07.999999999 +0000 UTC
	// timestamp 292277026304-08-26 15:42:51.145224192 +0000 UTC
}
