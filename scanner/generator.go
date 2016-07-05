package scanner

import (
	"fmt"
	"go/ast"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// Generator builds a scanner.
type Generator struct {
	genieql.MappingConfig
	genieql.Driver
	Columns []string
	Fields  []*ast.Field
}

func (t Generator) params() *ast.Field {
	return typeDeclarationField(
		astutil.Expr(fmt.Sprintf("*%s", t.MappingConfig.Type)),
		ast.NewIdent("arg0"),
	)
}
