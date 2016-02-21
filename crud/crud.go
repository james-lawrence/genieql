package crud

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/token"
	"io"
	"log"
	"strings"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
	"bitbucket.org/jatone/genieql/sqlutil"
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

	q, err := sqlutil.LookupColumnQuery(t.Configuration.Dialect, t.Table)
	if err != nil {
		log.Println("failure looking up column query", err)
		return nil, err
	}

	if columns, err = genieql.Columns(db, q); err != nil {
		return nil, err
	}

	primaryKeyQuery, err := sqlutil.LookupPrimaryKeyQuery(t.Configuration.Dialect, t.Table)
	if err != nil {
		log.Println("failure looking up primary key query", err)
		return nil, err
	}

	if naturalKey, err = genieql.ExtractPrimaryKey(db, primaryKeyQuery); err != nil {
		return nil, err
	}

	generator := scanner.Generator{
		Configuration: t.Configuration,
		MappingConfig: t.MappingConfig,
		Columns:       columns,
		Name:          fmt.Sprintf("%sCrudScanner", strings.Title(t.MappingConfig.Type)),
	}

	crud := genieql.NewCRUDWriter(
		buffer,
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

	if err := crud.Write(genieql.PostgresqlCRUDGen, fset); err != nil {
		log.Println("crud", err)
		return nil, err
	}

	return genieql.FormatOutput(buffer.Bytes())
}
