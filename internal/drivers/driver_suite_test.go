package drivers_test

import (
	"go/types"

	"bitbucket.org/jatone/genieql"

	. "bitbucket.org/jatone/genieql/internal/drivers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Driver Suite")
}

func nullableTypeTest(nullableType genieql.NullableType) func(typs string, nullable bool, exprs string) {
	return func(typs string, nullable bool, exprs string) {
		typ := MustParseExpr(typs)
		driverNullableType := MustParseExpr("nullableType")
		evaluated, nullable := nullableType(typ, driverNullableType)
		Expect(nullable).To(Equal(nullable))
		Expect(types.ExprString(evaluated)).To(Equal(exprs))
	}
}

func lookupNullableTypeTest(lookupNullableType genieql.LookupNullableType) func(typs, exprs string) {
	return func(typs, exprs string) {
		result := lookupNullableType(MustParseExpr(typs))
		Expect(types.ExprString(result)).To(Equal(exprs))
	}
}
