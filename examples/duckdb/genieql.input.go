//go:build genieql.generate
// +build genieql.generate

package duckdb

import (
	"context"

	genieql "github.com/james-lawrence/genieql/ginterp"
	"github.com/james-lawrence/genieql/internal/sqlx"
)

// Example1 ...
func Example1(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}

// generates a scanner that consumes the given parameters.
func Example1Scanner(genieql.Scanner, func(Example1)) {}

func Example1FindBy(gql genieql.QueryAutogen, ctx context.Context, q sqlx.Queryer, e Example1) NewExample1ScannerStaticRow {
	gql.From("example1").Ignore("created_at", "updated_at", "id")
}

// insert a single example1 record.
func Example1Insert(
	gql genieql.Insert,
	pattern func(ctx context.Context, q sqlx.Queryer, e Example1) NewExample1ScannerStaticRow,
) {
	gql.Into("example1")
}

func Example1FindByID(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, id string) NewExample1ScannerStaticRow,
) {
	gql = gql.Query(`SELECT ` + Example1ScannerStaticColumns + ` FROM example1 WHERE "uuid_field" = {id}`)
}
