package main

import (
	"go/build"
	"go/token"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"github.com/alecthomas/kingpin"
)

type generate struct {
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")
	x := cmd.Command("experimental", "experimental generation commands")

	(&generateCrud{}).configure(cmd)
	(&generateInsert{}).configure(cmd)
	(&GenerateStructure{}).configure(x)
	(&GenerateScanner{}).configure(x)
	(&generateCRUDFunctions{}).configure(x)
	(&generateFunctionTypes{}).configure(x)

	return cmd
}

func loadPackageContext(configName, pkg string, fset *token.FileSet) (genieql.Configuration, genieql.Dialect, *build.Package, error) {
	var (
		err     error
		config  genieql.Configuration
		dialect genieql.Dialect
		bpkg    *build.Package
	)

	config = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), configName),
		),
	)

	if dialect, err = genieql.LookupDialect(config); err != nil {
		return config, dialect, bpkg, err
	}

	if bpkg, err = locatePackage(pkg); err != nil {
		return config, dialect, bpkg, err
	}

	return config, dialect, bpkg, err
}

func loadMappingContext(config, pkg, typ, mName string) (genieql.Configuration, genieql.Dialect, genieql.MappingConfig, error) {
	var (
		err           error
		configuration genieql.Configuration
		mapping       genieql.MappingConfig
		dialect       genieql.Dialect
	)

	configuration = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), config),
		),
	)

	if err = configuration.ReadMap(pkg, typ, mName, &mapping); err != nil {
		return configuration, dialect, mapping, err
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		return configuration, dialect, mapping, err
	}

	return configuration, dialect, mapping, err
}
