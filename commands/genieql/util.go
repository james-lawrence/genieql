package main

import (
	"bytes"
	"go/ast"
	"go/token"
	"log"
	"os"
	"strings"
	"unicode"
	"unicode/utf8"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
)

func defaultIfBlank(s, defaultValue string) string {
	if len(strings.TrimSpace(s)) == 0 {
		return defaultValue
	}
	return s
}

func lowercaseFirstLetter(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func printScanner(output string, generator genieql.Generator, pkg *ast.Package) {
	var err error
	printer := genieql.ASTPrinter{}
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()

	if err = genieql.PrintPackage(printer, buffer, fset, pkg, os.Args[1:]); err != nil {
		log.Fatalln(err)
	}

	if err = generator.Generate(buffer, fset); err != nil {
		log.Fatalln(err)
	}

	if err = genieql.FormatOutput(formatted, buffer.Bytes()); err != nil {
		log.Fatalln(err)
	}

	if err = commands.WriteStdoutOrFile(output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, formatted); err != nil {
		log.Fatalln(err)
	}
}
