package genieql

import (
	"go/ast"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
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
	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")

	modes := generators.ScannerOptionNoop
	if len(t.params.List) > 1 && !generators.AllBuiltinTypes(astutil.MapFieldsToTypExpr(t.params.List...)...) {
		t.ctx.Println("multiple structures detected disabling dynamic scanner output for", t.name)
		modes = generators.ScannerOptionOutputMode(generators.ModeInterface | generators.ModeStatic | generators.ModeStaticDisableColumns)
	}

	return generators.NewScanner(
		generators.ScannerOptionContext(t.ctx),
		generators.ScannerOptionName(t.name),
		generators.ScannerOptionParameters(t.params),
		modes,
	).Generate(dst)
}
