package main

import (
	"go/build"
	"log"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/scanner"
)

type staticScanner struct {
	scanner scannerConfig
	table   string
}

func (t staticScanner) Options() []scannerOption {
	return []scannerOption{
		defaultScannerNameFormat("%sStaticScanner"),
		defaultRowScannerNameFormat("%sStaticRowScanner"),
		defaultInterfaceNameFormat("%sScanner"),
		defaultInterfaceRowNameFormat("%sRowScanner"),
		defaultErrScannerNameFormat("%sErrScanner"),
	}
}

func (t *staticScanner) Execute(*kingpin.ParseContext) error {
	var configuration genieql.Configuration
	var mappingConfig genieql.MappingConfig

	pkgName, typName := extractPackageType(t.scanner.packageType)

	if err := genieql.ReadConfiguration(filepath.Join(configurationDirectory(), t.scanner.configName), &configuration); err != nil {
		log.Fatalln(err)
	}

	if err := genieql.ReadMapper(configurationDirectory(), pkgName, typName, t.scanner.mapName, configuration, &mappingConfig); err != nil {
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

	fields, err := mappingConfig.TypeFields(build.Default, genieql.StrictPackageName(filepath.Base(pkgName)))
	if err != nil {
		log.Fatalln(err)
	}

	generator := scanner.StaticScanner{
		Generator: scanner.Generator{
			Mappings: []genieql.MappingConfig{mappingConfig},
			Fields:   fields,
			Columns:  details.Columns,
			Driver:   genieql.MustLookupDriver(configuration.Driver),
		},
		ScannerName:      t.scanner.scannerName,
		RowScannerName:   t.scanner.scannerRowName,
		InterfaceName:    t.scanner.interfaceName,
		InterfaceRowName: t.scanner.interfaceRowName,
		ErrScannerName:   t.scanner.errScannerName,
	}

	printScanner(t.scanner.output, generator, pkg)

	return nil
}

func (t *staticScanner) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	(&t.scanner).configure(cmd, t.Options()...)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want a scanner for",
	).Required().StringVar(&t.scanner.packageType)
	cmd.Arg("table", "name of the table to build the scanner for").Required().StringVar(&t.table)
	return cmd.Action(t.Execute)
}
