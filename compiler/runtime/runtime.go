// Package stdlib provides wrappers of standard library packages to be imported natively in Yaegi.
package runtime

import "reflect"

// Symbols variable stores the map of stdlib symbols per package
var Symbols = map[string]map[string]reflect.Value{}

// Provide access to go standard library (http://golang.org/pkg/)
// build yaegi-extract in the yaegi repository
// go build -o yaegi-extract ./internal/cmd/extract
// go list std | grep -v internal | grep -v '\.' | grep -v unsafe | grep -v syscall

// go:generate yaegi extract github.com/jackc/pgtype github.com/davecgh/go-spew/spew
