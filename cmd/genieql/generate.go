package main

import (
	"go/build"
	"go/token"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/generators"
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

func loadGeneratorContext(bctx build.Context, name, pkg string, tags ...string) (ctx generators.Context, err error) {
	var (
		config  genieql.Configuration
		dialect genieql.Dialect
		driver  genieql.Driver
		bpkg    *build.Package
	)

	bctx.BuildTags = tags

	config = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), name),
		),
	)

	if dialect, err = genieql.LookupDialect(config); err != nil {
		return ctx, err
	}

	if driver, err = genieql.LookupDriver(config.Driver); err != nil {
		return ctx, err
	}

	if bpkg, err = genieql.LocatePackage(pkg, bctx, genieql.StrictPackageImport(pkg)); err != nil {
		return ctx, err
	}

	return generators.Context{
		Build:          bctx,
		CurrentPackage: bpkg,
		FileSet:        token.NewFileSet(),
		Configuration:  config,
		Dialect:        dialect,
		Driver:         driver,
	}, err
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

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		return configuration, dialect, err
	}

	return configuration, dialect, err
}

func loadMappingContext(config string, pkg *build.Package, typ, mName string) (genieql.Configuration, genieql.Dialect, genieql.MappingConfig, error) {
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

	if err = configuration.ReadMap(mName, &mapping, genieql.MCOPackage(pkg), genieql.MCOType(typ)); err != nil {
		return configuration, dialect, mapping, err
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		return configuration, dialect, mapping, err
	}

	return configuration, dialect, mapping, err
}
