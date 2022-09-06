package genieql

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/generators/functions"
	"bitbucket.org/jatone/genieql/generators/typespec"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"
)

// InsertBatch configuration interface for generating batch inserts.
type InsertBatch interface {
	genieql.Generator              // must satisfy the generator interface
	Into(string) InsertBatch       // what table to insert into
	Ignore(...string) InsertBatch  // do not attempt to insert the specified column.
	Default(...string) InsertBatch // use the database default for the specified columns.
	Conflict(string) InsertBatch   // specify how conflicts should be handled.
	Batch(n int) InsertBatch       // specify a batch insert
}

// NewInsert instantiate a new insert generator. it uses the name of function
// that calls Define as the name of the generated function.
func NewBatchInsert(
	ctx generators.Context,
	name string,
	comment *ast.CommentGroup,
	cf *ast.Field,
	qf *ast.Field,
	tf *ast.Field,
	scanner *ast.FuncDecl,
) InsertBatch {
	return &batch{
		ctx:     ctx,
		name:    name,
		comment: comment,
		qf:      qf,
		cf:      cf,
		tf:      tf,
		scanner: scanner,
		n:       1,
	}
}

type batch struct {
	ctx      generators.Context
	n        int // number of records to support inserting
	name     string
	table    string
	conflict string
	defaults []string
	ignore   []string
	tf       *ast.Field    // type field.
	cf       *ast.Field    // context field, can be nil.
	qf       *ast.Field    // db Query field.
	scanner  *ast.FuncDecl // scanner being used for results.
	comment  *ast.CommentGroup
}

// Into specify the table the data will be inserted into.
func (t *batch) Into(s string) InsertBatch {
	t.table = s
	return t
}

// Default specify the table columns to be given their default values.
func (t *batch) Default(defaults ...string) InsertBatch {
	t.defaults = defaults
	return t
}

// Ignore specify the table columns to ignore during insert.
func (t *batch) Ignore(ignore ...string) InsertBatch {
	t.ignore = ignore
	return t
}

// Conflict specify how to handle conflict during an insert.
func (t *batch) Conflict(s string) InsertBatch {
	t.conflict = s
	return t
}

// Batch specify the maximum number of records to insert.
func (t *batch) Batch(size int) InsertBatch {
	t.n = size
	return t
}

func (t *batch) Generate(dst io.Writer) (err error) {
	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	t.ctx.Debugln("batch.insert type", t.ctx.CurrentPackage.Name, t.ctx.CurrentPackage.ImportPath, types.ExprString(t.tf.Type))
	t.ctx.Debugln("batch.insert table", t.table)

	initializesig := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				t.cf,
				t.qf,
				astutil.Field(&ast.Ellipsis{
					Elt: t.tf.Type,
				}, t.tf.Names...),
			},
		},
		Results: t.scanner.Type.Results,
	}

	typename := stringsx.ToPrivate(t.name)
	initialize := functions.NewFn(
		astutil.Return(
			&ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: ast.NewIdent(typename),
					Elts: []ast.Expr{
						&ast.KeyValueExpr{
							Key:   t.cf.Names[0],
							Value: t.cf.Names[0],
						},
						&ast.KeyValueExpr{
							Key:   t.qf.Names[0],
							Value: t.qf.Names[0],
						},
						&ast.KeyValueExpr{
							Key:   ast.NewIdent("remaining"),
							Value: t.tf.Names[0],
						},
					},
				},
			},
		),
	)

	typedecl := typespec.NewType(typename, &ast.StructType{
		Struct: token.Pos(0),
		Fields: &ast.FieldList{
			List: []*ast.Field{
				t.cf,
				t.qf,
				astutil.Field(
					&ast.ArrayType{Elt: t.tf.Type}, ast.NewIdent("remaining")),
			},
		},
	})

	fnrecv := astutil.FieldList(astutil.Field(&ast.StarExpr{X: astutil.Expr(typename)}, ast.NewIdent("t")))

	scansig := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(&ast.StarExpr{X: t.tf.Type}, t.tf.Names...),
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(ast.NewIdent("error")),
			},
		},
	}
	scanfn := functions.NewFn(
		astutil.Return(
			astutil.CallExpr(
				&ast.SelectorExpr{
					X: astutil.SelExpr(
						"t", "scanner",
					),
					Sel: ast.NewIdent("Scan"),
				},
				t.tf.Names[0],
			),
		),
	)

	errsig := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(ast.NewIdent("error")),
			},
		},
	}

	errfn := functions.NewFn(
		astutil.If(
			nil,
			astutil.BinaryExpr(astutil.SelExpr("t", "scanner"), token.EQL, ast.NewIdent("nil")),
			astutil.Block(
				astutil.Return(ast.NewIdent("nil")),
			),
			nil,
		),
		astutil.Return(
			astutil.CallExpr(
				astutil.SelExpr(
					types.ExprString(
						astutil.SelExpr(
							"t", "scanner",
						),
					),
					"Err",
				),
			),
		),
	)

	closesig := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(ast.NewIdent("error")),
			},
		},
	}

	closefn := functions.NewFn(
		astutil.If(
			nil,
			astutil.BinaryExpr(astutil.SelExpr("t", "scanner"), token.EQL, ast.NewIdent("nil")),
			astutil.Block(
				astutil.Return(ast.NewIdent("nil")),
			),
			nil,
		),
		astutil.Return(
			astutil.CallExpr(
				astutil.SelExpr(
					types.ExprString(
						astutil.SelExpr(
							"t", "scanner",
						),
					),
					"Close",
				),
			),
		),
	)

	nextsig := &ast.FuncType{
		Params: &ast.FieldList{},
		Results: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(ast.NewIdent("bool")),
			},
		},
	}

	nextfn := functions.NewFn(
		astutil.Return(
			ast.NewIdent("false"),
		),
	)

	advancesig := &ast.FuncType{
		Params:  &ast.FieldList{},
		Results: &ast.FieldList{},
	}
	advancefn := functions.NewFn(
		astutil.Return(
			ast.NewIdent("false"),
		),
	)

	return genieql.NewFuncGenerator(func(dst io.Writer) (err error) {
		if err = generators.GenerateComment(generators.DefaultFunctionComment(t.name), t.comment).Generate(dst); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("New"+stringsx.ToPublic(t.name), initializesig), initialize); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = typespec.CompileInto(dst, typedecl); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("Scan", scansig, functions.OptionRecv(fnrecv)), scanfn); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("Err", errsig, functions.OptionRecv(fnrecv)), errfn); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("Close", closesig, functions.OptionRecv(fnrecv)), closefn); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("Next", nextsig, functions.OptionRecv(fnrecv)), nextfn); err != nil {
			return err
		}

		if err = generators.GapLines(dst, 2); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New("advance", advancesig, functions.OptionRecv(fnrecv)), advancefn); err != nil {
			return err
		}
		return nil
	}).Generate(dst)
}
