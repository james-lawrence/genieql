package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
)

var _ = Describe("ImplementationError", func() {
	var buffer *bytes.Buffer
	var fset *token.FileSet

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		fset = token.NewFileSet()
	})

	Describe("Generate", func() {
		It("should generate a error scanner implementation", func() {
			decl := errorScannerImplementation{}.Generate("ErrorScanner", &ast.Field{
				Names: []*ast.Ident{{Name: "arg0"}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: "CustomType"}},
			})

			Expect(printer.Fprint(buffer, fset, decl)).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(ReadString("test_fixtures/implementation_error.txt")))
		})
	})
})
