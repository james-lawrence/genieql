package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"go/types"
)

var _ = Describe("Nullable", func() {
	Describe("DefaultNullableTypes", func() {
		nullableTypes := []struct {
			typ        string
			nullable   bool
			resultExpr string
		}{
			{"int", true, "int(myVar.Int64)"},
			{"int32", true, "int32(myVar.Int64)"},
			{"int64", true, "myVar.Int64"},
			{"float", true, "float(myVar.Float64)"},
			{"float32", true, "float32(myVar.Float64)"},
			{"float64", true, "myVar.Float64"},
			{"bool", true, "myVar.Bool"},
			{"string", true, "myVar.String"},
			{"time.Time", false, "(bad expr)"},
		}

		It("should properly determine if the type is nullable and return the proper expression", func() {
			for _, test := range nullableTypes {
				typ := mustParseExpr(test.typ)
				myVar := mustParseExpr("myVar")
				nullable, rhs := DefaultNullableTypes(myVar, typ)
				Expect(nullable).To(Equal(test.nullable), test.typ)
				Expect(types.ExprString(rhs)).To(Equal(test.resultExpr), test.typ)
			}
		})
	})
})
