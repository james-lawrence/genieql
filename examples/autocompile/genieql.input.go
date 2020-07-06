// +build genieql.generate

package autocompile

import (
	"context"
	"time"

	"bitbucket.org/jatone/genieql/internal/sqlx"
	genieql "bitbucket.org/jatone/genieql/interp"
)

// Example1 ...
func Example1(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}

// Example2 ...
func Example2(gql genieql.Structure) {
	gql.From(
		gql.Query("SELECT * FROM example2"),
	)
}

// Example3 ...
func Example3(gql genieql.Structure) {
	gql.From(
		gql.Table("example2"),
	)
}

func Timestamp(gql genieql.Structure) {
	gql.From(
		gql.Table("timestamp_examples"),
	)
}

// CustomScanner generates a scanner that consumes the given parameters.
func CustomScanner(gql genieql.Scanner, output func(i1, i2 int, b1 bool, t1 time.Time)) {}

// Example1Scanner generates a scanner that consumes the given parameters.
func Example1Scanner(genieql.Scanner, func(Example1)) {}

// Example2Scanner generates a scanner that consumes the given parameters.
func Example2Scanner(genieql.Scanner, func(Example2)) {}

// CombinedScanner generates a scanner that consumes the given parameters.
func CombinedScanner(genieql.Scanner, func(e1 Example1, e2 Example2)) {}

// TimestampScanner generates a scanner that consumes the given parameters.
func TimestampScanner(genieql.Scanner, func(Timestamp)) {}

// Example1FindByX1 generates function
func Example1FindByX1(
	gql genieql.Function,
	pattern func(ctx context.Context, q sqlx.Queryer, i1, i2 int) NewExample1ScannerStaticRow,
) {
	gql = gql.Query("SELECT " + Example1ScannerStaticColumns + " FROM example1 WHERE id = $1 AND foo = $2")
}

func Example1FindBy(gql genieql.QueryAutogen, ctx context.Context, q sqlx.Queryer, e Example1) NewExample1ScannerStaticRow {
	gql.From("example1").Ignore("created_at", "updated_at", "id")
}

func Example1LookupBy(gql genieql.QueryAutogen, ctx context.Context, q sqlx.Queryer, e Example1) NewExample1ScannerStatic {
	gql.From("example1")
}

// Example1Insert insert a single example1 record.
func Example1Insert(gql genieql.Insert, ctx context.Context, q sqlx.Queryer, e Example1) NewExample1ScannerStaticRow {
	gql.Into("example1").Default("uuid_field")
}
