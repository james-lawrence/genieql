//go:build genieql.generate

package pkgc

import (
	"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkga"
	"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgb"
	genieql "github.com/james-lawrence/genieql/ginterp"
)

func PackageCExample(gql genieql.Structure) {
	gql.From(
		gql.Query("SELECT 1 as id"),
	)
}

func PackageCScanner(genieql.Scanner, func(a pkga.PackageAExample, b pkgb.PackageBExample)) {}
