package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
)

var _ = Describe("Scanner", func() {
	var buffer *bytes.Buffer
	var fset *token.FileSet

	BeforeEach(func() {
		buffer = bytes.NewBuffer([]byte{})
		fset = token.NewFileSet()
	})

	Describe("BuildScannerInterface", func() {
		It("should build a correct interface", func() {
			f1 := &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "arg0"}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: "AType"}},
			}

			r := BuildScannerInterface("AScanner", f1)

			buffer := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()
			Expect(printer.Fprint(buffer, fset, r)).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(ReadString("test_fixtures/scanner_interface.txt")))
		})
	})

	Describe("BuildRowsScannerInterface", func() {
		It("should build a correct interface", func() {
			f1 := &ast.Field{
				Names: []*ast.Ident{&ast.Ident{Name: "arg0"}},
				Type:  &ast.StarExpr{X: &ast.Ident{Name: "AType"}},
			}

			r := BuildRowsScannerInterface("AScanner", f1)

			Expect(printer.Fprint(buffer, fset, r)).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(ReadString("test_fixtures/rows_scanner_interface.txt")))
		})
	})

	Describe("NewScannerFunc", func() {
		Describe("Build", func() {
			It("the new scanner function", func() {
				newScannerFunc := NewScannerFunc{
					InterfaceName:  "IScanner",
					ScannerName:    "Scanner",
					ErrScannerName: "ErrScanner",
				}.Build()

				Expect(printer.Fprint(buffer, fset, newScannerFunc)).ToNot(HaveOccurred())
				Expect(buffer.String()).To(Equal(ReadString("test_fixtures/new_scanner_function.txt")))
			})
		})
	})

	Describe("NewScannerFunc", func() {
		Describe("Build", func() {
			It("the new scanner function", func() {
				newRowScannerFunc := NewRowScannerFunc{
					InterfaceName:  "IRowScanner",
					ScannerName:    "RowScanner",
					ErrScannerName: "ErrRowScanner",
				}.Build()

				Expect(printer.Fprint(buffer, fset, newRowScannerFunc)).ToNot(HaveOccurred())
				Expect(buffer.String()).To(Equal(ReadString("test_fixtures/new_row_scanner_function.txt")))
			})
		})
	})

	Describe("Functions", func() {
		It("should generate the provided functions", func() {
			decl := Functions{
				Parameters: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{&ast.Ident{Name: "arg0"}},
						Type:  &ast.StarExpr{X: &ast.Ident{Name: "AType"}},
					},
				},
			}.Generate("Hello", &ast.BlockStmt{}, &ast.BlockStmt{}, &ast.BlockStmt{})

			Expect(printer.Fprint(buffer, fset, decl)).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(ReadString("test_fixtures/functions.txt")))
		})
	})
})
