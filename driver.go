package genieql

import (
	"fmt"
	"go/ast"
)

// ErrMissingDriver - returned when a driver has not been registered.
var ErrMissingDriver = fmt.Errorf("requested driver is not registered")

// ErrDuplicateDriver - returned when a ddriver gets registered twice.
var ErrDuplicateDriver = fmt.Errorf("driver has already been registered")

var drivers = driverRegistry{}

// NullableType interface for functions that resolve nullable types to their expression.
type NullableType func(typ, from ast.Expr) (ast.Expr, bool)

// LookupNullableType interface for functions that map type's to their nullable counter parts.
type LookupNullableType func(typ ast.Expr) ast.Expr

// RegisterDriver register a database driver with genieql. usually in an init function.
func RegisterDriver(driver string, imp Driver) error {
	return drivers.RegisterDriver(driver, imp)
}

// LookupDriver lookup a registered driver.
func LookupDriver(name string) (Driver, error) {
	return drivers.LookupDriver(name)
}

// MustLookupDriver panics if the driver cannot be found, convience method.
func MustLookupDriver(name string) Driver {
	driver, err := LookupDriver(name)
	if err != nil {
		panic(err)
	}
	return driver
}

// Driver - driver specific details.
type Driver interface {
	LookupNullableType(ast.Expr) ast.Expr
	NullableType(typ, from ast.Expr) (ast.Expr, bool)
}

type driverRegistry map[string]Driver

func (t driverRegistry) RegisterDriver(driver string, imp Driver) error {
	if _, exists := t[driver]; exists {
		return ErrDuplicateDriver
	}

	t[driver] = imp

	return nil
}

func (t driverRegistry) LookupDriver(name string) (Driver, error) {
	impl, exists := t[name]
	if !exists {
		return nil, ErrMissingDriver
	}

	return impl, nil
}

// NewDriver builds a new driver from the component parts
func NewDriver(nt NullableType, lnt LookupNullableType) Driver {
	return driver{nt: nt, lnt: lnt}
}

type driver struct {
	nt  NullableType
	lnt LookupNullableType
}

func (t driver) LookupNullableType(typ ast.Expr) ast.Expr         { return t.lnt(typ) }
func (t driver) NullableType(typ, from ast.Expr) (ast.Expr, bool) { return t.nt(typ, from) }
