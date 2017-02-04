package genieql_test

import (
	"go/ast"

	. "bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Mapper", func() {
	Describe("MapFieldToColumn", func() {
		examples := []struct {
			arg    *ast.Ident
			column string
			field  *ast.Field
			offset int
			Aliaser
		}{
			{
				arg:     &ast.Ident{Name: "arg0"},
				column:  "column1",
				field:   &ast.Field{Names: []*ast.Ident{&ast.Ident{Name: "Column1"}}, Type: &ast.Ident{Name: "int"}},
				offset:  0,
				Aliaser: AliasStrategyCamelcase,
			},
		}

		It("should return true if the column matches the field and its aliases", func() {
			for _, example := range examples {
				matchFound := MapFieldToColumn(example.column, example.offset, example.field, example.Aliaser)
				Expect(matchFound).ToNot(BeNil())
			}
		})
	})
})
