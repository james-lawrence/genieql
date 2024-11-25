package duckdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql/internal/errorsx"

	_ "github.com/marcboeker/go-duckdb"
)

func ExampleExample1Insert() {
	var (
		res Example1
	)
	ctx, done := context.WithTimeout(context.Background(), 5*time.Second)
	defer done()

	db := errorsx.Must(sql.Open("duckdb", "duck.db"))
	defer db.Close()

	uid := uuid.Must(uuid.NewV7()).String()
	ex := Example1{
		BigintField:   1,
		BoolField:     true,
		UUIDField:     uid,
		IntField:      2,
		RealField:     3.1,
		SmallintField: 4,
		TextField:     "hello world",
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
	)
	// Output: uid true bigint true int true smallint true float true bool true text true
}
