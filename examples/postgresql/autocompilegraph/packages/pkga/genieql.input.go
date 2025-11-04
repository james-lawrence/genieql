//go:build genieql.generate

package pkga

import (
	genieql "github.com/james-lawrence/genieql/ginterp"
)

func PackageAExample(gql genieql.Structure) {
	gql.From(
		gql.Table("example1"),
	)
}
