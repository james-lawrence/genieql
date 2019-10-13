package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/alecthomas/kingpin"
	"github.com/davecgh/go-spew/spew"
)

// ast utility - used to consume a golang file and print the AST.
type astcli struct {
	filepath string
}

func (t *astcli) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("ast", "consume a golang file and print its AST").Action(t.execute)
	cli.Arg("filepath", "file to print").StringVar(&t.filepath)
	return cli
}

func (t *astcli) execute(*kingpin.ParseContext) (err error) {
	var (
		filenode *ast.File
		fset     = token.NewFileSet()
	)

	if filenode, err = parser.ParseFile(fset, t.filepath, nil, parser.ParseComments); err != nil {
		return err
	}

	fmt.Println(spew.Sdump(filenode))
	// ast.Inspect(filenode, func(n ast.Node) bool {
	// 	if n != nil {
	// 		fmt.Println(spew.Sdump(n))
	// 	}
	// 	return true
	// })

	return nil
}
