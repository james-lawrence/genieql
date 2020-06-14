package crud

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/types"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"github.com/serenize/snaker"
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

	mg = append(mg, generators.NewQueryFunction(t.ctx, options...))

	for i, column := range t.TableDetails.Columns {
		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Columns[i:i+1]).ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOSharedParameters(fieldFromColumnInfo(t.ctx, column)...),
		}

		findOptions := append(
			options,
			generators.QFOName(fmt.Sprintf("%sFindBy%s", t.Type, snaker.SnakeToCamel(column.Name))),
			generators.QFOScanner(t.UniqScanner),
		)
		lookupOptions := append(
			options,
			generators.QFOName(fmt.Sprintf("%sLookupBy%s", t.Type, snaker.SnakeToCamel(column.Name))),
			generators.QFOScanner(t.Scanner),
		)

		mg = append(mg, generators.NewQueryFunction(t.ctx, findOptions...))
		mg = append(mg, generators.NewQueryFunction(t.ctx, lookupOptions...))
	}

	if len(naturalKey) > 0 {
		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOSharedParameters(fieldFromColumnInfo(t.ctx, naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sFindByKey", t.Type)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(t.ctx, options...))
		mg = append(mg, t.updateFunc(queryerOption, naturalKey, names))
		query = t.TableDetails.Dialect.Delete(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOSharedParameters(fieldFromColumnInfo(t.ctx, naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sDeleteByID", t.Type)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(t.ctx, options...))
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
		generators.QFOParameters(
			append(fieldFromColumnInfo(t.ctx, naturalKey...), updateParam),
			append(
				generators.StructureQueryParameters(updateParam, updateFields...),
				astutil.MapFieldsToNameExpr(fieldFromColumnInfo(t.ctx, naturalKey...)...)...,
			),
		),
		generators.QFOBuiltinQueryFromString(query),
		generators.QFOName(fmt.Sprintf("%sUpdateByID", t.Type)),
		generators.QFOScanner(t.UniqScanner),
	}

	return generators.NewQueryFunction(t.ctx, options...)
}

func fieldFromColumnInfo(ctx generators.Context, infos ...genieql.ColumnInfo) []*ast.Field {
	r := make([]*ast.Field, 0, len(infos))
	for _, info := range infos {
		ident := ast.NewIdent(info.Type)
		if d, err := ctx.Driver.LookupType(info.Type); err == nil {
			ident = ast.NewIdent(d.Native)
		}

		log.Println("fieldFromColumnInfo", types.ExprString(ident), info.Name)
		r = append(r, astutil.Field(ident, ast.NewIdent(info.Name)))
	}
	return r
}
