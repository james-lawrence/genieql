package genieql

import (
	"errors"
	"go/ast"
	"go/printer"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/generators/functions"
)

// Function configuration interface for generating functions.
type Function interface {
	genieql.Generator // must satisfy the generator interface
	Query(string) Function
}

// NewFunction instantiate a new function generator. it uses the name of function
// that calls Define as the name of the generated function.
func NewFunction(
	ctx generators.Context,
	name string,
	signature *ast.FuncType,
	comment *ast.CommentGroup,
) Function {
	return &function{
		ctx:       ctx,
		name:      name,
		signature: signature,
		comment:   comment,
	}
}

type function struct {
	ctx       generators.Context
	name      string
	signature *ast.FuncType
	comment   *ast.CommentGroup
	query     string
}

func (t *function) Query(q string) Function {
	t.query = q
	return t
}

func (t *function) Generate(dst io.Writer) (err error) {
	var (
		n  *ast.FuncDecl
		cf *ast.Field
		qf *ast.Field
	)

	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")

	if cf = functions.DetectContext(t.signature); cf != nil {
		// pop the context off the params.
		t.signature.Params.List = t.signature.Params.List[1:]
	}

	if len(t.signature.Params.List) < 1 {
		return errors.New("functions must start with a queryer param")
	}

	// pop the queryer off the params.
	qf = t.signature.Params.List[0]
	t.signature.Params.List = t.signature.Params.List[1:]

	qfn := functions.Query{
		Context:      t.ctx,
		Query:        astutil.StringLiteral(t.query),
		Scanner:      functions.DetectScanner(t.ctx, t.signature),
		Queryer:      qf.Type,
		ContextField: cf,
	}

	if n, err = qfn.Compile(functions.New(t.name, t.signature)); err != nil {
		return err
	}

	if err = printer.Fprint(dst, t.ctx.FileSet, n); err != nil {
		return err
	}

	return nil
}
