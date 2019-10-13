// +build genieql.generate

package autocompile

import (
	"context"
	"time"

	"bitbucket.org/jatone/genieql/genieql"
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

// CustomScanner generates a scanner that consumes the given parameters.
func CustomScanner(gql genieql.Scanner, output func(i1, i2 int, b1 bool, t1 time.Time)) {}

// Example1Scanner generates a scanner that consumes the given parameters.
func Example1Scanner(genieql.Scanner, func(Example1)) {}

// Example2Scanner generates a scanner that consumes the given parameters.
func Example2Scanner(genieql.Scanner, func(Example2)) {}

// CombinedScanner generates a scanner that consumes the given parameters.
func CombinedScanner(genieql.Scanner, func(e1 Example1, e2 Example2)) {}

// Example1FindByWhatever generates function
func Example1FindByWhatever(gql genieql.Function, pattern func(ctx context.Context, q sqlx.Queryer, i1, i2 int) NewExample1ScannerStaticRow) {
	gql.Query("SELECT " + Example1ScannerStaticColumns + " FROM example1")
}
