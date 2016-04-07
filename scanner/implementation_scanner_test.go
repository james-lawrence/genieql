package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"go/ast"
	"go/printer"
	"go/token"

	"bitbucket.org/jatone/genieql"
)

var _ = Describe("ImplementationScanner", func() {
	var buffer *bytes.Buffer
	var fset *token.FileSet

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		fset = token.NewFileSet()
	})

	Describe("Generate", func() {
		It("should generate a row scanner implementation", func() {
			decl := scannerImplementation{
				ColumnMaps: []genieql.ColumnMap{
					genieql.ColumnMap{
						ColumnName:   "column1",
						ColumnOffset: 1,
						FieldName:    "Field1",
						Type:         &ast.Ident{Name: "bool"},
					},
					genieql.ColumnMap{
						ColumnName:   "column2",
						ColumnOffset: 2,
						FieldName:    "Field2",
						Type:         &ast.StarExpr{X: &ast.Ident{Name: "bool"}},
					},
				},
				Driver: genieql.NewDriver(DefaultNullableTypes, DefaultLookupNullableType),
			}.Generate("Scanner", &ast.Field{
				Names: []*ast.Ident{{Name: "arg0"}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: "CustomType"}},
			})

			Expect(printer.Fprint(buffer, fset, decl)).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(ReadString("test_fixtures/implementation_scanner.txt")))
		})
	})
})
