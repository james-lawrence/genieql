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

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
	"bitbucket.org/jatone/genieql/sqlutil"
	"golang.org/x/tools/imports"
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
	buffer := bytes.NewBuffer([]byte{})

	if db, err = genieql.ConnectDB(t.Configuration); err != nil {
		return nil, err
	}

	q, err := sqlutil.LookupColumnQuery(t.Configuration.Dialect, t.Table)
	if err != nil {
		log.Println("failure looking up column query", err)
		return nil, err
	}

	if columns, err = genieql.ExtractColumns(db, q); err != nil {
		return nil, err
	}

	if err = genieql.AmbiguityCheck(columns...); err != nil {
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
		t.MappingConfig.NaturalKey,
		columns,
	)

	fset := token.NewFileSet()

	if err := generator.Scanner(buffer, fset); err != nil {
		log.Println("scanner", err)
		return nil, err
	}

	if err := crud.Write(genieql.PostgresqlCRUDGen, fset); err != nil {
		log.Println("crud", err)
		return nil, err
	}
	log.Println("Raw:", string(buffer.Bytes()))

	raw, err := imports.Process("", buffer.Bytes(), nil)
	if err != nil {
		log.Println("imports", err)
		return nil, err
	}

	raw, err = format.Source(raw)
	if err != nil {
		log.Println("format", err)
		return nil, err
	}

	return bytes.NewReader(raw), err
}
