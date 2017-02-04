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

	"github.com/alecthomas/kingpin"
	"github.com/serenize/snaker"
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
	pkg        string
}

func (t *GenerateTableCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	var (
		err error
		wd  string
		pkg *build.Package
	)

	cli := cmd.Command("cli", "generates a structure for the provided options and table").Action(t.execute)
	cli.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Flag("package", "package").StringVar(&t.pkg)
	cli.Arg("table", "name of the table to generate the mapping from").Required().StringVar(&t.table)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg = currentPackage(wd); pkg == nil {
		t.pkg = pkg.ImportPath
	}

	return cli
}

func (t *GenerateTableCLI) execute(*kingpin.ParseContext) error {
	var (
		err           error
		columns       []genieql.ColumnInfo
		configuration genieql.Configuration
		dialect       genieql.Dialect
		pkg           *build.Package
		fset          = token.NewFileSet()
	)

	if configuration, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
	}

	if columns, err = dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  configuration,
		Dialect:        dialect,
	}
	pg := printGenerator{
		pkg: pkg,
		delegate: generators.NewStructure(
			generators.StructOptionContext(ctx),
			generators.StructOptionName(
				stringsx.DefaultIfBlank(t.typeName, snaker.SnakeToCamel(t.table)),
			),
			generators.StructOptionMappingConfigOptions(
				genieql.MCOColumns(columns...),
				genieql.MCOPackage(pkg.ImportPath),
			),
		),
	}

	return commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags)
}

// GenerateTableConstants creates a genieql mappings for the tables defined in the specified package.
type GenerateTableConstants struct {
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

	constants := cmd.Command("constants", "generates structures for the tables defined in the specified file").Action(t.execute)
	constants.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").StringVar(&t.pkg)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg != nil {
		t.pkg = pkg.ImportPath
	}

	return cmd
}

func (t *GenerateTableConstants) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		pkg           *build.Package
		fset          = token.NewFileSet()
	)
	if configuration, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
	}

	taggedFiles, err := findTaggedFiles(t.pkg, "genieql", "generate", "structure", "table")
	if err != nil {
		return err
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged")
		// nothing to do.
		return nil
	}

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  configuration,
		Dialect:        dialect,
	}

	g := []genieql.Generator{}
	genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
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
					genieql.MCOPackage(pkg.ImportPath),
				),
			)
		}, consts...)

		g = append(g, decls...)
	}, pkg)

	mg := genieql.MultiGenerate(g...)
	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      pkg,
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
	pkg        string
}

func (t *GenerateQueryCLI) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	var (
		err error
		wd  string
	)

	cli := cmd.Command("cli", "generates a structure for the provided options and query").Action(t.execute)
	cli.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	cli.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	cli.Flag("output", "output filename").Short('o').StringVar(&t.output)
	cli.Flag("package", "package to look for the type within").StringVar(&t.pkg)
	cli.Arg("query", "query to generate the mapping from").Required().StringVar(&t.query)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg == nil {
		t.pkg = pkg.ImportPath
	}

	return cli
}

func (t *GenerateQueryCLI) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		pkg           *build.Package
		fset          = token.NewFileSet()
	)
	if configuration, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
	}

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  configuration,
		Dialect:        dialect,
	}
	pg := printGenerator{
		pkg: pkg,
		delegate: generators.NewStructure(
			generators.StructOptionContext(ctx),
			generators.StructOptionName(t.typeName),
			generators.StructOptionMappingConfigOptions(
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
		err     error
		wd      string
		pkgPath string
	)

	if wd, err = os.Getwd(); err != nil {
		log.Fatalln(err)
	}

	if pkg := currentPackage(wd); pkg != nil {
		pkgPath = pkg.ImportPath
	}

	constants := cmd.Command("constants", "generates structures for the queries defined in the specified file").Action(t.execute)
	constants.Flag("configName", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	constants.Flag("name", "name of the type to generate").StringVar(&t.typeName)
	constants.Flag("output", "output filename").Short('o').StringVar(&t.output)
	constants.Arg("package", "package to search for constant definitions").Default(pkgPath).StringVar(&t.pkg)

	return cmd
}

func (t *GenerateQueryConstants) execute(*kingpin.ParseContext) error {
	var (
		err           error
		configuration genieql.Configuration
		dialect       genieql.Dialect
		pkg           *build.Package
		fset          = token.NewFileSet()
	)
	if configuration, dialect, pkg, err = loadPackageContext(t.configName, t.pkg, fset); err != nil {
		return err
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

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  configuration,
		Dialect:        dialect,
	}
	g := []genieql.Generator{}
	genieql.NewUtils(fset).WalkFiles(func(k string, f *ast.File) {
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
					genieql.MCOPackage(pkg.ImportPath),
				),
			)
		}, genieql.FindConstants(f)...)
		g = append(g, decls...)
	}, pkg)

	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}

	pg := printGenerator{
		pkg:      pkg,
		delegate: genieql.MultiGenerate(hg, genieql.MultiGenerate(g...)),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}
