package generators_test

import (
	"go/ast"
	"go/parser"
	"io/ioutil"
	"log"
	"reflect"

	"bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenerators(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generators Suite")
}

type noopDriver struct{}

func (t noopDriver) LookupType(s string) (td genieql.NullableTypeDefinition, b bool) { return td, b }
func (t noopDriver) LookupNullableType(x ast.Expr) ast.Expr {
	return x
}

func (t noopDriver) NullableType(typ, from ast.Expr) (ast.Expr, bool) {
	return typ, false
}

func (t noopDriver) Exported() map[string]reflect.Value {
	return map[string]reflect.Value{}
}

func mustParseExpr(s string) ast.Expr {
	x, err := parser.ParseExpr(s)
	panicOnError(err)
	return x
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

type dialect struct{}

func (t dialect) Insert(n int, table string, columns, defaults []string) string {
	return "INSERT QUERY"
}

func (t dialect) Select(table string, columns, predicates []string) string {
	return "SELECT QUERY"
}

func (t dialect) Update(table string, columns, predicates, returning []string) string {
	return "INSERT QUERY"
}

func (t dialect) Delete(table string, columns, predicates []string) string {
	return "INSERT QUERY"
}

func (t dialect) ColumnValueTransformer() genieql.ColumnTransformer {
	return genieql.NewColumnInfoNameTransformer("")
}

func (t dialect) ColumnNameTransformer() genieql.ColumnTransformer {
	return genieql.NewColumnInfoNameTransformer("")
}

func (t dialect) ColumnInformationForTable(table string) ([]genieql.ColumnInfo, error) {
	switch table {
	case "struct_a":
		return []genieql.ColumnInfo{
			{Name: "a", Type: "int"},
			{Name: "b", Type: "int"},
			{Name: "c", Type: "int"},
			{Name: "d", Type: "bool"},
			{Name: "e", Type: "bool"},
			{Name: "f", Type: "bool"},
		}, nil
	default:
		return []genieql.ColumnInfo(nil), nil
	}
}

func (t dialect) ColumnInformationForQuery(query string) ([]genieql.ColumnInfo, error) {
	return []genieql.ColumnInfo(nil), nil
}
