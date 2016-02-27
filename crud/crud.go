package crud

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/format"
	"go/token"
	"io"
	"log"
	"strings"

	"github.com/serenize/snaker"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
)

// New builds a generator that generates a CRUD scanner and associated
// queries.
func New(c genieql.Configuration, m genieql.MappingConfig, table string) genieql.Generator {
	return generator{
		Configuration: c,
		MappingConfig: m,
		Table:         table,
	}
}

type generator struct {
	genieql.Configuration
	genieql.MappingConfig
	Table string
}

func (t generator) Generate() (io.Reader, error) {
	var err error
	var db *sql.DB
	var columns []string
	var naturalKey []string

	buffer := bytes.NewBuffer([]byte{})

	if db, err = genieql.ConnectDB(t.Configuration); err != nil {
		return nil, err
	}

	dialect, err := genieql.LookupDialect(t.Configuration.Dialect)
	if err != nil {
		log.Println("unknown dialect", t.Configuration.Dialect)
		return nil, err
	}

	if columns, err = genieql.Columns(db, dialect.ColumnQuery(t.Table)); err != nil {
		return nil, err
	}

	if naturalKey, err = genieql.ExtractPrimaryKey(db, dialect.PrimaryKeyQuery(t.Table)); err != nil {
		return nil, err
	}

	generator := scanner.Generator{
		Configuration: t.Configuration,
		MappingConfig: t.MappingConfig,
		Columns:       columns,
		Name:          fmt.Sprintf("%sCrud", strings.Title(t.MappingConfig.Type)),
	}

	crud := NewCRUDWriter(
		buffer,
		dialect,
		t.MappingConfig.Type,
		t.Table,
		naturalKey,
		columns,
	)

	fset := token.NewFileSet()

	if err := generator.Scanner(buffer, fset); err != nil {
		log.Println("scanner", err)
		return nil, err
	}

	fmt.Fprintf(buffer, "\n\n")

	if err := crud.Write(fset); err != nil {
		log.Println("crud", err)
		return nil, err
	}

	return genieql.FormatOutput(buffer.Bytes())
}

// NewCRUDWriter generates crud queries. implements the genieql.CrudWriter interface.
func NewCRUDWriter(out io.Writer, dialect genieql.Dialect, prefix, table string, naturalkey []string, columns []string) genieql.CrudWriter {
	return crudWriter{
		out:        out,
		dialect:    dialect,
		prefix:     prefix,
		table:      table,
		naturalkey: naturalkey,
		columns:    columns,
	}
}

type crudWriter struct {
	out        io.Writer
	dialect    genieql.Dialect
	prefix     string
	table      string
	naturalkey []string
	columns    []string
}

func (t crudWriter) Write(fset *token.FileSet) error {
	constName := fmt.Sprintf("%sInsert", t.prefix)
	query := t.dialect.Insert(t.table, t.columns, []string{})
	if err := format.Node(t.out, fset, genieql.QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	for i, column := range t.columns {
		constName := fmt.Sprintf("%sFindBy%s", t.prefix, snaker.SnakeToCamel(column))
		query := t.dialect.Select(t.table, t.columns, t.columns[i:i+1])
		if err := format.Node(t.out, fset, genieql.QueryLiteral(constName, query)); err != nil {
			return err
		}
		fmt.Fprintf(t.out, "\n")
	}

	constName = fmt.Sprintf("%sUpdateByID", t.prefix)
	query = t.dialect.Update(t.table, t.columns, t.naturalkey)
	if err := format.Node(t.out, fset, genieql.QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	constName = fmt.Sprintf("%sDeleteByID", t.prefix)
	query = t.dialect.Delete(t.table, t.columns, t.naturalkey)
	if err := format.Node(t.out, fset, genieql.QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	return nil
}
