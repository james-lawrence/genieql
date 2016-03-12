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
			{"int", "myVar = c0"},
			{"int32", "myVar = c0"},
			{"int64", "myVar = c0"},
			{"float", "myVar = c0"},
			{"float32", "myVar = c0"},
			{"float64", "myVar = c0"},
			{"bool", "myVar = c0"},
			{"string", "myVar = c0"},
			{"time.Time", "myVar = c0"},
			{"*float64", "if c0.Valid {\n\ttmp := c0.Float64\n\tmyVar = &tmp\n}"},
			{"*bool", "if c0.Valid {\n\ttmp := c0.Bool\n\tmyVar = &tmp\n}"},
			{"*string", "if c0.Valid {\n\ttmp := c0.String\n\tmyVar = &tmp\n}"},
		}

		It("correctly generate the assignment statements", func() {
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
