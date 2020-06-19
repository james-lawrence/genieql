package genieql

import (
	"go/ast"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Driver", func() {
	Describe("driverRegistry", func() {
		Describe("RegisterDriver", func() {
			It("should err if the driver is already registered", func() {
				driver := testDriver{}
				reg := driverRegistry{}
				Expect(reg.RegisterDriver("testDriver", driver)).ToNot(HaveOccurred())
				Expect(reg.RegisterDriver("testDriver", driver)).To(MatchError(ErrDuplicateDriver))
			})

			It("should register a driver", func() {
				driver := testDriver{}
				reg := driverRegistry{}
				Expect(reg.RegisterDriver("testDriver", driver)).ToNot(HaveOccurred())
			})
		})

		Describe("LookupDriver", func() {
			It("should err if the driver is not registered", func() {
				reg := driverRegistry{}
				driver, err := reg.LookupDriver("testDriver")
				Expect(driver).To(BeNil())
				Expect(err).To(MatchError("requested driver is not registered: 'testDriver'"))
			})

			It("should return the driver if its been registered", func() {
				driverName := "testDriver"
				driver := testDriver{}
				reg := driverRegistry{}
				Expect(reg.RegisterDriver(driverName, driver)).ToNot(HaveOccurred())
				foundDriver, err := reg.LookupDriver(driverName)
				Expect(err).ToNot(HaveOccurred())
				Expect(foundDriver).To(Equal(driver))
			})
		})
	})
})

type testDriver struct{}

func (t testDriver) LookupType(s string) (td ColumnDefinition, b error) { return td, b }
func (t testDriver) LookupNullableType(ast.Expr) ast.Expr                     { return nil }
func (t testDriver) NullableType(typ, from ast.Expr) (ast.Expr, bool)         { return nil, false }
func (t testDriver) Exported() map[string]reflect.Value {
	return map[string]reflect.Value{}
}
