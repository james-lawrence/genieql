//go:build genieql.generate
// +build genieql.generate

package pkga

import (
	genieql "bitbucket.org/jatone/genieql/interp/genieql"
)

// Example1 ...
func Example1(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}
