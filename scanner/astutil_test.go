package scanner

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"go/printer"
	"go/token"
	"go/types"
)

var _ = Describe("Astutil", func() {
	Describe("assignmentStatement", func() {
		examples := []struct {
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
			for _, example := range examples {
				buffer := bytes.NewBuffer([]byte{})
				dst := mustParseExpr("myVar")
				from := mustParseExpr("c0")
				typ := mustParseExpr(example.input)
				ifstmt := assignmentStatement(dst, from, typ, DefaultNullableTypes)
				printer.Fprint(buffer, fset, ifstmt)
				Expect(buffer.String()).To(Equal(example.expected), example.input)
			}
		})
	})

	Describe("composeNullableTypes", func() {
		examples := []struct {
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
			{"*time.Time", false, "*time.Time"},
		}

		It("should return true if one of the provided functions returns true", func() {
			for _, example := range examples {
				typ := mustParseExpr(example.typ)
				myVar := mustParseExpr("myVar")

				rhs, nullable := composeNullableType(DefaultNullableTypes)(typ, myVar)

				Expect(nullable).To(Equal(example.nullable), example.typ)
				Expect(types.ExprString(rhs)).To(Equal(example.resultExpr), example.typ)
			}
		})
	})

	Describe("composeNullableTypes", func() {
		examples := []struct {
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
			for _, example := range examples {
				typ := mustParseExpr(example.input)

				rhs := composeLookupNullableType(DefaultLookupNullableType)(typ)

				Expect(types.ExprString(rhs)).To(Equal(example.expected), example.input)
			}
		})
	})
})
