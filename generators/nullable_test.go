package generators_test

import (
	"bitbucket.org/jatone/genieql/astutil"
	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"go/types"
)

var _ = ginkgo.Describe("Nullable", func() {
	DescribeTable("DefaultNullableTypes",
		func(input string, nullable bool, castExpr string) {
			typ := astutil.Expr(input)
			myVar := astutil.Expr("myVar")
			rhs, nullable := DefaultNullableTypes(typ, myVar)
			Expect(nullable).To(Equal(nullable), input)
			Expect(types.ExprString(rhs)).To(Equal(castExpr), input)
		},
		Entry("handle int", "int", true, "int(myVar.Int64)"),
		Entry("handle int32", "int32", true, "int32(myVar.Int64)"),
		Entry("handle int64 ", "int64", true, "myVar.Int64"),
		Entry("handle float", "float", true, "float(myVar.Float64)"),
		Entry("handle float32", "float32", true, "float32(myVar.Float64)"),
		Entry("handle float64", "float64", true, "myVar.Float64"),
		Entry("handle bool", "bool", true, "myVar.Bool"),
		Entry("handle string", "string", true, "myVar.String"),
		Entry("not handle time.Time", "time.Time", false, "time.Time"),
		Entry("handle *int", "*int", true, "int(myVar.Int64)"),
		Entry("handle *int32", "*int32", true, "int32(myVar.Int64)"),
		Entry("handle *int64", "*int64", true, "myVar.Int64"),
		Entry("handle *float", "*float", true, "float(myVar.Float64)"),
		Entry("handle *float32", "*float32", true, "float32(myVar.Float64)"),
		Entry("handle *float64", "*float64", true, "myVar.Float64"),
		Entry("handle *bool", "*bool", true, "myVar.Bool"),
		Entry("handle *string", "*string", true, "myVar.String"),
		Entry("not handle *time.Time", "*time.Time", false, "*time.Time"),
	)

	DescribeTable("DefaultLookupNullableType",
		func(input, output string) {
			result := DefaultLookupNullableType(astutil.Expr(input))
			Expect(types.ExprString(result)).To(Equal(output))
		},
		Entry("int", "int", "sql.NullInt64"),
		Entry("int32", "int32", "sql.NullInt64"),
		Entry("int64", "int64", "sql.NullInt64"),
		Entry("float", "float", "sql.NullFloat64"),
		Entry("float32", "float32", "sql.NullFloat64"),
		Entry("float64", "float64", "sql.NullFloat64"),
		Entry("bool", "bool", "sql.NullBool"),
		Entry("string", "string", "sql.NullString"),
		Entry("time.Time", "time.Time", "time.Time"),
		Entry("*int", "*int", "sql.NullInt64"),
		Entry("*int32", "*int32", "sql.NullInt64"),
		Entry("*int64", "*int64", "sql.NullInt64"),
		Entry("*float", "*float", "sql.NullFloat64"),
		Entry("*float32", "*float32", "sql.NullFloat64"),
		Entry("*float64", "*float64", "sql.NullFloat64"),
		Entry("*bool", "*bool", "sql.NullBool"),
		Entry("*string", "*string", "sql.NullString"),
		Entry("*time.Time", "*time.Time", "time.Time"),
	)
})
