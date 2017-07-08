package drivers_test

import (
	"bitbucket.org/jatone/genieql"
	. "bitbucket.org/jatone/genieql/internal/drivers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
)

var _ = Describe("drivers", func() {
	DescribeTable("NullableTypes",
		nullableTypeTest(DefaultNullableTypes),
		Entry("handle int", "int", true, "int(localVariable.Int64)"),
		Entry("handle int32", "int32", true, "int32(localVariable.Int64)"),
		Entry("handle int64 ", "int64", true, "localVariable.Int64"),
		Entry("handle float", "float", true, "float(localVariable.Float64)"),
		Entry("handle float32", "float32", true, "float32(localVariable.Float64)"),
		Entry("handle float64", "float64", true, "localVariable.Float64"),
		Entry("handle bool", "bool", true, "localVariable.Bool"),
		Entry("handle string", "string", true, "localVariable.String"),
		Entry("not handle time.Time", "time.Time", false, "time.Time"),
		Entry("handle *int", "*int", true, "int(localVariable.Int64)"),
		Entry("handle *int32", "*int32", true, "int32(localVariable.Int64)"),
		Entry("handle *int64", "*int64", true, "localVariable.Int64"),
		Entry("handle *float", "*float", true, "float(localVariable.Float64)"),
		Entry("handle *float32", "*float32", true, "float32(localVariable.Float64)"),
		Entry("handle *float64", "*float64", true, "localVariable.Float64"),
		Entry("handle *bool", "*bool", true, "localVariable.Bool"),
		Entry("handle *string", "*string", true, "localVariable.String"),
		Entry("not handle *time.Time", "*time.Time", false, "*time.Time"),
	)

	DescribeTable("LookupNullableType",
		lookupNullableTypeTest(DefaultLookupNullableType),
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

	var (
		exampleDriver = NewDriver(
			NullableType{Type: "string", NullType: "sql.NullString", NullField: "String"},
			NullableType{Type: "int", NullType: "sql.NullInt64", NullField: "Int64"},
		)
	)
	DescribeTable("driver lookup nullable",
		func(_type, nulltype string, driver genieql.Driver) {
			lookupNullableTypeTest(driver.LookupNullableType)(_type, nulltype)
		},
		Entry("basic test", "int", "sql.NullInt64", exampleDriver),
		Entry("basic test", "*int", "sql.NullInt64", exampleDriver),
	)
	DescribeTable("driver nullableTypes",
		func(_type string, nullable bool, nulltype string, driver genieql.Driver) {
			nullableTypeTest(driver.NullableType)(_type, nullable, nulltype)
		},
		Entry("handle int", "int", true, "localVariable.Int64", exampleDriver),
		Entry("handle *int", "*int", true, "localVariable.Int64", exampleDriver),
		Entry("handle string", "string", true, "localVariable.String", exampleDriver),
		Entry("handle *string", "*string", true, "localVariable.String", exampleDriver),
		Entry("not handle bool", "bool", false, "bool", exampleDriver),
		Entry("not handle *bool", "*bool", false, "*bool", exampleDriver),
	)
})
