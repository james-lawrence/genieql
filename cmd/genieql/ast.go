package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"bitbucket.org/jatone/genieql"
	"github.com/alecthomas/kingpin"
	"github.com/davecgh/go-spew/spew"
)

// ast utility - used to consume a golang file and print the AST.
type astcli struct {
	format    string
	filepaths []string
}

func (t *astcli) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("ast", "consume a golang file and print its AST").Action(t.execute)
	cli.Flag("format", "output format. can be ast, spew, human").Default("ast").EnumVar(&t.format, "ast", "spew", "human")
	cli.Arg("filepaths", "files to print").StringsVar(&t.filepaths)
	return cli
}

func (t *astcli) execute(*kingpin.ParseContext) (err error) {
	for _, filepath := range t.filepaths {
		switch t.format {
		case "spew":
			if err = t.spewprint(t.parse(filepath)); err != nil {
				return err
			}
		case "human":
			if err = t.humanprint(t.parse(filepath)); err != nil {
				return err
			}
		default:
			if err = t.astprint(t.parse(filepath)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *astcli) parse(filename string) (fset *token.FileSet, filenode *ast.File, err error) {
	fset = token.NewFileSet()
	if filenode, err = parser.ParseFile(fset, filename, nil, parser.ParseComments); err != nil {
		return fset, filenode, err
	}
	return fset, filenode, err
}

func (t *astcli) spewprint(fset *token.FileSet, filenode *ast.File, err error) error {
	if err != nil {
		return err
	}
	_, err = fmt.Println(spew.Sdump(filenode))
	return err
}

func (t *astcli) astprint(fset *token.FileSet, filenode *ast.File, err error) error {
	if err != nil {
		return err
	}

	return ast.Print(fset, filenode)
}

func (t *astcli) humanprint(fset *token.FileSet, filenode *ast.File, err error) error {
	if err != nil {
		return err
	}

	printer := genieql.ASTPrinter{}
	printer.FprintAST(os.Stdout, fset, filenode)

	return printer.Err()
}
