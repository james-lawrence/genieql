package main

import (
	"go/build"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/dialects"
	"github.com/alecthomas/kingpin"
)

type generate struct {
	buildInfo
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")
	x := cmd.Command("experimental", "experimental generation commands")
	gic := generateInsertConfig{
		buildInfo: t.buildInfo,
	}
	(&generateCrud{
		buildInfo: t.buildInfo,
	}).configure(cmd)
	(&generateInsert{
		generateInsertConfig: gic,
	}).configure(cmd)
	(&GenerateStructure{
		buildInfo: t.buildInfo,
	}).configure(x)
	(&GenerateScanner{
		buildInfo: t.buildInfo,
	}).configure(x)
	(&generateCRUDFunctions{
		buildInfo: t.buildInfo,
	}).configure(x)
	(&generateFunctionTypes{
		buildInfo: t.buildInfo,
	}).configure(x)

	return cmd
}

func loadContext(config string) (genieql.Configuration, genieql.Dialect, error) {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
	)

	configuration = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), config),
		),
	)

	if dialect, err = dialects.LookupDialect(configuration); err != nil {
		return configuration, dialect, err
	}

	return configuration, dialect, err
}

func loadMappingContext(config string, pkg *build.Package, typ string) (genieql.Configuration, genieql.Dialect, genieql.MappingConfig, error) {
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

	if err = configuration.ReadMap(&mapping, genieql.MCOPackage(pkg), genieql.MCOType(typ)); err != nil {
		return configuration, dialect, mapping, err
	}

	if dialect, err = dialects.LookupDialect(configuration); err != nil {
		return configuration, dialect, mapping, err
	}

	return configuration, dialect, mapping, err
}
