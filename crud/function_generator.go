package crud

import (
	"fmt"
	"go/ast"
	"go/parser"
	"io"

	"github.com/serenize/snaker"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
)

func NewFunctions(ctx generators.Context, mapper genieql.MappingConfig, queryer string, details genieql.TableDetails, pkg, typ string, scanner, uniqScanner *ast.FuncDecl, fields []*ast.Field) genieql.Generator {
	q, err := parser.ParseExpr(queryer)
	if err != nil {
		return genieql.NewErrGenerator(err)
	}

	return funcGenerator{
		ctx:          ctx,
		mapper:       mapper,
		TableDetails: details,
		Package:      pkg,
		Type:         typ,
		Scanner:      scanner,
		UniqScanner:  uniqScanner,
		Queryer:      q,
		Fields:       fields,
	}
}

type funcGenerator struct {
	genieql.TableDetails
	ctx         generators.Context
	mapper      genieql.MappingConfig
	Package     string
	Type        string
	Scanner     *ast.FuncDecl
	UniqScanner *ast.FuncDecl
	Queryer     ast.Expr
	Fields      []*ast.Field
}

func (t funcGenerator) Generate(dst io.Writer) error {
	mg := make([]genieql.Generator, 0, 10)
	names := genieql.ColumnInfoSet(t.TableDetails.Columns).ColumnNames()
	naturalKey := genieql.ColumnInfoSet(t.TableDetails.Columns).PrimaryKey()
	queryerOption := generators.QFOQueryer("q", t.Queryer)

	query := t.TableDetails.Dialect.Insert(1, t.TableDetails.Table, names, []string{})
	options := []generators.QueryFunctionOption{
		queryerOption,
		generators.QFOName(fmt.Sprintf("%sInsert", t.Type)),
		generators.QFOScanner(t.UniqScanner),
		generators.QFOExplodeStructParam(
			astutil.Field(ast.NewIdent(t.Type), ast.NewIdent("arg1")),
			t.Fields...,
		),
		generators.QFOBuiltinQueryFromString(query),
	}

	mg = append(mg, generators.NewQueryFunction(options...))

	for i, column := range t.TableDetails.Columns {
		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Columns[i:i+1]).ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOParameters(fieldFromColumnInfo(column)...),
			generators.QFOName(fmt.Sprintf("%sFindBy%s", t.Type, snaker.SnakeToCamel(column.Name))),
			generators.QFOScanner(t.Scanner),
		}

		mg = append(mg, generators.NewQueryFunction(options...))
	}

	if len(naturalKey) > 0 {
		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOParameters(fieldFromColumnInfo(naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sFindByKey", t.Type)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(options...))
		mg = append(mg, t.updateFunc(queryerOption, naturalKey, names))
		query = t.TableDetails.Dialect.Delete(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOParameters(fieldFromColumnInfo(naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sDeleteByID", t.Type)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(options...))
	}

	return genieql.MultiGenerate(mg...).Generate(dst)
}

func (t funcGenerator) updateFunc(queryerOption generators.QueryFunctionOption, naturalKey genieql.ColumnInfoSet, names []string) genieql.Generator {
	otherColumns := genieql.ColumnInfoSet(t.TableDetails.Columns).Filter(genieql.NotPrimaryKeyFilter)
	updateFields, _, err := t.mapper.MappedFields(t.ctx.Dialect, t.ctx.FileSet, t.ctx.CurrentPackage, naturalKey.ColumnNames()...)
	if err != nil {
		return genieql.NewErrGenerator(err)
	}
	updateParam := astutil.Field(ast.NewIdent(t.Type), ast.NewIdent("update"))
	query := t.TableDetails.Dialect.Update(t.TableDetails.Table, otherColumns.ColumnNames(), naturalKey.ColumnNames(), names)
	options := []generators.QueryFunctionOption{
		queryerOption,
		generators.QFOParameters2(
			append(fieldFromColumnInfo(naturalKey...), updateParam),
			append(
				generators.StructureQueryParameters(updateParam, updateFields...),
				astutil.MapFieldsToNameExpr(fieldFromColumnInfo(naturalKey...)...)...,
			),
		),
		generators.QFOBuiltinQueryFromString(query),
		generators.QFOName(fmt.Sprintf("%sUpdateByID", t.Type)),
		generators.QFOScanner(t.UniqScanner),
	}

	return generators.NewQueryFunction(options...)
}

func fieldFromColumnInfo(infos ...genieql.ColumnInfo) []*ast.Field {
	r := make([]*ast.Field, 0, len(infos))
	for _, info := range infos {
		r = append(r, astutil.Field(ast.NewIdent(info.Type), ast.NewIdent(info.Name)))
	}
	return r
}
