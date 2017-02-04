package genieql_test

import (
	"go/ast"

	. "bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"bitbucket.org/jatone/genieql/astutil"
)

var _ = Describe("Mapper", func() {
	DescribeTable("MapFieldToColumn",
		func(column string, field *ast.Field, aliaser Aliaser) {
			matchFound := MapFieldToColumn(column, field, aliaser)
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
