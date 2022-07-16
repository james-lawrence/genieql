package genieql_test

import (
	"bytes"
	"go/printer"
	"go/token"

	. "bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("QueryLiteral", func() {
	It("return a literal const string with the provided name and value", func() {
		fset := token.NewFileSet()
		dst := bytes.NewBuffer([]byte{})
		decl := QueryLiteral("MyLiteral", "Value")

		Expect(printer.Fprint(dst, fset, decl)).ToNot(HaveOccurred())
		Expect(dst.String()).To(Equal("const MyLiteral = `Value`"))
	})
})
