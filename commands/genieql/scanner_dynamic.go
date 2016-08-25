package main

import (
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/scanner"
)

type dynamicScanner struct {
	scanner scannerConfig
	table   string
}

func (t dynamicScanner) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sDynamicScanner"),
		defaultRowScannerNameFormat("%sDynamicRowScanner"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRowScanner"),
		defaultErrScannerNameFormat("%sErrScanner"),
	}
}

func (t *dynamicScanner) Execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		mappingConfig genieql.MappingConfig
		fset          = token.NewFileSet()
	)

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.scanner.configName),
		),
	)
	if err = genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	pkgName, typName := extractPackageType(t.scanner.packageType)

	if err = genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	if err = genieql.ReadMapper(configuration, pkgName, typName, t.scanner.mapName, &mappingConfig); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(pkgName, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	details, err := genieql.LoadInformation(configuration, t.table)
	if err != nil {
		log.Fatalln(err)
	}

	fields, err := mappingConfig.TypeFields(fset, build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	gen := scanner.DynamicScannerGenerator{
		Generator: scanner.Generator{
			Mappings: []genieql.MappingConfig{mappingConfig},
			Fields:   fields,
			Columns:  details.Columns,
			Driver:   genieql.MustLookupDriver(configuration.Driver),
		},
		ScannerName:      t.scanner.scannerName,
		InterfaceName:    t.scanner.interfaceName,
		InterfaceRowName: t.scanner.interfaceRowName,
		ErrScannerName:   t.scanner.errScannerName,
	}

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, gen),
	}

	if err = commands.WriteStdoutOrFile(pg, t.scanner.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func (t *dynamicScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(cmd, t.Options()...)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)
	return cmd.Action(t.Execute)
}
