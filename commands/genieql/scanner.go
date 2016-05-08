package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"log"
	"os"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"gopkg.in/alecthomas/kingpin.v2"
)

type scanners struct{}

func (t *scanners) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("scanner", "generate scanners")
	(&queryLiteral{}).configure(cmd.Command("query-literal", "build a scanner for the provided type/query"))
	(&staticScanner{}).configure(cmd.Command("static", "build a static scanner for the provided type/table"))
	(&dynamicScanner{}).configure(cmd.Command("dynamic", "build a dynamic scanner for the provided type/table"))
	(&staticScanner{}).configure(cmd.Command("default", "build the default (which is a static scanner) for the provided type/table"))

	return cmd
}

func printScanner(output string, generator genieql.ScannerGenerator, pkg *ast.Package) {
	var err error
	printer := genieql.ASTPrinter{}
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()

	if err = genieql.PrintPackage(printer, buffer, fset, pkg, os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	if err = generator.Scanner(buffer, fset); err != nil {
		log.Fatalln(err)
	}

	if err = genieql.FormatOutput(formatted, buffer.Bytes()); err != nil {
		log.Fatalln(err)
	}

	if err = commands.WriteStdoutOrFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, formatted); err != nil {
		log.Fatalln(err)
	}
}
