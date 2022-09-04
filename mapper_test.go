package genieql_test

import (
	"go/ast"

	. "bitbucket.org/jatone/genieql"
	"golang.org/x/text/transform"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"bitbucket.org/jatone/genieql/astutil"
)

var _ = Describe("Mapper", func() {
	DescribeTable("MapFieldToColumn",
		func(column string, field *ast.Field, aliaser transform.Transformer) {
			matchFound := MapFieldToNativeType(ColumnInfo{Name: column}, field, aliaser)
			Expect(matchFound).ToNot(BeNil())
		},
		Entry(
			"example 1 - simple match",
			"column1",
			astutil.Field(ast.NewIdent("int"), ast.NewIdent("Column1")),
			AliasStrategyCamelcase,
		),
	)
})
