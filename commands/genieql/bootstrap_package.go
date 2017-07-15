package main

import (
	"go/build"
	"log"

	"bitbucket.org/jatone/genieql"
	pbootstrap "bitbucket.org/jatone/genieql/bootstrap"
	"bitbucket.org/jatone/genieql/commands"

	"github.com/alecthomas/kingpin"
)

type bootstrapPackage struct {
	buildInfo
	definitionFileNames struct {
		TableStructures string
		Functions       string
		Scanners        string
		BatchInserts    string
		GoGenerate      string
	}
	importPaths []string
}

func (t *bootstrapPackage) Bootstrap(*kingpin.ParseContext) error {
	for _, importPath := range t.importPaths {
		var (
			err error
			pkg *build.Package
		)
		log.Println("importPath", importPath)
		if pkg, err = genieql.LocatePackage(importPath, build.Default, nil); err != nil {
			log.Println("failed to bootstrap package", importPath, err)
			continue
		}

		commands.WriteStdoutOrFile(
			printGenerator{delegate: pbootstrap.NewTableStructure(pkg)},
			t.definitionFileNames.TableStructures,
			commands.DefaultWriteFlags,
		)

		commands.WriteStdoutOrFile(
			printGenerator{delegate: pbootstrap.NewScanners(pkg)},
			t.definitionFileNames.Scanners,
			commands.DefaultWriteFlags,
		)

		commands.WriteStdoutOrFile(
			printGenerator{delegate: pbootstrap.NewFunctions(pkg)},
			t.definitionFileNames.Functions,
			commands.DefaultWriteFlags,
		)

		commands.WriteStdoutOrFile(
			printGenerator{delegate: pbootstrap.NewInsertBatch(pkg)},
			t.definitionFileNames.BatchInserts,
			commands.DefaultWriteFlags,
		)

		commands.WriteStdoutOrFile(
			printGenerator{delegate: pbootstrap.NewGoGenerateDefinitions(pkg)},
			t.definitionFileNames.GoGenerate,
			commands.DefaultWriteFlags,
		)
	}

	return nil
}

func (t *bootstrapPackage) configure(bootstrap *kingpin.CmdClause) *kingpin.CmdClause {
	bootstrap.Flag("tableStructureDefinitionsOutput", "filename for table structures definitions").
		Default("00_structs.table.genieql.go").StringVar(&t.definitionFileNames.TableStructures)
	bootstrap.Flag("scannerDefinitionsOutput", "filename for scanner definitions").
		Default("01_scanners.genieql.go").StringVar(&t.definitionFileNames.Scanners)
	bootstrap.Flag("functionDefinitionsOutput", "filename for functions definitions").
		Default("02_functions.genieql.go").StringVar(&t.definitionFileNames.Functions)
	bootstrap.Flag("batchInsertDefinitionsOutput", "filename for batch insert definitions").
		Default("03_insert.batch.genieql.go").StringVar(&t.definitionFileNames.BatchInserts)
	bootstrap.Flag("goGenerateOutput", "filename for the go generate file").
		Default("10_genieql.go").StringVar(&t.definitionFileNames.GoGenerate)
	bootstrap.Arg("package", "import paths where boilerplate configuration files will be generated").
		Default(t.CurrentPackageImport()).StringsVar(&t.importPaths)

	bootstrap.Action(t.Bootstrap)

	return bootstrap
}
