package crud

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io"
	"log"
	"os"
	"strings"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/sqlutil"
	"golang.org/x/tools/imports"
)

// New builds a generator that generates a CRUD scanner and associated
// queries.
func New(c genieql.Configuration, m genieql.MappingConfig) genieql.Generator {
	return generator{
		Configuration: c,
		MappingConfig: m,
	}
}

type generator struct {
	genieql.Configuration
	genieql.MappingConfig
}

func (t generator) Generate() (io.Reader, error) {
	var err error
	var db *sql.DB
	var columns []string
	buffer := bytes.NewBuffer([]byte{})

	if db, err = genieql.ConnectDB(t.Configuration); err != nil {
		return nil, err
	}

	packages, err := genieql.LocatePackage(t.MappingConfig.Package)
	if err != nil {
		log.Println("Failed to locate package", err)
		return nil, err
	}

	decls := genieql.FilterDeclarations(genieql.FilterType(t.MappingConfig.Type), packages...)

	q, err := sqlutil.LookupColumnQuery(t.Configuration.Dialect, t.Table)
	if err != nil {
		log.Println("failure looking up column query", err)
		return nil, err
	}

	log.Println("column query", q)
	if columns, err = genieql.ExtractColumns(db, q); err != nil {
		return nil, err
	}

	if err = genieql.AmbiguityCheck(columns...); err != nil {
		return nil, err
	}

	switch len(decls) {
	case 1:
	// happy case, fallthrough
	case 0:
		return nil, fmt.Errorf("failed to locate: %s.%s", t.MappingConfig.Package, t.MappingConfig.Type)
	default:
		return nil, fmt.Errorf("ambiguous type, located multiple matches: %v", decls)
	}

	typeDecl := decls[0]

	mer := genieql.Mapper{Aliasers: []genieql.Aliaser{genieql.AliaserBuilder(t.MappingConfig.Transformations...)}}
	fields := genieql.ExtractFields(typeDecl.Specs[0]).List

	columnMap, err := mer.MapColumns(&ast.Ident{Name: "arg0"}, fields, columns...)

	if err != nil {
		log.Println("failed to map columns", err)
		return nil, err
	}

	file := &ast.File{
		Name: genieql.Ident("sso"),
	}
	scanner := genieql.Scanner{Name: "CrudScanner"}.Build(columnMap, genieql.Ident(t.MappingConfig.Type))
	crud := genieql.NewCRUDWriter(
		buffer,
		t.MappingConfig.Type,
		t.MappingConfig.Table,
		t.MappingConfig.NaturalKey,
		columns,
	)

	fset := token.NewFileSet()
	if err := printer.Fprint(buffer, fset, file); err != nil {
		log.Println("package", err)
		return nil, err
	}

	if _, err := fmt.Fprintf(buffer, genieql.Preface, strings.Join(os.Args[1:], " ")); err != nil {
		log.Println("preface", err)
		return nil, err
	}

	if err := printer.Fprint(buffer, fset, scanner); err != nil {
		log.Println("scanner", err)
		return nil, err
	}
	fmt.Fprintf(buffer, "\n\n")

	if err := crud.Write(genieql.PostgresqlCRUDGen, fset); err != nil {
		log.Println("crud", err)
		return nil, err
	}

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
