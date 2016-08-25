package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"path/filepath"
	"strings"

	"bitbucket.org/jatone/genieql"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type functionScanner struct {
	functionType string
}

func (t *functionScanner) Execute(*kingpin.ParseContext) error {
	fset := token.NewFileSet()
	pkgName, typName := extractPackageType(t.functionType)

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	spec, err := genieql.NewUtils(fset).FindUniqueType(genieql.FilterName(typName), pkg)
	log.Println("error", err)
	x := spec.Type.(*ast.FuncType)

	mapNames := func(x ...*ast.Ident) []string {
		r := make([]string, 0, len(x))
		for _, n := range x {
			r = append(r, n.Name)
		}
		return r
	}

	log.Println("Package", pkg.Name, pkg.Imports)
	log.Printf("spec %s, %#v\n", spec.Name, spec.Type.(*ast.FuncType).Func)
	for _, params := range x.Params.List {
		log.Printf("Type %T\n", params.Type)

		log.Println(params.Type, strings.Join(mapNames(params.Names...), ","))
	}

	// for _, params := range x.Results.List {
	// 	log.Println(params.Type, strings.Join(mapNames(params.Names...), ","))
	// }

	return nil
}

func (t *functionScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cmd.Arg(
		"package.FunctionType",
		"package prefixed function type definition to build a scanner for",
	).Required().StringVar(&t.functionType)

	return cmd.Action(t.Execute)
}
