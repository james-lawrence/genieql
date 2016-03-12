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
			{"int", false, "int"},
			{"int32", false, "int32"},
			{"int64", false, "int64"},
			{"float", false, "float"},
			{"float32", false, "float32"},
			{"float64", false, "float64"},
			{"bool", false, "bool"},
			{"string", false, "string"},
			{"time.Time", false, "time.Time"},
			{"*int", true, "int(myVar.Int64)"},
			{"*int32", true, "int32(myVar.Int64)"},
			{"*int64", true, "myVar.Int64"},
			{"*float", true, "float(myVar.Float64)"},
			{"*float32", true, "float32(myVar.Float64)"},
			{"*float64", true, "myVar.Float64"},
			{"*bool", true, "myVar.Bool"},
			{"*string", true, "myVar.String"},
			{"*time.Time", false, "(bad expr)"},
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

	Describe("DefaultLookupNullableType", func() {
		typeTable := []struct {
			input, expected string
		}{
			{"int", "int"},
			{"int32", "int32"},
			{"int64", "int64"},
			{"float", "float"},
			{"float32", "float32"},
			{"float64", "float64"},
			{"bool", "bool"},
			{"string", "string"},
			{"time.Time", "time.Time"},
			{"*int", "sql.NullInt64"},
			{"*int32", "sql.NullInt64"},
			{"*int64", "sql.NullInt64"},
			{"*float", "sql.NullFloat64"},
			{"*float32", "sql.NullFloat64"},
			{"*float64", "sql.NullFloat64"},
			{"*bool", "sql.NullBool"},
			{"*string", "sql.NullString"},
			{"*time.Time", "*time.Time"},
		}
		It("should properly convert types to their Null Equivalents", func() {
			for _, test := range typeTable {
				result := DefaultLookupNullableType(mustParseExpr(test.input))
				Expect(types.ExprString(result)).To(Equal(test.expected), test.input)
			}
		})
	})
})
