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
				matchFound, err := MapFieldToColumn(example.column, example.offset, example.field, example.Aliaser)
				Expect(err).ToNot(HaveOccurred())
				Expect(matchFound).To(BeTrue())
			}
		})
	})

	Describe("MapColumns", func() {
		examples := []struct {
			arg     *ast.Ident
			columns []string
			fields  []*ast.Field
		}{
			{
				arg:     &ast.Ident{Name: "arg0"},
				columns: []string{"column1", "column2"},
				fields: []*ast.Field{
					&ast.Field{Names: []*ast.Ident{&ast.Ident{Name: "Column1"}}, Type: &ast.Ident{Name: "int"}},
				},
			},
		}
		It("should return mapped columns for the given fields", func() {
			for _, example := range examples {
				columnMaps, err := Mapper{Aliasers: []Aliaser{AliasStrategySnakecase}}.UnmappedColumns(example.fields, example.columns...)
				Expect(err).ToNot(HaveOccurred())
				for idx, m := range columnMaps {
					Expect(m).To(Equal(example.columns[idx]))
				}
			}
		})
	})
})
