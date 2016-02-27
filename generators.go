package genieql

import (
	"go/token"
	"io"
)

type Generator interface {
	Generate() (io.Reader, error)
}

type CrudWriter interface {
	Write(fset *token.FileSet) error
}
