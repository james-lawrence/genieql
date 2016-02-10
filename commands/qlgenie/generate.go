package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/imports"

	_ "github.com/lib/pq"
	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
)

type generateCrud struct {
	packageType string
	table       string
	output      string
}

func (t *generateCrud) Execute(*kingpin.ParseContext) error {
	var err error
	var db *sql.DB
	var columns []string
	buffer := bytes.NewBuffer([]byte{})
	log.Println("package.Type", t.packageType)
	log.Println("table", t.table)

	pkgName, typName := extractPackageType(t.packageType)
	packages, err := genieql.LocatePackage(pkgName)

	if err != nil {
		log.Fatalln("Failed to locate package", err)
	}

	naturalkey := []string{"id"}

	if db, err = genieql.ConnectDB(); err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT * FROM %s LIMIT 1", t.table)
	log.Println("column query", q)
	if columns, err = genieql.ExtractColumns(db, q); err != nil {
		return err
	}

	if err = genieql.AmbiguityCheck(columns...); err != nil {
		return err
	}

	decls := genieql.FilterDeclarations(genieql.FilterType(typName), packages...)

	switch len(decls) {
	case 1:
	// happy case, fallthrough
	case 0:
		log.Fatalln("Failed to locate", pkgName, typName)
	default:
		log.Fatalln("Ambiguous type, located multiple matches", decls)
	}

	typeDecl := decls[0]

	mer := genieql.MapperV2{Aliasers: []genieql.Aliaser{genieql.AliaserBuilder("snakecase", "lowercase")}}
	fields := genieql.ExtractFields(typeDecl.Specs[0]).List

	columnMap, err := mer.MapColumns(&ast.Ident{Name: "arg0"}, fields, columns...)

	if err != nil {
		log.Fatalln("failed to map columns", err)
	}

	file := &ast.File{
		Name: &ast.Ident{Name: "sso"},
	}
	scanner := genieql.Scanner{Name: "CrudScanner"}.Build(columnMap, genieql.Ident(typName))
	crud := genieql.NewCRUDWriter(
		buffer,
		typName,
		t.table,
		naturalkey,
		columns,
	)

	fset := token.NewFileSet()
	if err := printer.Fprint(buffer, fset, file); err != nil {
		log.Fatalln("package", err)
	}

	if _, err := fmt.Fprintf(buffer, genieql.Preface, strings.Join(os.Args[1:], " ")); err != nil {
		log.Fatalln("preface", err)
	}

	if err := printer.Fprint(buffer, fset, scanner); err != nil {
		log.Fatalln("scanner", err)
	}
	fmt.Fprintf(buffer, "\n\n")

	if err := crud.Write(genieql.PostgresqlCRUDGen, fset); err != nil {
		log.Fatalln("crud", err)
	}

	raw, err := imports.Process("", buffer.Bytes(), nil)
	if err != nil {
		log.Fatalln("imports", err)
	}

	raw, err = format.Source(raw)
	if err != nil {
		log.Fatalln("format", err)
	}

	f, err := os.OpenFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalln(err)
	}

	f.Write(raw)
	fmt.Println(string(raw))

	return nil
}

type generate struct {
	crud *generateCrud
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")
	crud := cmd.Command("crud", "generate crud queries (INSERT, SELECT, UPDATE, DELETE)").Action(t.crud.Execute)
	crud.Arg("output", "path of output file").Required().StringVar(&t.crud.output)
	crud.Arg("package.Type", "package prefixed structure we want to build the scanner/query for").Required().
		StringVar(&t.crud.packageType)
	crud.Arg("table", "table you want to build the queries for").Required().
		StringVar(&t.crud.table)
	return cmd
}
