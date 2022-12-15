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
	"bitbucket.org/jatone/genieql/internal/stringsx"
)

// InsertBatch configuration interface for generating batch inserts.
type InsertBatch interface {
	genieql.Generator              // must satisfy the generator interface
	Into(string) InsertBatch       // what table to insert into
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
		astutil.DeclStmt(
			astutil.VarList(
				astutil.ValueSpec(ast.NewIdent("bool"), ast.NewIdent("advanced")),
			),
		),
		astutil.If(
			nil, astutil.BinaryExpr(
				astutil.BinaryExpr(astutil.SelExpr("t", "scanner"), token.NEQ, ast.NewIdent("nil")),
				token.LAND,
				astutil.CallExpr(
					&ast.SelectorExpr{
						X:   astutil.SelExpr("t", "scanner"),
						Sel: ast.NewIdent("Next"),
					},
				),
			),
			astutil.Block(
				astutil.Return(ast.NewIdent("true")),
			),
			nil,
		),
		astutil.If(
			nil, astutil.BinaryExpr(
				astutil.BinaryExpr(astutil.CallExpr(ast.NewIdent("len"), astutil.SelExpr("t", "remaining")), token.GTR, astutil.IntegerLiteral(0)),
				token.LAND,
				astutil.BinaryExpr(
					astutil.CallExpr(
						astutil.SelExpr("t", "Close"),
					),
					token.EQL,
					ast.NewIdent("nil"),
				),
			),
			astutil.Block(
				astutil.Assign(
					astutil.ExprList(
						astutil.SelExpr("t", "scanner"),
						astutil.SelExpr("t", "remaining"),
						ast.NewIdent("advanced"),
					),
					token.ASSIGN,
					astutil.ExprList(
						astutil.CallExprEllipsis(
							astutil.SelExpr("t", "advance"),
							astutil.SelExpr("t", "remaining"),
						),
					),
				),
				astutil.Return(
					astutil.BinaryExpr(ast.NewIdent("advanced"), token.LAND, astutil.CallExpr(&ast.SelectorExpr{
						X:   astutil.SelExpr("t", "scanner"),
						Sel: ast.NewIdent("Next"),
					})),
				),
			),
			nil,
		),
		astutil.Return(
			ast.NewIdent("false"),
		),
	)

	advancesig := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				astutil.Field(&ast.Ellipsis{Elt: t.tf.Type}, t.tf.Names[0]),
			},
		},
		Results: &ast.FieldList{
			List: []*ast.Field{
				t.scanner.Type.Results.List[0],
				astutil.Field(&ast.ArrayType{Elt: t.tf.Type}),
				astutil.Field(ast.NewIdent("bool")),
			},
		},
	}

	genscanning := func() *ast.BlockStmt {
		stmts := make([]ast.Stmt, 0, t.n)
		stmts = append(stmts, astutil.CaseClause(
			astutil.ExprList(astutil.IntegerLiteral(0)),
			astutil.Return(
				ast.NewIdent("nil"),
				astutil.CallExpr(
					&ast.ArrayType{Elt: t.tf.Type},
					ast.NewIdent("nil"),
				),
				ast.NewIdent("false"),
			),
		))

		for i := 1; len(stmts) < cap(stmts); i++ {
			stmts = append(stmts, astutil.CaseClause(
				astutil.ExprList(astutil.IntegerLiteral(i)),
				astutil.Return(
					astutil.CallExpr(
						t.scanner.Name,
						astutil.CallExprEllipsis(
							&ast.SelectorExpr{
								X: astutil.SelExpr(
									"t",
									"q",
								),
								Sel: ast.NewIdent("QueryContext"),
							},
							ast.NewIdent("t.ctx"),
							ast.NewIdent("query"),
							&ast.SliceExpr{
								X: ast.NewIdent("tmp"),
							},
						),
					),
					astutil.CallExpr(
						&ast.ArrayType{Elt: t.tf.Type},
						ast.NewIdent("nil"),
					),
					ast.NewIdent("false"),
				),
			))
		}

		stmts = append(stmts, astutil.CaseClause(
			nil,
			astutil.Return(
				astutil.CallExpr(
					t.scanner.Name,
					astutil.CallExprEllipsis(
						&ast.SelectorExpr{
							X: astutil.SelExpr(
								"t",
								"q",
							),
							Sel: ast.NewIdent("QueryContext"),
						},
						ast.NewIdent("t.ctx"),
						ast.NewIdent("query"),
						&ast.SliceExpr{
							X: ast.NewIdent("tmp"),
						},
					),
				),
				&ast.SliceExpr{
					X:   t.tf.Names[0],
					Low: astutil.IntegerLiteral(t.n),
				},
				ast.NewIdent("true"),
			),
		))
		return astutil.Block(stmts...)
	}

	advancefn := functions.NewFn(
		astutil.Switch(
			nil,
			astutil.CallExpr(
				ast.NewIdent("len"),
				ast.NewIdent(t.tf.Names[0].String()),
			),
			genscanning(),
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
