package crud

import (
	"database/sql"
	"fmt"
	"go/token"
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
	Table   string
	Package string
	Prefix  string
}

func (t generator) Generate(dst io.Writer, fset *token.FileSet) error {
	crud := NewCRUDWriter(
		dst,
		t.Prefix,
		t.TableDetails,
	)

	return crud.Write(fset)
}

// LoadInformation loads table information based on the configuration and
// table name.
func LoadInformation(configuration genieql.Configuration, table string) (genieql.TableDetails, error) {
	var err error
	var db *sql.DB
	var dialect genieql.Dialect
	var details genieql.TableDetails

	if db, err = genieql.ConnectDB(configuration); err != nil {
		return details, err
	}

	dialect, err = genieql.LookupDialect(configuration.Dialect)
	if err != nil {
		return details, err
	}

	details, err = genieql.LookupTableDetails(db, dialect, table)
	return details, err
}

// NewCRUDWriter generates crud queries. implements the genieql.CrudWriter interface.
func NewCRUDWriter(out io.Writer, prefix string, details genieql.TableDetails) genieql.CrudWriter {
	return crudWriter{
		out:     out,
		prefix:  prefix,
		details: details,
	}
}

type crudWriter struct {
	out     io.Writer
	prefix  string
	details genieql.TableDetails
}

func (t crudWriter) Write(fset *token.FileSet) error {
	constName := fmt.Sprintf("%sInsert", t.prefix)
	if err := Insert(t.details).Build(constName, []string{}).Generate(t.out, fset); err != nil {
		return err
	}

	for i, column := range t.details.Columns {
		constName = fmt.Sprintf("%sFindBy%s", t.prefix, snaker.SnakeToCamel(column))
		if err := Select(t.details).Build(constName, t.details.Columns[i:i+1]).Generate(t.out, fset); err != nil {
			return err
		}
	}

	constName = fmt.Sprintf("%sUpdateByID", t.prefix)
	if err := Update(t.details).Build(constName, t.details.Naturalkey).Generate(t.out, fset); err != nil {
		return err
	}

	constName = fmt.Sprintf("%sDeleteByID", t.prefix)
	if err := Delete(t.details).Build(constName, t.details.Naturalkey).Generate(t.out, fset); err != nil {
		return err
	}

	return nil
}
