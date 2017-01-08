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

func NewFunctions(c genieql.Configuration, queryer string, details genieql.TableDetails, pkg, prefix string, scanner, uniqScanner *ast.FuncDecl) genieql.Generator {
	q, err := parser.ParseExpr(queryer)
	if err != nil {
		return genieql.NewErrGenerator(err)
	}

	return funcGenerator{
		TableDetails: details,
		Package:      pkg,
		Prefix:       prefix,
		Scanner:      scanner,
		UniqScanner:  uniqScanner,
		Queryer:      q,
	}
}

type funcGenerator struct {
	genieql.TableDetails
	Package     string
	Prefix      string
	Scanner     *ast.FuncDecl
	UniqScanner *ast.FuncDecl
	Queryer     ast.Expr
}

func (t funcGenerator) Generate(dst io.Writer) error {
	mg := make([]genieql.Generator, 0, 10)
	names := genieql.ColumnInfoSet(t.TableDetails.Columns).ColumnNames()
	naturalKey := genieql.ColumnInfoSet(t.TableDetails.Columns).PrimaryKey()
	queryerOption := generators.QFOQueryer("q", t.Queryer)

	query := t.TableDetails.Dialect.Insert(1, t.TableDetails.Table, names, []string{})
	options := []generators.QueryFunctionOption{
		queryerOption,
		generators.QFOName(fmt.Sprintf("%sInsert", t.Prefix)),
		generators.QFOScanner(t.UniqScanner),
		generators.QFOParameters(fieldFromColumnInfo(t.TableDetails.Columns...)...),
		generators.QFOBuiltinQueryFromString(query),
	}

	mg = append(mg, generators.NewQueryFunction(options...))

	for i, column := range t.TableDetails.Columns {
		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Columns[i:i+1]).ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOParameters(fieldFromColumnInfo(column)...),
			generators.QFOName(fmt.Sprintf("%sFindBy%s", t.Prefix, snaker.SnakeToCamel(column.Name))),
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
			generators.QFOName(fmt.Sprintf("%sFindByKey", t.Prefix)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(options...))

		query = t.TableDetails.Dialect.Update(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOParameters(fieldFromColumnInfo(naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sUpdateByID", t.Prefix)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(options...))

		query = t.TableDetails.Dialect.Delete(t.TableDetails.Table, names, naturalKey.ColumnNames())
		options = []generators.QueryFunctionOption{
			queryerOption,
			generators.QFOParameters(fieldFromColumnInfo(naturalKey...)...),
			generators.QFOBuiltinQueryFromString(query),
			generators.QFOName(fmt.Sprintf("%sDeleteByID", t.Prefix)),
			generators.QFOScanner(t.UniqScanner),
		}
		mg = append(mg, generators.NewQueryFunction(options...))
	}

	return genieql.MultiGenerate(mg...).Generate(dst)
}

func fieldFromColumnInfo(infos ...genieql.ColumnInfo) []*ast.Field {
	r := make([]*ast.Field, 0, len(infos))
	for _, info := range infos {
		r = append(r, astutil.Field(ast.NewIdent(info.Type), ast.NewIdent(info.Name)))
	}
	return r
}
