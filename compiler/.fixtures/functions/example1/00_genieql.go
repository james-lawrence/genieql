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

func Example1Scanner(genieql.Scanner, func(i Example1)) {}

func Example1InsertWithDefaults1(gql genieql.Insert, ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow {
	gql.Into("example1").Default("uuid_field")
}

func Example1InsertWithDefaults2(gql genieql.Insert, ctx context.Context, q sqlx.Queryer, a Example1) NewExample1ScannerStaticRow {
	gql.Into("example1").Ignore("uuid_field")
}
