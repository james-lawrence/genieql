package generators_test

import (
	"log"
	"reflect"

	"bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGenerators(t *testing.T) {
	// log.SetOutput(io.Discard)
	log.SetFlags(log.Flags() | log.Lshortfile)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Generators Suite")
}

type noopDriver struct{}

func (t noopDriver) LookupType(s string) (td genieql.ColumnDefinition, b bool) { return td, b }
func (t noopDriver) Exported() map[string]reflect.Value {
	return map[string]reflect.Value{}
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func mustLookupType(d genieql.ColumnDefinition, err error) genieql.ColumnDefinition {
	if err != nil {
		panic(err)
	}
	return d
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

func (t dialect) ColumnInformationForTable(d genieql.Driver, table string) ([]genieql.ColumnInfo, error) {
	switch table {
	case "struct_a":
		return []genieql.ColumnInfo{
			{Name: "a", Definition: mustLookupType(d.LookupType("int"))},
			{Name: "b", Definition: mustLookupType(d.LookupType("int"))},
			{Name: "c", Definition: mustLookupType(d.LookupType("int"))},
			{Name: "d", Definition: mustLookupType(d.LookupType("bool"))},
			{Name: "e", Definition: mustLookupType(d.LookupType("bool"))},
			{Name: "f", Definition: mustLookupType(d.LookupType("bool"))},
		}, nil
	default:
		return []genieql.ColumnInfo(nil), nil
	}
}

func (t dialect) ColumnInformationForQuery(d genieql.Driver, query string) ([]genieql.ColumnInfo, error) {
	return []genieql.ColumnInfo(nil), nil
}
