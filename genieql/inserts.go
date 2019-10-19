package genieql

import (
	"go/ast"
	"go/types"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
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
	texp ast.Expr,
) Insert {
	return &insert{
		ctx:     ctx,
		name:    name,
		comment: comment,
		qf:      qf,
		cf:      cf,
		texp:    texp,
	}
}

type insert struct {
	ctx      generators.Context
	name     string
	defaults []string
	ignore   []string
	texp     ast.Expr // type expression
	cf       *ast.Field
	qf       *ast.Field
	table    string
	comment  *ast.CommentGroup
}

func (t *insert) Into(s string) Insert {
	t.table = s
	return t
}

func (t *insert) Default(defaults ...string) Insert {
	t.defaults = defaults
	return t
}

func (t *insert) Ignore(ignore ...string) Insert {
	t.ignore = ignore
	return t
}

func (t *insert) Generate(dst io.Writer) (err error) {
	var (
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
	)
	driver := t.ctx.Driver
	dialect := t.ctx.Dialect
	fset := t.ctx.FileSet

	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	t.ctx.Println("insert type", types.ExprString(t.texp))
	t.ctx.Println("insert table", t.table)

	err = t.ctx.Configuration.ReadMap(
		t.ctx.CurrentPackage.Name,
		types.ExprString(t.texp),
		"default", // vestigate hopefully we'll be able to drop at some point.
		&mapping,
	)

	if err != nil {
		return err
	}

	if columns, err = t.ctx.Dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	mapping.Apply(genieql.MCOColumns(columns...))

	if columns, _, err = mapping.MappedColumnInfo(driver, dialect, fset, t.ctx.CurrentPackage); err != nil {
		return err
	}
	_ = columns

	ignore := genieql.ColumnInfoFilterIgnore(t.defaults...)
	cset := genieql.ColumnInfoSet(columns)

	if fields, _, err = mapping.MapFieldsToColumns(fset, pkg, cset.Filter(ignore)...); err != nil {
		return errors.Wrapf(err, "failed to map fields to columns for: %s", t.packageType)
	}

	return nil
}
