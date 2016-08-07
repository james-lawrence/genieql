package scanner

import (
	"fmt"
	"go/ast"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// Generator builds a scanner.
type Generator struct {
	Mappings []genieql.MappingConfig
	genieql.Driver
	Columns []string
	Fields  []*ast.Field
}

func (t Generator) params() []*ast.Field {
	fields := []*ast.Field{}
	for idx, config := range t.Mappings {
		fields = append(fields, typeDeclarationField(
			astutil.Expr(fmt.Sprintf("*%s", config.Type)),
			ast.NewIdent(fmt.Sprintf("arg%d", idx)),
		))
	}

	return fields
}

func (t Generator) mapColumns() ([]genieql.ColumnMap, error) {
	var (
		err     error
		r       []genieql.ColumnMap
		mapping = []genieql.ColumnMap{}
	)
	for _, config := range t.Mappings {
		if r, err = config.Mapper().MapColumns(t.Fields, t.Columns...); err != nil {
			break
		}

		mapping = append(mapping, r...)
	}

	return mapping, err
}
