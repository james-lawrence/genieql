//go:build genieql.generate

package pkgb

import (
	genieql "github.com/james-lawrence/genieql/ginterp"
)

func PackageBExample(gql genieql.Structure) {
	gql.From(
		gql.Table("example2"),
	)
}
