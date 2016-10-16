package main

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/x/stringsx"

	"github.com/serenize/snaker"
	"gopkg.in/alecthomas/kingpin.v2"
)

// GenerateStructure root command for generating structures.
type GenerateStructure struct{}

func (t *GenerateStructure) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	structure := cmd.Command("structure", "commands for generating structs from databases")
	tables := structure.Command("table", "commands for generating from tables")
	queries := structure.Command("query", "commands for generating from queries")
	(&GenerateTableCLI{}).configure(tables).Default()
	(&GenerateTableConstants{}).configure(tables)
	(&GenerateQueryConstants{}).configure(queries)

	return structure
}

// GenerateTableCLI creates a genieql mapping for the table specified from the command line.
type GenerateTableCLI struct {
	table      string
	typeName   string
	output     string
	configName string
}

func (t *GenerateTableCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cli := cmd.Command("cli", "generates a structure for the provided options and table").Action(t.execute)
	cli.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Arg("table", "name of the table to generate the mapping from").Required().StringVar(&t.table)

	return cli
}

func (t *GenerateTableCLI) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		wd            string
		pkg           *build.Package
	)

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg = currentPackage(wd); pkg == nil {
		log.Fatalln("error no package found in", wd)
	}

	if err = genieql.ReadConfiguration(&configuration); err != nil {
		return err
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		log.Fatalln(err)
	}

	pg := printGenerator{
		delegate: generators.NewStructure(
			generators.StructOptionConfiguration(configuration),
			generators.StructOptionName(
				stringsx.DefaultIfBlank(t.typeName, snaker.SnakeToCamel(t.table)),
			),
			generators.StructOptionMappingConfigOptions(
				genieql.MCOCustom(false),
				genieql.MCOColumnInfo(t.table),
				genieql.MCODialect(dialect),
				genieql.MCOPackage(pkg.ImportPath),
			),
		),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

// GenerateTableConstants creates a genieql mappings for the tables defined in the specified package.
type GenerateTableConstants struct {
	table      string
	typeName   string
	pkg        string
	configName string
	output     string
}

func (t *GenerateTableConstants) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	var (
		err error
		wd  string
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg != nil {
		t.pkg = pkg.ImportPath
	}

	constants := cmd.Command("constants", "generates structures for the tables defined in the specified file").Action(t.execute)
	constants.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").StringVar(&t.pkg)

	return cmd
}

func (t *GenerateTableConstants) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		fset          = token.NewFileSet()
	)
	configuration = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(t.pkg, build.Default, genieql.StrictPackageName(filepath.Base(t.pkg)))
	if err != nil {
		log.Fatalln(err)
	}

	taggedFiles, err := findTaggedFiles(t.pkg, "genieql", "generate", "structure", "table")
	if err != nil {
		log.Fatalln(err)
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}

	genieql.NewUtils(fset).WalkFiles([]*build.Package{pkg}, func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}
		consts := genieql.FindConstants(file)

		decls := mapDeclsToGenerator(func(decl *ast.GenDecl) []genieql.Generator {
			return generators.StructureFromGenDecl(
				decl,
				generators.StructOptionConfiguration(configuration),
				generators.StructOptionMappingConfigOptions(
					genieql.MCOCustom(false),
					genieql.MCODialect(dialect),
					genieql.MCOPackage(pkg.ImportPath),
				),
			)
		}, consts...)

		g = append(g, decls...)
	})

	mg := genieql.MultiGenerate(g...)
	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, mg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

// GenerateQueryCLI creates a struct from the query specified from the command line.
type GenerateQueryCLI struct {
	query      string
	typeName   string
	output     string
	configName string
}

func (t *GenerateQueryCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cli := cmd.Command("cli", "generates a structure for the provided options and query").Action(t.execute)
	cli.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Arg("query", "query to generate the mapping from").Required().StringVar(&t.query)

	return cli
}

func (t *GenerateQueryCLI) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		wd            string
		pkg           *build.Package
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg = currentPackage(wd); pkg == nil {
		log.Fatalln("error no package found in", wd)
	}

	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	if err = genieql.ReadConfiguration(&configuration); err != nil {
		return err
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		log.Fatalln(err)
	}

	pg := printGenerator{
		delegate: generators.NewStructure(
			generators.StructOptionConfiguration(configuration),
			generators.StructOptionName(
				stringsx.DefaultIfBlank(t.typeName, snaker.SnakeToCamel(t.query)),
			),
			generators.StructOptionMappingConfigOptions(
				genieql.MCOCustom(true),
				genieql.MCODialect(dialect),
				genieql.MCOPackage(pkg.ImportPath),
			),
		),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

// GenerateQueryConstants generates structures from the defined constant within
// a file tagged with `//+build genieql,generate,structure,query`.
type GenerateQueryConstants struct {
	table      string
	typeName   string
	pkg        string
	configName string
	output     string
}

func (t *GenerateQueryConstants) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	var (
		err error
		wd  string
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg != nil {
		t.pkg = pkg.ImportPath
	}

	constants := cmd.Command("constants", "generates structures for the queries defined in the specified file").Action(t.execute)
	constants.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").StringVar(&t.pkg)

	return cmd
}

func (t *GenerateQueryConstants) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		fset          = token.NewFileSet()
	)
	configuration = genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)

	if err = genieql.ReadConfiguration(&configuration); err != nil {
		log.Fatalln(err)
	}

	if dialect, err = genieql.LookupDialect(configuration); err != nil {
		log.Fatalln(err)
	}

	pkg, err := genieql.LocatePackage(t.pkg, build.Default, genieql.StrictPackageName(filepath.Base(t.pkg)))
	if err != nil {
		log.Fatalln(err)
	}

	taggedFiles, err := findTaggedFiles(t.pkg, "genieql", "generate", "structure", "query")
	if err != nil {
		log.Fatalln(err)
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}

	genieql.NewUtils(fset).WalkFiles([]*build.Package{pkg}, func(k string, f *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(k)) {
			return
		}

		decls := mapDeclsToGenerator(func(decl *ast.GenDecl) []genieql.Generator {
			return generators.StructureFromGenDecl(
				decl,
				generators.StructOptionConfiguration(configuration),
				generators.StructOptionMappingConfigOptions(
					genieql.MCOCustom(true),
					genieql.MCODialect(dialect),
					genieql.MCOPackage(pkg.ImportPath),
				),
			)
		}, genieql.FindConstants(f)...)
		g = append(g, decls...)
	})

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, genieql.MultiGenerate(g...)),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}
