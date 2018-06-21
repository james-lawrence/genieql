package generators_test

import (
	"go/ast"
	"go/parser"
	"path/filepath"
	"runtime"

	"bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	localdirectory string
	localfile      string
)

func TestGenerators(t *testing.T) {
	var (
		file string
		ok   bool
	)

	if _, file, _, ok = runtime.Caller(0); !ok {
		t.Error("failed to resolve file")
		t.FailNow()
	}

	localdirectory = filepath.Dir(file)
	localfile = filepath.Join(localdirectory, "foo.go")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Generators Suite")
}

type noopDriver struct{}

func (t noopDriver) LookupNullableType(x ast.Expr) ast.Expr {
	return x
}

func (t noopDriver) NullableType(typ, from ast.Expr) (ast.Expr, bool) {
	return typ, false
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
	return nil
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
