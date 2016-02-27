package genieql

import (
	"go/token"
	"io"
)

type Generator interface {
	Generate(dst io.Writer, fset *token.FileSet) error
}

type CrudWriter interface {
	Write(fset *token.FileSet) error
}
