//go:build genieql.generate
// +build genieql.generate

package pkga

import (
	genieql "github.com/james-lawrence/genieql/ginterp"
)

// Example1 ...
func Example1(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}
