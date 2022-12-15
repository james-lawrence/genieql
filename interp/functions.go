package genieql

import (
	"go/ast"
	"go/printer"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/generators/functions"
	"github.com/pkg/errors"
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
		n          *ast.FuncDecl
		cf         *ast.Field
		qf         *ast.Field
		cmaps      []genieql.ColumnMap
		qinputs    []ast.Expr
		encodings  []ast.Stmt
		localspec  []ast.Spec
		transforms []ast.Stmt
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
	t.signature.Params.List = generators.NormalizeFieldNames(t.signature.Params.List[1:]...)

	scanner := functions.DetectScanner(t.ctx, t.signature)

	errHandler := functions.ScannerErrorHandling(scanner)
	encode := generators.ColumnMapEncoder(t.ctx)

	if cmaps, err = generators.ColumnMapFromFields(t.ctx, t.signature.Params.List...); err != nil {
		return errors.Wrap(err, "unable to generate mapping")
	}

	for idx, cmap := range cmaps {
		var (
			tmp []ast.Stmt
		)

		local := cmap.Local(idx)

		if tmp, err = encode(idx, cmap, errHandler); err != nil {
			return errors.Wrap(err, "failed to generate encode")
		}

		if tmp == nil {
			qinputs = append(qinputs, ast.NewIdent(cmap.Name))
			continue
		}

		qinputs = append(qinputs, local)
		encodings = append(encodings, tmp...)

		vspec := astutil.ValueSpec(astutil.MustParseExpr(t.ctx.FileSet, cmap.Definition.ColumnType), local)
		vspec.Comment = &ast.CommentGroup{
			List: []*ast.Comment{
				{
					Text: "// " + cmap.ColumnInfo.Name,
				},
			},
		}

		localspec = append(localspec, vspec)
	}

	transforms = []ast.Stmt{
		&ast.DeclStmt{
			Decl: astutil.VarList(localspec...),
		},
	}

	transforms = append(transforms, encodings...)

	qfn := functions.Query{
		Context: t.ctx,
		Query: astutil.StringLiteral(
			functions.QueryLiteralColumnMapReplacer(t.ctx, cmaps...).Replace(t.query),
		),
		Scanner:      scanner,
		Queryer:      qf.Type,
		ContextField: cf,
		Transforms:   transforms,
		QueryInputs:  qinputs,
	}

	if n, err = qfn.Compile(functions.New(t.name, t.signature)); err != nil {
		return err
	}

	if err = generators.GenerateComment(t.comment, generators.DefaultFunctionComment(t.name)).Generate(dst); err != nil {
		return err
	}

	if err = printer.Fprint(dst, t.ctx.FileSet, n); err != nil {
		return err
	}

	return nil
}
