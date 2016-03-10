package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"go/printer"
	"go/token"
)

var _ = Describe("Astutil", func() {

	Describe("assignmentStatement", func() {
		typeTable := []struct {
			input, expected string
		}{
			{"int", "if c0.Valid {\n\tmyVar = int(c0.Int64)\n}"},
			{"int32", "if c0.Valid {\n\tmyVar = int32(c0.Int64)\n}"},
			{"int64", "if c0.Valid {\n\tmyVar = c0.Int64\n}"},
			{"float", "if c0.Valid {\n\tmyVar = float(c0.Float64)\n}"},
			{"float32", "if c0.Valid {\n\tmyVar = float32(c0.Float64)\n}"},
			{"float64", "if c0.Valid {\n\tmyVar = c0.Float64\n}"},
			{"bool", "if c0.Valid {\n\tmyVar = c0.Bool\n}"},
			{"string", "if c0.Valid {\n\tmyVar = c0.String\n}"},
			{"time.Time", "myVar = c0"},
		}

		It("should do something", func() {
			fset := token.NewFileSet()
			for _, test := range typeTable {
				buffer := bytes.NewBuffer([]byte{})
				dst := mustParseExpr("myVar")
				from := mustParseExpr("c0")
				typ := mustParseExpr(test.input)
				ifstmt := assignmentStatement(dst, from, typ, DefaultNullableTypes)
				printer.Fprint(buffer, fset, ifstmt)
				Expect(buffer.String()).To(Equal(test.expected), test.input)
			}
		})
	})
})
