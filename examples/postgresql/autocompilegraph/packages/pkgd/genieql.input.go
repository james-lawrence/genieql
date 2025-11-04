//go:build genieql.generate

package pkgd

import (
	"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/packages/pkgc"
	genieql "github.com/james-lawrence/genieql/ginterp"
)

func PackageDExample(gql genieql.Structure) {
	gql.From(
		gql.Query("SELECT 2 as id"),
	)
}

func PackageDScanner(genieql.Scanner, func(c pkgc.PackageCExample)) {}
