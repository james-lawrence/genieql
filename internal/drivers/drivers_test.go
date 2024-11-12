package drivers_test

import (
	"errors"
	"reflect"

	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/internal/drivers"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("drivers", func() {
	var (
		exampleDriver = NewDriver(
			"",
			map[string]reflect.Value{},
			genieql.ColumnDefinition{Type: "string", ColumnType: "sql.NullString", Native: "string"},
			genieql.ColumnDefinition{Type: "*string", ColumnType: "sql.NullString", Native: "*string"},
			genieql.ColumnDefinition{Type: "int", ColumnType: "sql.NullInt64", Native: "int"},
			genieql.ColumnDefinition{Type: "*int", ColumnType: "sql.NullInt64", Native: "*int"},
		)
	)

	DescribeTable("DefaultLookupNullableType",
		lookupDefinitionTest(DefaultTypeDefinitions),
		Entry("int", "int", "sql.NullInt64", nil),
		Entry("int32", "int32", "sql.NullInt32", nil),
		Entry("int64", "int64", "sql.NullInt64", nil),
		Entry("float32", "float32", "sql.NullFloat64", nil),
		Entry("float64", "float64", "sql.NullFloat64", nil),
		Entry("bool", "bool", "sql.NullBool", nil),
		Entry("string", "string", "sql.NullString", nil),
		Entry("time.Time", "time.Time", "sql.NullTime", nil),
		Entry("*int", "*int", "sql.NullInt64", nil),
		Entry("*int32", "*int32", "sql.NullInt32", nil),
		Entry("*int64", "*int64", "sql.NullInt64", nil),
		Entry("*float32", "*float32", "sql.NullFloat64", nil),
		Entry("*float64", "*float64", "sql.NullFloat64", nil),
		Entry("*bool", "*bool", "sql.NullBool", nil),
		Entry("*string", "*string", "sql.NullString", nil),
		Entry("*time.Time", "*time.Time", "sql.NullTime", nil),
	)

	DescribeTable("driver column definitions",
		lookupDefinitionTest(exampleDriver.LookupType),
		Entry("handle int", "int", "sql.NullInt64", nil),
		Entry("handle *int", "*int", "sql.NullInt64", nil),
		Entry("handle string", "string", "sql.NullString", nil),
		Entry("handle *string", "*string", "sql.NullString", nil),
		Entry("not handle bool", "bool", "", errors.New("not found")),
		Entry("not handle *bool", "*bool", "", errors.New("not found")),
	)
})
