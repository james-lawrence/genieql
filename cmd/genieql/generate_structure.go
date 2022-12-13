package main

import (
	"go/ast"
	"go/build"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/compiler"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/stringsx"

	"github.com/alecthomas/kingpin"
	"github.com/serenize/snaker"
)

// GenerateStructure root command for generating structures.
type GenerateStructure struct {
	buildInfo
}

func (t *GenerateStructure) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	structure := cmd.Command("structure", "commands for generating structs from databases")
	tables := structure.Command("table", "commands for generating from tables")
	queries := structure.Command("query", "commands for generating from queries")
	(&GenerateTableCLI{
		buildInfo: t.buildInfo,
	}).configure(tables).Default()
	(&GenerateTableConstants{
		buildInfo: t.buildInfo,
	}).configure(tables)
	(&GenerateQueryConstants{
		buildInfo: t.buildInfo,
	}).configure(queries)

	return structure
}

// GenerateTableCLI creates a genieql mapping for the table specified from the command line.
type GenerateTableCLI struct {
	buildInfo
	table      string
	typeName   string
	output     string
	configName string
	pkg        string
}

func (t *GenerateTableCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cli := cmd.Command("cli", "generates a structure for the provided options and table").Action(t.execute)
	cli.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Flag("package", "package").Default(t.CurrentPackageImport()).StringVar(&t.pkg)
	cli.Arg("table", "name of the table to generate the mapping from").Required().StringVar(&t.table)

	return cli
}

func (t *GenerateTableCLI) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		columns []genieql.ColumnInfo
	)

	if ctx, err = loadGeneratorContext(build.Default, t.configName, t.pkg); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(ctx.Driver, t.table); err != nil {
		return err
	}

	pg := printGenerator{
		pkg: ctx.CurrentPackage,
		delegate: generators.NewStructure(
			generators.StructOptionContext(ctx),
			generators.StructOptionName(
				stringsx.DefaultIfBlank(t.typeName, snaker.SnakeToCamel(t.table)),
			),
			generators.StructOptionMappingConfigOptions(
				genieql.MCOColumns(columns...),
				genieql.MCOPackage(ctx.CurrentPackage),
			),
		),
	}

	return cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags)
}

// GenerateTableConstants creates a genieql mappings for the tables defined in the specified package.
type GenerateTableConstants struct {
	buildInfo
	typeName   string
	pkg        string
	configName string
	output     string
}

func (t *GenerateTableConstants) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	constants := cmd.Command("constants", "generates structures for the tables defined in the specified file").Action(t.execute)
	constants.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").Default(t.CurrentPackageImport()).StringVar(&t.pkg)

	return cmd
}

func (t *GenerateTableConstants) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx  generators.Context
		tags = []string{
			"genieql", "generate", "structure", "table",
		}
	)

	if ctx, err = loadGeneratorContext(build.Default, t.configName, t.pkg, tags...); err != nil {
		return err
	}

	taggedFiles, err := compiler.FindTaggedFiles(t.pkg, tags...)
	if err != nil {
		return err
	}

	if taggedFiles.Empty() {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}
	err = genieql.NewUtils(ctx.FileSet).WalkFiles(func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		consts := genieql.FindConstants(file)

		decls := mapDeclsToGenerator(func(decl *ast.GenDecl) []genieql.Generator {
			return generators.StructureFromGenDecl(
				decl,
				func(table string) generators.StructOption {
					return generators.StructOptionTableStrategy(table)
				},
				generators.StructOptionContext(ctx),
				generators.StructOptionMappingConfigOptions(
					genieql.MCOPackage(ctx.CurrentPackage),
				),
			)
		}, consts...)

		g = append(g, decls...)
	}, ctx.CurrentPackage)

	if err != nil {
		return err
	}

	mg := genieql.MultiGenerate(g...)
	hg := headerGenerator{
		fset: ctx.FileSet,
		pkg:  ctx.CurrentPackage,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      ctx.CurrentPackage,
		delegate: genieql.MultiGenerate(hg, mg),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

// GenerateQueryConstants generates structures from the defined constant within
// a file tagged with `//+build genieql,generate,structure,query`.
type GenerateQueryConstants struct {
	buildInfo
	typeName   string
	pkg        string
	configName string
	output     string
}

func (t *GenerateQueryConstants) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	constants := cmd.Command("constants", "generates structures for the queries defined in the specified file").Action(t.execute)
	constants.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").
		Default(t.CurrentPackageImport()).StringVar(&t.pkg)

	return cmd
}

func (t *GenerateQueryConstants) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx  generators.Context
		tags = []string{
			"genieql", "generate", "structure", "query",
		}
	)

	if ctx, err = loadGeneratorContext(build.Default, t.configName, t.pkg, tags...); err != nil {
		return err
	}

	taggedFiles, err := compiler.FindTaggedFiles(t.pkg, tags...)
	if err != nil {
		return err
	}

	if taggedFiles.Empty() {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}
	err = genieql.NewUtils(ctx.FileSet).WalkFiles(func(k string, f *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(k)) {
			return
		}

		decls := mapDeclsToGenerator(func(decl *ast.GenDecl) []genieql.Generator {
			return generators.StructureFromGenDecl(
				decl,
				func(query string) generators.StructOption {
					return generators.StructOptionQueryStrategy(query)
				},
				generators.StructOptionContext(ctx),
				generators.StructOptionMappingConfigOptions(
					genieql.MCOPackage(ctx.CurrentPackage),
				),
			)
		}, genieql.FindConstants(f)...)
		g = append(g, decls...)
	}, ctx.CurrentPackage)

	if err != nil {
		return err
	}

	hg := headerGenerator{
		fset: ctx.FileSet,
		pkg:  ctx.CurrentPackage,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      ctx.CurrentPackage,
		delegate: genieql.MultiGenerate(hg, genieql.MultiGenerate(g...)),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}
