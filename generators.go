package genieql

import (
	"go/token"
	"io"
)

// Generator TODO...
type Generator interface {
	Generate(dst io.Writer, fset *token.FileSet) error
}

// CrudWriter TODO...
type CrudWriter interface {
	Write(fset *token.FileSet) error
}
