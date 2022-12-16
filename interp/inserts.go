package genieql

import (
	"fmt"
	"go/ast"
	"go/types"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/generators/functions"
	"github.com/pkg/errors"
)

// Insert configuration interface for generating Insert.
type Insert interface {
	genieql.Generator         // must satisfy the generator interface
	Into(string) Insert       // what table to insert into
	Ignore(...string) Insert  // do not attempt to insert the specified column.
	Default(...string) Insert // use the database default for the specified columns.
	Conflict(string) Insert   // specify how conflicts should be handled.
}

// NewInsert instantiate a new insert generator. it uses the name of function
// that calls Define as the name of the generated function.
func NewInsert(
	ctx generators.Context,
	name string,
	comment *ast.CommentGroup,
	cf *ast.Field,
	qf *ast.Field,
	tf *ast.Field,
	scanner *ast.FuncDecl,
) Insert {
	return &insert{
		ctx:     ctx,
		name:    name,
		comment: comment,
		qf:      qf,
		cf:      cf,
		tf:      tf,
		scanner: scanner,
	}
}

type insert struct {
	ctx      generators.Context
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
func (t *insert) Into(s string) Insert {
	t.table = s
	return t
}

// Default specify the table columns to be given their default values.
func (t *insert) Default(defaults ...string) Insert {
	t.defaults = defaults
	return t
}

// Ignore specify the table columns to ignore during insert.
// - ignored columns should be defaulted in the static columns.
// - ignored columns should not be read from the structures during explode.
// - ignored columns should not be returned by the query.
func (t *insert) Ignore(ignore ...string) Insert {
	t.ignore = ignore
	return t
}

func (t *insert) Conflict(s string) Insert {
	t.conflict = s
	return t
}

func (t *insert) Generate(dst io.Writer) (err error) {
	var (
		mapping    genieql.MappingConfig
		columns    []genieql.ColumnInfo
		cmaps      []genieql.ColumnMap
		fields     []*ast.Field
		qinputs    []ast.Expr
		encodings  []ast.Stmt
		locals     []ast.Spec
		transforms []ast.Stmt
	)

	driver := t.ctx.Driver
	dialect := t.ctx.Dialect
	fset := t.ctx.FileSet

	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	t.ctx.Debugln("insert type", t.ctx.CurrentPackage.Name, t.ctx.CurrentPackage.ImportPath, types.ExprString(t.tf.Type))
	t.ctx.Debugln("insert table", t.table)

	err = t.ctx.Configuration.ReadMap(
		&mapping,
		genieql.MCOPackage(t.ctx.CurrentPackage),
		genieql.MCOType(types.ExprString(t.tf.Type)),
	)

	if err != nil {
		return err
	}

	if columns, err = t.ctx.Dialect.ColumnInformationForTable(t.ctx.Driver, t.table); err != nil {
		return err
	}

	mapping.Apply(
		genieql.MCOColumns(columns...),
	)

	if columns, _, err = mapping.MappedColumnInfo(driver, dialect, fset, t.ctx.CurrentPackage); err != nil {
		return err
	}

	ignored := genieql.ColumnInfoFilterIgnore(t.ignore...)
	defaulted := genieql.ColumnInfoFilterIgnore(t.defaults...)

	cset := genieql.ColumnInfoSet(columns)
	ignoredcset := cset.Filter(ignored)
	projectioncset := ignoredcset.Filter(defaulted)

	if cmaps, _, err = mapping.MapColumns(fset, t.ctx.CurrentPackage, t.tf.Names[0], projectioncset...); err != nil {
		return errors.Wrapf(
			err,
			"failed to map columns for: %s:%s",
			t.ctx.CurrentPackage.Name, types.ExprString(t.tf.Type),
		)
	}

	if locals, encodings, qinputs, err = generators.QueryInputsFromColumnMap(t.ctx, t.scanner, cmaps...); err != nil {
		return errors.Wrap(err, "unable to transform query inputs")
	}

	transforms = []ast.Stmt{
		&ast.DeclStmt{
			Decl: astutil.VarList(locals...),
		},
	}
	transforms = append(transforms, encodings...)

	g1 := generators.NewColumnConstants(
		fmt.Sprintf("%sStaticColumns", t.name),
		genieql.ColumnValueTransformer{
			Defaults:           append(t.defaults, t.ignore...),
			DialectTransformer: dialect.ColumnValueTransformer(),
		},
		cset,
	)

	g2 := generators.NewExploderFunction(
		t.ctx,
		astutil.Field(ast.NewIdent(types.ExprString(t.tf.Type)), ast.NewIdent("arg1")),
		fields,
		generators.QFOName(fmt.Sprintf("%sExplode", t.name)),
	)

	qfn := functions.Query{
		Context:      t.ctx,
		Scanner:      t.scanner,
		Queryer:      t.qf.Type,
		Transforms:   transforms,
		QueryInputs:  qinputs,
		ContextField: t.cf,
		Query: astutil.StringLiteral(
			dialect.Insert(1, t.table, t.conflict, cset.ColumnNames(), ignoredcset.ColumnNames(), append(t.defaults, t.ignore...)),
		),
	}

	sig := &ast.FuncType{
		Params: &ast.FieldList{
			List: astutil.FlattenFields(t.tf),
		},
	}

	return genieql.MultiGenerate(
		g1,
		g2,
		genieql.NewFuncGenerator(func(dst io.Writer) (err error) {
			if err = generators.GenerateComment(generators.DefaultFunctionComment(t.name), t.comment).Generate(dst); err != nil {
				return err
			}

			if err = functions.CompileInto(dst, functions.New(t.name, sig), qfn); err != nil {
				return err
			}

			return nil
		}),
	).Generate(dst)
}
