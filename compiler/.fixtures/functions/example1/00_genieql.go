//go:build genieql.generate
// +build genieql.generate

package example1

import (
	genieql "bitbucket.org/jatone/genieql/interp"
	"github.com/jackc/pgtype"
)

func Example1(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}

func Example2(gql genieql.Structure) {
	gql.From(
		gql.Table("example2"),
	)
}

func Example1Scanner(genieql.Scanner, func(i Example1)) {}

func ExampleComboScanner(
	gql genieql.Scanner,
	pattern func(i int, ts time.Time, e1 Example1, e2 Example2),
) {
}

func Example1Insert1(
	gql genieql.Insert,
	pattern func(ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow,
) {
	gql.Into("example1").Default("uuid_field")
}

func Example1Insert2(
	gql genieql.Insert,
	pattern func(ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow,
) {
	gql.Into("example1").Ignore("uuid_field")
}

func Example1Insert3(
	gql genieql.Insert,
	pattern func(ctx context.Context, q sqlx.Queryer, id int, a Example1) NewExample1ScannerStaticRow,
) {
	gql.Into("example1").Ignore("uuid_field").Conflict("ON CONFLICT id = {id} AND b = {a.B} WHERE id = {id}")
}

func Example1InsertBatch1(
	gql genieql.InsertBatch,
	pattern func(ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStatic,
) {
	gql.Into("example1").Batch(2)
}

func Example1Update1(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, i int, camelCaseID int, snake_case int, e1 Example1, e2 Example2) NewExample1ScannerStaticRow,
) {
	gql = gql.Query(`UPDATE example2 SET WHERE bigint_field = {e1.BigintField} RETURNING ` + Example1ScannerStaticColumns)
}

func Example1Update2(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, i int, camelCaseID int, snake_case int, e1 Example1, e2 Example2) NewExample1ScannerStatic,
) {
	gql = gql.Query(`UPDATE example2 SET WHERE bigint_field = {e1.BigintField} RETURNING ` + Example1ScannerStaticColumns)
}

func Example1Update3(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, i int, ts pgtype.Timestamp) NewExample1ScannerStatic,
) {
	gql = gql.Query(`UPDATE example2 SET WHERE id = {i} AND timestamp = {ts} RETURNING ` + Example1ScannerStaticColumns)
}
