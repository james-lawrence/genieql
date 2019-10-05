package genieql

import (
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/x/errorsx"
)

// Structure - configuration interface for generating structures.
type Structure interface {
	genieql.Generator // must satisfy the generator interface
	// From generate the structure based on the record definition.
	From(definition) Structure
	Table(string) definition
	Query(string) definition
	// OptionTransformColumns(x ...func(genieql.ColumnInfo) genieql.ColumnInfo) Structure
}

// NewStructure instantiate a new structure generator. it uses the name of function
// that calls Define as the name of the emitted type.
func NewStructure(ctx generators.Context, name string) Structure {
	return &sconfig{ctx: ctx, name: name}
}

type sconfig struct {
	name string
	d    definition
	ctx  generators.Context
}

func (t *sconfig) Generate(dst io.Writer) error {
	if t.d == nil {
		return errorsx.String("missing definition, unable to generate structure. please call the From method")
	}

	log.Println("defining a struct", t.name)

	return generators.NewStructure(
		generators.StructOptionContext(t.ctx),
		generators.StructOptionName(t.name),
		generators.StructOptionColumnsStrategy(func(genieql.Dialect) ([]genieql.ColumnInfo, error) {
			return t.d.Columns()
		}),
		generators.StructOptionMappingConfigOptions(
			genieql.MCOPackage(t.ctx.CurrentPackage.ImportPath),
		),
	).Generate(dst)
}

func (t *sconfig) From(d definition) Structure {
	t.d = d
	return t
}

func (t sconfig) Table(s string) definition {
	return Table(t.ctx.Dialect, s)
}

func (t sconfig) Query(s string) definition {
	return Query(t.ctx.Dialect, s)
}

// func (t sconfig) OptionTransformColumns(x ...func(genieql.ColumnInfo) genieql.ColumnInfo) Structure {
// 	return t
// 	// return func(s sconfig) sconfig {
// 	// 	return s
// 	// }
// }
//
// func (t sconfig) OptionRenameColumn(from, to string) Structure {
// 	return t
// 	// return func(s sconfig) sconfig {
// 	// 	return s
// 	// }
// }
