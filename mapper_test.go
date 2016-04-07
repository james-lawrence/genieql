package genieql_test

import (
	"fmt"
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
				Aliaser: AliasStrategyLowercase,
			},
		}

		It("should return a mapped column if the column matches the field and its aliases", func() {
			for _, example := range examples {
				mappedColumn, matchFound, err := MapFieldToColumn(example.column, example.offset, example.field, example.Aliaser)
				Expect(err).ToNot(HaveOccurred())
				Expect(matchFound).To(BeTrue())
				Expect(mappedColumn.ColumnName).To(Equal(example.column))
				Expect(mappedColumn.ColumnOffset).To(Equal(example.offset))
				Expect(mappedColumn.Type).To(Equal(example.field.Type))
				Expect(mappedColumn.AssignmentExpr(example.arg)).To(Equal(&ast.SelectorExpr{
					X:   example.arg,
					Sel: example.field.Names[0],
				}))
				Expect(mappedColumn.LocalVariableExpr()).To(Equal(&ast.Ident{Name: fmt.Sprintf("c%d", example.offset)}))
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
				columnMaps, err := Mapper{Aliasers: []Aliaser{AliasStrategySnakecase}}.MapColumns(example.fields, example.columns...)
				Expect(err).ToNot(HaveOccurred())
				for idx, m := range columnMaps {
					Expect(m.ColumnName).To(Equal(example.columns[idx]))
					Expect(m.ColumnOffset).To(Equal(idx))
					Expect(m.FieldName).To(Equal(example.fields[idx].Names[0].Name))
					Expect(m.Type).To(Equal(example.fields[idx].Type))
				}
			}
		})
	})
})
