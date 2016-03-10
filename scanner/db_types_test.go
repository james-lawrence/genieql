package scanner

import (
	// . "bitbucket.org/jatone/genieql/scanner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// "bytes"
	// "go/ast"
	// "go/printer"
	// "go/token"
	"go/types"
)

var _ = Describe("DbTypes", func() {

})

var _ = Describe("mapColumnType", func() {
	typeTable := []struct {
		input, expected string
	}{
		{"int", "sql.NullInt64"},
		{"int32", "sql.NullInt64"},
		{"int64", "sql.NullInt64"},
		{"float", "sql.NullFloat64"},
		{"float32", "sql.NullFloat64"},
		{"float64", "sql.NullFloat64"},
		{"bool", "sql.NullBool"},
		{"string", "sql.NullString"},
	}
	It("should properly convert types to their Null Equivalents", func() {
		for _, test := range typeTable {
			result := LookupNullableType(mustParseExpr(test.input))
			Expect(types.ExprString(result)).To(Equal(test.expected), test.input)
		}
	})
})

var _ = Describe("assignStatement", func() {
	// typeTable := []struct {
	// 	input, expected string
	// }{
	// 	{"int", "if c0.Valid {\n\tmyVar = int(c0.Int64)\n}"},
	// 	{"int32", "if c0.Valid {\n\tmyVar = int32(c0.Int64)\n}"},
	// 	{"int64", "if c0.Valid {\n\tmyVar = int64(c0.Int64)\n}"},
	// 	{"float", "if c0.Valid {\n\tmyVar = float(c0.Float64)\n}"},
	// 	{"float32", "if c0.Valid {\n\tmyVar = float32(c0.Float64)\n}"},
	// 	{"float64", "if c0.Valid {\n\tmyVar = float64(c0.Float64)\n}"},
	// 	{"bool", "if c0.Valid {\n\tmyVar = bool(c0.Bool)\n}"},
	// 	{"string", "if c0.Valid {\n\tmyVar = string(c0.String)\n}"},
	// }
	//
	// It("should do something", func() {
	// 	fset := token.NewFileSet()
	// 	for _, test := range typeTable {
	// 		buffer := bytes.NewBuffer([]byte{})
	// 		dst := mustParseExpr("myVar").(*ast.Ident)
	// 		from := mustParseExpr("c0").(*ast.Ident)
	// 		typ := mustParseExpr(test.input).(*ast.Ident)
	// 		ifstmt := assignStatement(dst, from, typ)
	// 		printer.Fprint(buffer, fset, ifstmt)
	// 		Expect(buffer.String()).To(Equal(test.expected), test.input)
	// 	}
	// })
})
