package genieql

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
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
	Ignore(...string) Insert  // ignore the specified columns.
	Default(...string) Insert // use the database default for the specified columns.
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
		n:       1,
	}
}

type insert struct {
	ctx      generators.Context
	n        int // number of records to support inserting
	name     string
	table    string
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

// Ingore specify the table columns to ignore.
func (t *insert) Ignore(ignore ...string) Insert {
	t.ignore = ignore
	return t
}

// Batch specify the maximum number of records to insert.
func (t *insert) Batch(size int) Insert {
	t.n = size
	return t
}

func (t *insert) Generate(dst io.Writer) (err error) {
	var (
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
		fields  []*ast.Field
	)
	driver := t.ctx.Driver
	dialect := t.ctx.Dialect
	fset := t.ctx.FileSet

	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	t.ctx.Debugln("insert type", types.ExprString(t.tf.Type))
	t.ctx.Debugln("insert table", t.table)
	t.ctx.Debugln("insert package", t.ctx.CurrentPackage.Name)
	t.ctx.Debugln("insert package", t.ctx.CurrentPackage.ImportPath)

	err = t.ctx.Configuration.ReadMap(
		"default", // deprecated hopefully we'll be able to drop at some point.
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

	mapping.Apply(genieql.MCOColumns(columns...))

	if columns, _, err = mapping.MappedColumnInfo(driver, dialect, fset, t.ctx.CurrentPackage); err != nil {
		return err
	}

	ignore := genieql.ColumnInfoFilterIgnore(append(t.ignore, t.defaults...)...)
	cset := genieql.ColumnInfoSet(columns)

	if fields, _, err = mapping.MapFieldsToColumns(fset, t.ctx.CurrentPackage, cset.Filter(ignore)...); err != nil {
		return errors.Wrapf(
			err,
			"failed to map fields to columns for: %s:%s",
			t.ctx.CurrentPackage.Name, types.ExprString(t.tf.Type),
		)
	}

	g1 := generators.NewColumnConstants(
		fmt.Sprintf("%sStaticColumns", t.name),
		genieql.ColumnValueTransformer{
			Defaults:           t.defaults,
			DialectTransformer: dialect.ColumnValueTransformer(),
		},
		columns,
	)

	g2 := generators.NewExploderFunction(
		t.ctx,
		astutil.Field(ast.NewIdent(types.ExprString(t.tf.Type)), ast.NewIdent("arg1")),
		fields,
		generators.QFOName(fmt.Sprintf("%sExplode", t.name)),
	)

	qfn := functions.Query{
		Context: t.ctx,
		Query: astutil.StringLiteral(
			dialect.Insert(t.n, t.table, genieql.ColumnInfoSet(columns).ColumnNames(), t.defaults),
		),
		Scanner:      t.scanner,
		Queryer:      t.qf.Type,
		QueryInputs:  astutil.StructureFieldSelectors(t.tf, fields...),
		ContextField: t.cf,
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
			var (
				n ast.Node
			)

			if err = GenerateComment(t.comment, newFunctionComment(t.name)).Generate(dst); err != nil {
				return err
			}

			if n, err = qfn.Compile(functions.New(t.name, sig)); err != nil {
				return err
			}

			return printer.Fprint(dst, token.NewFileSet(), n)
		}),
	).Generate(dst)
}
