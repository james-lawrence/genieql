package genieql

import (
	"go/token"
	"io"
)

// Generator interface for the code generators.
type Generator interface {
	Generate(dst io.Writer, fset *token.FileSet) error
}

// CrudWriter TODO...
type CrudWriter interface {
	Write(fset *token.FileSet) error
}

// MultiGenerate generate multiple scanners into a single buffer.
func MultiGenerate(generators ...Generator) Generator {
	return multiGenerator{
		generators: generators,
	}
}

type multiGenerator struct {
	generators []Generator
}

func (t multiGenerator) Generate(dst io.Writer, fset *token.FileSet) error {
	for _, generator := range t.generators {
		if err := generator.Generate(dst, fset); err != nil {
			return err
		}
	}
	return nil
}
