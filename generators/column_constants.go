package generators

import (
	"go/ast"
	"go/build"
	"go/types"
	"html/template"
	"io"
	"strings"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
)

// ColumnConstantContext for building column sets.
// TODO consider turning this into a general generator context.
type ColumnConstantContext struct {
	Config  genieql.Configuration
	Package *build.Package
}

// NewColumnConstantFromFieldList generates column constants from field list.
func NewColumnConstantFromFieldList(ctx Context, name string, trans genieql.ColumnTransformer, fields *ast.FieldList, ignoreset ...string) genieql.Generator {
	var (
		infos []genieql.ColumnInfo
	)

	for _, param := range fields.List {
		i, err := columnInfo(ctx, param)
		if err != nil {
			return genieql.NewErrGenerator(err)
		}
		infos = append(infos, i...)
	}
	return NewColumnConstants(name, trans, infos)
}

// NewColumnConstants builds a generator from a set of ColumnInfo.
func NewColumnConstants(name string, trans genieql.ColumnTransformer, columns []genieql.ColumnInfo) genieql.Generator {
	return constants{
		Name:        name,
		Columns:     columns,
		Transformer: trans,
	}
}

type constants struct {
	Name        string
	Columns     []genieql.ColumnInfo
	Transformer genieql.ColumnTransformer
}

func (t constants) Generate(dst io.Writer) error {
	const templatename = "column-constants"
	type context struct {
		Name    string
		Columns []genieql.ColumnInfo
	}
	var (
		err error
		ctx = context{
			Name:    t.Name,
			Columns: t.Columns,
		}
	)

	funcMap := template.FuncMap{
		"transform": transformer{ColumnTransformer: t.Transformer}.transform,
		"columns": func(i []string) string {
			return strings.Join(i, ",")
		},
	}

	tmpl := template.Must(template.New(templatename).Funcs(funcMap).Parse(columnConstantsTemplate))
	if err = tmpl.Execute(dst, ctx); err != nil {
		return errors.Wrap(err, "failed to generate columns constant")
	}

	_, err = dst.Write([]byte("\n"))

	return errors.Wrap(err, "")
}

const columnConstantsTemplate = `const {{.Name}} = "{{ .Columns | transform | columns}}"`

type transformer struct {
	genieql.ColumnTransformer
}

func (t transformer) transform(m []genieql.ColumnInfo) []string {
	s := make([]string, 0, len(m))
	for _, c := range m {
		s = append(s, t.ColumnTransformer.Transform(c))
	}
	return s
}

func columnInfo(ctx Context, param *ast.Field) ([]genieql.ColumnInfo, error) {
	if builtinType(param.Type) {
		return builtinParamColumnInfo(param)
	}
	_, info, err := mappedParam(ctx, param)
	return info, err
}

// builtinParamColumnInfo converts a *ast.Field that represents a builtin type
// (time.Time,int,float,bool, etc) into an array of ColumnInfo.
func builtinParamColumnInfo(param *ast.Field) ([]genieql.ColumnInfo, error) {
	columns := make([]genieql.ColumnInfo, 0, len(param.Names))
	for _, name := range param.Names {
		columns = append(columns, genieql.ColumnInfo{
			Name: name.Name,
			Type: types.ExprString(param.Type),
		})
	}
	return columns, nil
}

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnMap.
func mappedParam(ctx Context, param *ast.Field) (genieql.MappingConfig, []genieql.ColumnInfo, error) {
	var (
		err   error
		infos []genieql.ColumnInfo
		m     genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return m, infos, err
	}

	infos, _, err = m.MappedColumnInfo2(ctx.Dialect, ctx.FileSet, ctx.CurrentPackage)
	return m, infos, err
}

// converts a *ast.Field that represents a struct into a list of fields that map
// to columns.
func mappedFields(ctx Context, param *ast.Field, ignoreSet ...string) ([]*ast.Field, error) {
	var (
		err   error
		infos []*ast.Field
		m     genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return infos, err
	}

	infos, _, err = m.MappedFields(ctx.Dialect, ctx.FileSet, ctx.CurrentPackage, ignoreSet...)
	return infos, err
}

func mappedStructure(ctx Context, param *ast.Field, ignoreSet ...string) ([]genieql.ColumnInfo, []*ast.Field, error) {
	var (
		err     error
		infos   []*ast.Field
		columns []genieql.ColumnInfo
		m       genieql.MappingConfig
	)

	if err = ctx.Configuration.ReadMap(packageName(ctx.CurrentPackage, param.Type), types.ExprString(unwrapExpr(param.Type)), "default", &m); err != nil {
		return columns, infos, err
	}

	if columns, err = m.ColumnInfo(ctx.Dialect); err != nil {
		return columns, infos, err
	}

	infos, _, err = m.MapFieldsToColumns(
		ctx.FileSet,
		ctx.CurrentPackage,
		genieql.ColumnInfoSet(columns).Filter(genieql.ColumnInfoFilterIgnore(ignoreSet...))...,
	)

	return columns, infos, err
}
