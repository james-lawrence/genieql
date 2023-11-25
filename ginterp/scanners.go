package ginterp

import (
	"go/ast"
	"io"
	"strings"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/transformx"
)

// Scanner - configuration interface for generating scanners.
type Scanner interface {
	genieql.Generator // must satisfy the generator interface
	ColumnNamePrefix(string) Scanner
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
	name             string
	ctx              generators.Context
	params           *ast.FieldList
	columnNamePrefix string
}

func (t *scanner) ColumnNamePrefix(s string) Scanner {
	t.columnNamePrefix = s
	return t
}

func (t *scanner) Generate(dst io.Writer) error {
	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")

	modes := generators.ScannerOptionNoop
	if len(t.params.List) > 1 && !generators.AllBuiltinTypes(astutil.MapFieldsToTypeExpr(t.params.List...)...) {
		t.ctx.Println("multiple structures detected disabling dynamic scanner output for", t.name)
		modes = generators.ScannerOptionOutputMode(generators.ModeInterface | generators.ModeStatic | generators.ModeStaticDisableColumns)
	}

	columnNamePrefix := generators.ScannerOptionNoop
	if s := strings.TrimSpace(t.columnNamePrefix); s != "" {
		columnNamePrefix = generators.ScannerOptionColumnNameTransformer(transformx.Prefix(s))
	}

	return generators.NewScanner(
		generators.ScannerOptionContext(t.ctx),
		generators.ScannerOptionName(t.name),
		generators.ScannerOptionParameters(t.params),
		columnNamePrefix,
		modes,
	).Generate(dst)
}
