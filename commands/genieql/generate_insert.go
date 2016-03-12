package main

import (
	"bytes"
	"fmt"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/crud"
)

type generateInsert struct {
	configName  string
	constSuffix string
	packageType string
	table       string
	output      string
	defaults    []string
}

func (t *generateInsert) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	pkgName, typName := extractPackageType(t.packageType)

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.configName), &configuration); err != nil {
		return err
	}

	details, err := genieql.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	constName := fmt.Sprintf("%sInsert%s", typName, t.constSuffix)

	fset := token.NewFileSet()
	buffer := bytes.NewBuffer([]byte{})
	formatted := bytes.NewBuffer([]byte{})
	printer := genieql.ASTPrinter{}

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		return err
	}

	if err := genieql.PrintPackage(printer, buffer, fset, pkg, os.Args[1:]); err != nil {
		log.Fatalln("PrintPackage failed:", err)
	}

	if err := crud.Insert(details).Build(constName, t.defaults).Generate(buffer, fset); err != nil {
		log.Fatalln("insert generation failed:", err)
	}

	if err := genieql.FormatOutput(formatted, buffer.Bytes()); err != nil {
		log.Fatalln("format output failed:", err)
	}

	if err = commands.WriteStdoutOrFile(t.output, os.O_CREATE|os.O_TRUNC|os.O_RDWR, formatted); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *generateInsert) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	insert := cmd.Command("insert", "generate more complicated insert queries that can be used by the crud scanner").Action(t.Execute)

	insert.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.configName)

	insert.Flag(
		"suffix",
		"suffix for the name of the generated constant",
	).Required().StringVar(&t.constSuffix)

	insert.Flag("default", "specifies a name of a column to default to database value").
		StringsVar(&t.defaults)

	insert.Flag(
		"output",
		"path of output file",
	).Default("").StringVar(&t.output)

	insert.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	insert.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	return insert
}
