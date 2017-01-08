package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/commands"
	"bitbucket.org/jatone/genieql/crud"
	"bitbucket.org/jatone/genieql/generators"
)

type generateInsert struct {
	configName         string
	constSuffix        string
	packageType        string
	table              string
	output             string
	mapName            string
	pkg                string
	batch              int
	defaults           []string
	emitColumnConstant bool
}

func (t *generateInsert) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	insert := cmd.Command("insert", "generate more complicated insert queries")
	insert.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.configName)

	insert.Flag(
		"mapping",
		"name of the map to use",
	).Default("default").StringVar(&t.mapName)

	insert.Flag("default", "specifies a name of a column to default to database value").
		StringsVar(&t.defaults)

	insert.Flag(
		"output",
		"path of output file",
	).Default("").StringVar(&t.output)

	cmd = insert.Command("constant", "output the query as a constant").Action(t.constant).Default()
	cmd.Flag(
		"suffix",
		"suffix for the name of the generated constant",
	).Required().StringVar(&t.constSuffix)

	cmd.Flag("column-constant", "controls column constant being output").Default("true").BoolVar(&t.emitColumnConstant)

	cmd.Flag("batch", "number of records to insert").Default("1").IntVar(&t.batch)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	cmd.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	x := insert.Command("experimental", "experimental insert commands")
	cmd = x.Command("batch-function", "generate a batch insert function").Action(t.batchCmd)
	cmd.Flag("column-constant", "controls column constant being output").Default("true").BoolVar(&t.emitColumnConstant)
	cmd.Flag("batch", "number of records to insert").Default("100").IntVar(&t.batch)
	cmd.Arg(
		"package",
		"package to search for definitions",
	).Required().StringVar(&t.packageType)

	return insert
}

func (t *generateInsert) batchCmd(*kingpin.ParseContext) error {
	var (
		err     error
		config  genieql.Configuration
		dialect genieql.Dialect
		pkg     *build.Package
		fset    = token.NewFileSet()
	)

	if config, dialect, pkg, err = loadPackageContext(t.configName, t.packageType, fset); err != nil {
		return err
	}

	ctx := generators.Context{
		CurrentPackage: pkg,
		FileSet:        fset,
		Configuration:  config,
		Dialect:        dialect,
	}

	taggedFiles, err := findTaggedFiles(t.packageType, "genieql", "generate", "insert", "batch")
	if err != nil {
		log.Fatalln(err)
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged, ignoring")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}
	genieql.NewUtils(fset).WalkFiles(func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		functionsTypes := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			return generators.NewBatchFunctionFromGenDecl(
				ctx,
				d,
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, functionsTypes...)
	}, pkg)
	// pkgName, typName := extractPackageType(t.packageType)
	// if _, dialect, mapping, err = loadMappingContext(t.configName, pkgName, typName, t.mapName); err != nil {
	// 	log.Println("WAAT")
	// 	return err
	// }
	//
	// mapping.CustomQuery = false
	// mapping.TableOrQuery = t.table
	//
	// if columns, _, err = mapping.MappedColumnInfo(fset, build.Default, genieql.StrictPackageName(filepath.Base(pkgName))); err != nil {
	// 	log.Println("failed to load mapped column info")
	// 	return err
	// }

	// functionName := fmt.Sprintf("%sBatch%dInsert%s", typName, t.batch, t.constSuffix)
	// field := astutil.Field(
	// 	astutil.SelExpr(filepath.Base(pkgName), typName),
	// 	ast.NewIdent("in"),
	// )
	// builder := func(n int) ast.Decl {
	// 	return genieql.QueryLiteral(
	// 		"query",
	// 		dialect.Insert(n, t.table, genieql.ColumnInfoSet(columns).ColumnNames(), t.defaults),
	// 	)
	// }
	//
	// exampleScanner := &ast.FuncDecl{
	// 	Name: ast.NewIdent("StaticExampleScanner"),
	// 	Type: &ast.FuncType{
	// 		Params: &ast.FieldList{
	// 			List: []*ast.Field{
	// 				astutil.Field(astutil.Expr("*sql.Rows"), ast.NewIdent("rows")),
	// 				astutil.Field(astutil.Expr("error"), ast.NewIdent("err")),
	// 			},
	// 		},
	// 		Results: &ast.FieldList{
	// 			List: []*ast.Field{astutil.Field(ast.NewIdent("ExampleScanner"))},
	// 		},
	// 	},
	// }
	hg := headerGenerator{
		fset: fset,
		pkg:  pkg,
		args: os.Args[1:],
	}
	// cc := maybeColumnConstants(t.emitColumnConstant, fmt.Sprintf("%sStaticColumns", functionName), dialect, t.defaults, columns)
	// bg := generators.NewBatchFunction(t.batch, field,
	// 	generators.BatchFunctionQueryBuilder(builder),
	// 	generators.BatchFunctionQFOptions(
	// 		generators.QFOName(functionName),
	// 		generators.QFOScanner(exampleScanner),
	// 		generators.QFOQueryer("q", astutil.SelExpr("sqlx", "Queryer")),
	// 		generators.QFOQueryerFunction(ast.NewIdent("Query")),
	// 	),
	// )
	// pg := printGenerator{
	// 	delegate: genieql.MultiGenerate(hg, cc, bg),
	// }
	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, genieql.MultiGenerate(g...)),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Println("failed to write results")
		log.Fatalln(err)
	}

	return nil
}

func (t *generateInsert) constant(*kingpin.ParseContext) error {
	var (
		err     error
		mapping genieql.MappingConfig
		dialect genieql.Dialect
		columns []genieql.ColumnInfo
		pkg     *build.Package
		fset    = token.NewFileSet()
	)

	pkgName, typName := extractPackageType(t.packageType)
	if _, dialect, mapping, err = loadMappingContext(t.configName, pkgName, typName, t.mapName); err != nil {
		return err
	}
	if pkg, err = locatePackage(pkgName); err != nil {
		return err
	}

	mapping.CustomQuery = false
	mapping.TableOrQuery = t.table

	if columns, _, err = mapping.MappedColumnInfo2(dialect, fset, pkg); err != nil {
		return err
	}

	details := genieql.TableDetails{Columns: columns, Dialect: dialect, Table: t.table}
	constName := fmt.Sprintf("%sInsert%s", typName, t.constSuffix)

	hg := newHeaderGenerator(fset, t.packageType, os.Args[1:]...)
	cc := maybeColumnConstants(t.emitColumnConstant, fmt.Sprintf("%sStaticColumns", constName), dialect, t.defaults, columns)
	cg := crud.Insert(details).Build(t.batch, constName, t.defaults)
	pg := printGenerator{
		delegate: genieql.MultiGenerate(hg, cc, cg),
	}

	if err = commands.WriteStdoutOrFile(pg, t.output, commands.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

func maybeColumnConstants(enabled bool, name string, dialect genieql.Dialect, defaults []string, columns []genieql.ColumnInfo) genieql.Generator {
	cc := genieql.NewErrGenerator(nil)
	if enabled {
		cc = generators.NewColumnConstants(
			name,
			genieql.ColumnValueTransformer{
				Defaults:           defaults,
				DialectTransformer: dialect.ColumnValueTransformer(),
			},
			columns,
		)
	}
	return cc
}
