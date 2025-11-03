//go:build genieql.generate

package autocompilegraph

import (
	"github.com/james-lawrence/genieql/examples/postgresql/autocompilegraph/pkgd"
	genieql "github.com/james-lawrence/genieql/ginterp"
)

func RootExample(gql genieql.Structure) {
	gql.From(
		gql.Query("SELECT 3 as id"),
	)
}

func RootScanner(genieql.Scanner, func(d pkgd.PackageDExample)) {}
