package genieql

import (
	"database/sql"
	sqldriver "database/sql/driver"
	"fmt"
	"go/ast"
	"log"
	"reflect"
	"strings"
)

// ErrMissingDriver - returned when a driver has not been registered.
type missingDriver struct {
	driver string
}

func (t missingDriver) Error() string {
	return fmt.Sprintf("requested driver is not registered: '%s'", t.driver)
}

// ErrDuplicateDriver - returned when a ddriver gets registered twice.
var ErrDuplicateDriver = fmt.Errorf("driver has already been registered")

var drivers = driverRegistry{}

// NullableType interface for functions that resolve nullable types to their expression.
type NullableType func(typ, from ast.Expr) (ast.Expr, bool)

// LookupNullableType interface for functions that map type's to their nullable counter parts.
type LookupNullableType func(typ ast.Expr) ast.Expr

// LookupTypeDefinition converts a expression into a type definition.
type LookupTypeDefinition func(typ ast.Expr) (NullableTypeDefinition, bool)

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

// PrintRegisteredDrivers print drivers in the registry, debugging utility.
func PrintRegisteredDrivers() {
	for key := range map[string]Driver(drivers) {
		log.Println("Driver", key)
	}
}

// Driver - driver specific details.
type Driver interface {
	LookupType(s string) (NullableTypeDefinition, bool)
	LookupNullableType(ast.Expr) ast.Expr
	NullableType(typ, from ast.Expr) (ast.Expr, bool)
	Exported() (res map[string]reflect.Value)
}

type decoder interface {
	sql.Scanner
	sqldriver.Valuer
}

// NullableTypeDefinition defines a type supported by the driver.
type NullableTypeDefinition struct {
	Type         string // dialect type
	Native       string // golang type
	NullType     string
	NullField    string
	CastRequired bool
	Decoder      decoder
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
		return nil, missingDriver{driver: name}
	}

	return impl, nil
}

// NewDriver builds a new driver from the component parts
func NewDriver(nt NullableType, lnt LookupNullableType, supported ...NullableTypeDefinition) Driver {
	return driver{nt: nt, lnt: lnt, supported: supported}
}

type driver struct {
	nt        NullableType
	lnt       LookupNullableType
	supported []NullableTypeDefinition
}

func (t driver) LookupType(l string) (NullableTypeDefinition, bool) {
	for _, s := range t.supported {
		if s.Type == l {
			return s, true
		}
	}

	return NullableTypeDefinition{}, false
}

func (t driver) LookupNullableType(typ ast.Expr) ast.Expr         { return t.lnt(typ) }
func (t driver) NullableType(typ, from ast.Expr) (ast.Expr, bool) { return t.nt(typ, from) }
func (t driver) Exported() (res map[string]reflect.Value) {
	res = map[string]reflect.Value{}
	for _, typ := range t.supported {
		if typ.Decoder == nil {
			continue
		}

		switch idx := strings.IndexRune(typ.NullType, '.'); idx {
		case -1:
			res[typ.NullType] = reflect.ValueOf(typ.Decoder)
		default:
			res[typ.NullType[idx+1:]] = reflect.ValueOf(typ.Decoder)
		}
	}

	return res
}
