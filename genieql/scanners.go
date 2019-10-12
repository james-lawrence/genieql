package genieql

import (
	"go/ast"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
)

// Scanner - configuration interface for generating scanners.
type Scanner interface {
	genieql.Generator // must satisfy the generator interface
}

// NewScanner instantiate a new scanner generator. it uses the name of function
// that calls Define as the name of the emitted type.
func NewScanner(
	ctx generators.Context,
	name string,
	params *ast.FieldList,
) Scanner {
	return &scanner{ctx: ctx, name: name, params: params}
}

type scanner struct {
	name   string
	ctx    generators.Context
	params *ast.FieldList
}

func (t *scanner) Generate(dst io.Writer) error {
	log.Println("generation of", t.name, "initiated")
	err := generators.NewScanner(
		generators.ScannerOptionContext(t.ctx),
		generators.ScannerOptionName(t.name),
		generators.ScannerOptionParameters(t.params),
	).Generate(dst)
	if err == nil {
		log.Println("generation of", t.name, "completed")
	}
	return err
}
