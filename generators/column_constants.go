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
	Config    genieql.Configuration
	Package   *build.Package
	IgnoreSet []string
}

// NewColumnConstantFromFieldList generates column constants from field list.
func NewColumnConstantFromFieldList(ctx ColumnConstantContext, name string, trans genieql.ColumnTransformer, fields *ast.FieldList) genieql.Generator {
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

func columnInfo(ctx ColumnConstantContext, param *ast.Field) ([]genieql.ColumnInfo, error) {
	if builtinType(param.Type) {
		return builtinParamColumnInfo(param)
	}
	return mappedParam(ctx, param)
}

// builtinParamColumnInfo converts a *ast.Field that represents a builtin type
// (time.Time, int,float,bool, etc) into an array of ColumnInfo.
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
func mappedParam(ctx ColumnConstantContext, param *ast.Field) ([]genieql.ColumnInfo, error) {
	var (
		err   error
		infos []genieql.ColumnInfo
		m     genieql.MappingConfig
	)

	if err = ctx.Config.ReadMap(packageName(ctx.Package, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return infos, err
	}

	return m.ColumnInfo()
}