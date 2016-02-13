package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"

	"gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	var files []string

	app := kingpin.New("printspike", "prints ast of a file to stdout")
	app.Arg("types", "types that will be filled in by the scanner").StringsVar(&files)

	kingpin.MustParse(app.Parse(os.Args[1:]))
	for _, s := range files {
		printspike(s)
	}

	// printspike("example1.go")
	// // printspike("example2.go")
	// // printspike("example3.go")
}

func printspike(filename string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)

	if err != nil {
		log.Fatalln(err)
	}

	ast.Print(fset, f)
}
