package crud

import (
	"fmt"
	"go/ast"
	"io"

	"github.com/serenize/snaker"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
)

func NewFunctions(c genieql.Configuration, details genieql.TableDetails, pkg, prefix string, options ...generators.QueryFunctionOption) genieql.Generator {
	return funcGenerator{
		TableDetails: details,
		Package:      pkg,
		Prefix:       prefix,
		Options:      options,
	}
}

type funcGenerator struct {
	genieql.TableDetails
	Package string
	Prefix  string
	Options []generators.QueryFunctionOption
}

func (t funcGenerator) Generate(dst io.Writer) error {
	mg := make([]genieql.Generator, 0, 10)
	names := genieql.ColumnInfoSet(t.TableDetails.Columns).ColumnNames()
	query := t.TableDetails.Dialect.Insert(t.TableDetails.Table, names, []string{})
	options := append(
		t.Options,
		generators.QFOParameters(fieldFromColumnInfo(t.TableDetails.Columns...)...),
		generators.QFOBuiltinQuery(query),
		generators.QFOName(fmt.Sprintf("%sInsert", t.Prefix)),
	)

	mg = append(mg, generators.NewQueryFunction(options...))

	for i, column := range t.TableDetails.Columns {

		query = t.TableDetails.Dialect.Select(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Columns[i:i+1]).ColumnNames())
		options = append(
			t.Options,
			generators.QFOBuiltinQuery(query),
			generators.QFOParameters(fieldFromColumnInfo(column)...),
			generators.QFOName(fmt.Sprintf("%sFindBy%s", t.Prefix, snaker.SnakeToCamel(column.Name))),
		)

		mg = append(mg, generators.NewQueryFunction(options...))
	}

	if len(t.TableDetails.Naturalkey) > 0 {
		query = t.TableDetails.Dialect.Update(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Naturalkey).ColumnNames())
		options = append(
			t.Options,
			generators.QFOParameters(fieldFromColumnInfo(t.TableDetails.Naturalkey...)...),
			generators.QFOBuiltinQuery(query),
			generators.QFOName(fmt.Sprintf("%sUpdateByID", t.Prefix)),
		)
		mg = append(mg, generators.NewQueryFunction(options...))

		query = t.TableDetails.Dialect.Delete(t.TableDetails.Table, names, genieql.ColumnInfoSet(t.TableDetails.Naturalkey).ColumnNames())
		options = append(
			t.Options,
			generators.QFOParameters(fieldFromColumnInfo(t.TableDetails.Naturalkey...)...),
			generators.QFOBuiltinQuery(query),
			generators.QFOName(fmt.Sprintf("%sDeleteByID", t.Prefix)),
		)
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
