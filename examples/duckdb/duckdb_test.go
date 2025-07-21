package duckdb

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"math"
	"path/filepath"
	"time"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
	_ "github.com/marcboeker/go-duckdb/v2"
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
		// InetField: netip.IPv6LinkLocalAllNodes(),
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
		// "ip", res.InetField.Compare(netip.IPv6LinkLocalAllNodes()) == 0,
	)
	// Output: uid true bigint true int true smallint true float true bool true text true uinteger true binary 0 ubigint true
}
