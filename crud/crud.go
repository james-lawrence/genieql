package crud

import (
	"fmt"
	"io"

	"github.com/serenize/snaker"

	"bitbucket.org/jatone/genieql"
)

// New builds a generator that generates a CRUD scanner and associated
// queries.
func New(c genieql.Configuration, details genieql.TableDetails, pkg, prefix string) genieql.Generator {
	return generator{
		Configuration: c,
		TableDetails:  details,
		Package:       pkg,
		Prefix:        prefix,
	}
}

type generator struct {
	genieql.Configuration
	genieql.TableDetails
	Package string
	Prefix  string
}

func (t generator) Generate(dst io.Writer) error {
	crud := NewCRUDWriter(
		t.Prefix,
		t.TableDetails,
	)

	return crud.Generate(dst)
}

// NewCRUDWriter generates crud queries. implements the genieql.CrudWriter interface.
func NewCRUDWriter(prefix string, details genieql.TableDetails) genieql.Generator {
	return crudWriter{
		prefix:  prefix,
		details: details,
	}
}

type crudWriter struct {
	prefix  string
	details genieql.TableDetails
}

func (t crudWriter) Generate(out io.Writer) error {
	names := genieql.ColumnInfoSet(t.details.Columns).ColumnNames()
	naturalKeyNames := genieql.ColumnInfoSet(t.details.Columns).PrimaryKey().ColumnNames()
	gens := make([]genieql.Generator, 0, 10)

	constName := fmt.Sprintf("%sInsert", t.prefix)
	gens = append(gens, Insert(t.details).Build(1, constName, []string{}))

	for i, column := range t.details.Columns {
		constName = fmt.Sprintf("%sFindBy%s", t.prefix, snaker.SnakeToCamel(column.Name))
		gens = append(gens, Select(t.details).Build(constName, names[i:i+1]))
	}

	if len(naturalKeyNames) > 0 {
		constName = fmt.Sprintf("%sUpdateByID", t.prefix)
		gens = append(gens, Update(t.details).Build(constName, naturalKeyNames))

		constName = fmt.Sprintf("%sDeleteByID", t.prefix)
		gens = append(gens, Delete(t.details).Build(constName, naturalKeyNames))
	}

	return genieql.MultiGenerate(gens...).Generate(out)
}
