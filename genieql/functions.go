package genieql

import (
	"go/ast"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
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

func (t *function) Generate(dst io.Writer) error {
	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	return generators.NewQueryFunctionFromFuncType(
		t.ctx,
		t.signature,
		generators.QFOName(t.name),
		generators.QFOBuiltinQueryFromString(t.query),
		generators.QFOComment(t.comment),
	).Generate(dst)
}
