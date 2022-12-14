//go:build genieql.generate
// +build genieql.generate

package example1

import (
	genieql "bitbucket.org/jatone/genieql/interp"
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

func Example1Insert1(gql genieql.Insert, ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow {
	gql.Into("example1").Default("uuid_field")
}

func Example1Insert2(gql genieql.Insert, ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow {
	gql.Into("example1").Ignore("uuid_field")
}

func Example1Update(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, i int, e1 Example1, e2 Example2) NewExample1ScannerStaticRow,
) {
	gql = gql.Query(`UPDATE example2 SET WHERE bigint_field = {e1.BigintField.query.input} RETURNING ` + Example1ScannerStaticColumns)
}
