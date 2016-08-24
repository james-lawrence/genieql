package genieql

import (
	"fmt"
	"io"
)

// Generator interface for the code generators.
type Generator interface {
	Generate(dst io.Writer) error
}

// CrudWriter TODO...
type CrudWriter interface {
	Write() error
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

func (t multiGenerator) Generate(dst io.Writer) error {
	for _, generator := range t.generators {
		if err := generator.Generate(dst); err != nil {
			return err
		}
		fmt.Fprintf(dst, "\n\n")
	}
	return nil
}
