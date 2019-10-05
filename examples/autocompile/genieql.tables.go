// +build genieql,autogenerate

package autocompile

import (
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

// // CustomScanner generates a scanner that consumes the given parameters.
// func CustomScanner(gql genieql.Scanner, pattern func(i1, i2 int, b1 bool, t1 time.Time)) {}
//
// // Example1Scanner generates a scanner that consumes the given parameters.
// func Example1Scanner(gql genieql.Scanner, pattern func(Example1)) {}
//
// // Example2Scanner generates a scanner that consumes the given parameters.
// func Example1Scanner(gql genieql.Scanner, e2 Example1) {}
//
// // CombinedScanner generates a scanner that consumes the given parameters.
// func CombinedScanner(gql genieql.Scanner, e1 Example1, e2 Example2) {}
