package drivers

import (
	"go/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pq", func() {
	Describe("pqNullableTypes", func() {
		examples := []struct {
			typ        string
			nullable   bool
			resultExpr string
		}{
			{"int", false, "int"},
			{"time.Time", false, "time.Time"},
			{"*int", false, "*int"},
			{"*time.Time", true, "myVar.Time"},
		}

		It("should properly determine if the type is nullable and return the proper expression", func() {
			for _, example := range examples {
				typ := mustParseExpr(example.typ)
				myVar := mustParseExpr("myVar")
				rhs, nullable := pqNullableTypes(typ, myVar)
				Expect(nullable).To(Equal(example.nullable), example.typ)
				Expect(types.ExprString(rhs)).To(Equal(example.resultExpr), example.typ)
			}
		})
	})

	Describe("pqLookupNullableType", func() {
		typeTable := []struct {
			input, expected string
		}{
			{"int", "int"},
			{"time.Time", "time.Time"},
			{"*int", "*int"},
			{"*time.Time", "pq.NullTime"},
		}
		It("should properly convert types to their Null Equivalents", func() {
			for _, test := range typeTable {
				result := pqLookupNullableType(mustParseExpr(test.input))
				Expect(types.ExprString(result)).To(Equal(test.expected), test.input)
			}
		})
	})
})
