package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/cmd"
	"bitbucket.org/jatone/genieql/crud"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"
)

type generateInsertConfig struct {
	buildInfo
	configName string
	output     string
	mapName    string
	table      string
	defaults   []string
}

type generateInsert struct {
	generateInsertConfig
}

func (t *generateInsert) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	insert := cmd.Command("insert", "generate more complicated insert queries")
	insert.Flag(
		"config",
		"name of configuration file to use",
	).Default("default.config").StringVar(&t.generateInsertConfig.configName)

	insert.Flag(
		"mapping",
		"name of the map to use",
	).Default("default").StringVar(&t.generateInsertConfig.mapName)

	insert.Flag("default", "specifies a name of a column to default to database value").
		StringsVar(&t.defaults)

	insert.Flag(
		"output",
		"path of output file",
	).Short('o').Default("").StringVar(&t.generateInsertConfig.output)

	cmd = (&insertQueryCmd{
		generateInsertConfig: &t.generateInsertConfig,
	}).configure(insert.Command("constant", "output the query as a constant")).Default()

	x := insert.Command("experimental", "experimental insert commands")
	cmd = (&insertBatchCmd{
		generateInsertConfig: &t.generateInsertConfig,
	}).configure(x.Command("batch-function", "generate a batch insert function"))

	cmd = (&insertFunctionCmd{
		generateInsertConfig: &t.generateInsertConfig,
	}).configure(x.Command("function", "output the insert as a function"))

	return insert
}

type insertBatchCmd struct {
	*generateInsertConfig
	pkg                string
	emitColumnConstant bool
	batch              int
}

func (t *insertBatchCmd) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cmd.Action(t.execute)
	cmd.Flag("column-constant", "controls column constant being output").Default("true").BoolVar(&t.emitColumnConstant)
	cmd.Flag("batch", "number of records to insert").Default("100").IntVar(&t.batch)
	cmd.Arg(
		"package",
		"package to search for definitions",
	).Default(t.CurrentPackageImport()).StringVar(&t.pkg)

	return cmd
}

func (t *insertBatchCmd) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx  generators.Context
		tags = []string{
			"genieql", "generate", "insert", "batch",
		}
	)

	if ctx, err = loadGeneratorContext(build.Default, t.configName, t.pkg, tags...); err != nil {
		return err
	}

	taggedFiles, err := findTaggedFiles(t.pkg, tags...)
	if err != nil {
		return err
	}

	if len(taggedFiles.files) == 0 {
		log.Println("no files tagged, ignoring")
		// nothing to do.
		return nil
	}

	g := []genieql.Generator{}
	genieql.NewUtils(ctx.FileSet).WalkFiles(func(path string, file *ast.File) {
		if !taggedFiles.IsTagged(filepath.Base(path)) {
			return
		}

		functionsTypes := mapDeclsToGenerator(func(d *ast.GenDecl) []genieql.Generator {
			var (
				ok       bool
				defaults []string
				table    string
			)
			options, _ := generators.ParseCommentOptions(d.Doc)

			if table, ok = generators.CommentOptionTable(options); !ok {
				return []genieql.Generator{genieql.NewErrGenerator(errors.New("table is required for batch insert"))}
			}

			defaults, _ = generators.CommentOptionDefaultColumns(options)
			builder := func(local string, n int, columns ...string) ast.Decl {
				return genieql.QueryLiteral(
					local,
					ctx.Dialect.Insert(n, table, columns, defaults),
				)
			}

			return generators.NewBatchFunctionFromGenDecl(
				ctx,
				d,
				builder,
				defaults,
			)
		}, genieql.SelectFuncType(genieql.FindTypes(file)...)...)

		g = append(g, functionsTypes...)
	}, ctx.CurrentPackage)

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
		return errors.Wrap(err, "failed to write results")
	}

	return nil
}

type insertQueryCmd struct {
	*generateInsertConfig
	constSuffix         string
	emitColumnConstant  bool
	emitExplodeFunction bool
	batch               int
	packageType         string
}

func (t *insertQueryCmd) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cmd.Action(t.execute)

	cmd.Flag(
		"suffix",
		"suffix for the name of the generated constant",
	).Required().StringVar(&t.constSuffix)

	cmd.Flag("column-constant", "controls column constant being output").Default("true").BoolVar(&t.emitColumnConstant)
	cmd.Flag("explode-function", "controls explode function being output").Default("true").BoolVar(&t.emitExplodeFunction)

	cmd.Flag("batch", "number of records to insert").Default("1").IntVar(&t.batch)

	cmd.Arg(
		"package.Type",
		"package prefixed structure we want to build the scanner/query for",
	).Required().StringVar(&t.packageType)

	cmd.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	return cmd
}

func (t *insertQueryCmd) execute(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
		fields  []*ast.Field
	)

	pkgRelativePath, typName := t.extractPackageType(t.packageType)
	if ctx, err = loadGeneratorContext(build.Default, t.configName, pkgRelativePath); err != nil {
		return err
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	if err = ctx.Configuration.ReadMap(t.mapName, &mapping, genieql.MCOColumns(columns...), genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return err
	}

	if columns, _, err = mapping.MappedColumnInfo(ctx.Driver, ctx.Dialect, ctx.FileSet, ctx.CurrentPackage); err != nil {
		return err
	}

	if fields, _, err = mapping.MapFieldsToColumns2(ctx.FileSet, ctx.CurrentPackage, genieql.ColumnInfoSet(columns).Filter(genieql.ColumnInfoFilterIgnore(t.defaults...))...); err != nil {
		return errors.Wrapf(err, "failed to map fields to columns for: %s", t.packageType)
	}

	details := genieql.TableDetails{Columns: columns, Dialect: ctx.Dialect, Table: t.table}
	constName := fmt.Sprintf("%sInsert%s", typName, t.constSuffix)

	hg := newHeaderGenerator(t.buildInfo, ctx.FileSet, t.packageType, os.Args[1:]...)
	cc := maybeColumnConstants(t.emitColumnConstant, fmt.Sprintf("%sStaticColumns", constName), ctx.Dialect, t.defaults, columns)

	ef := maybeGenerator(
		t.emitExplodeFunction,
		generators.NewExploderFunction(
			ctx,
			astutil.Field(ast.NewIdent(typName), ast.NewIdent("arg1")),
			fields,
			generators.QFOName(fmt.Sprintf("%sExplode", constName)),
		),
	)
	cg := crud.Insert(details).Build(t.batch, constName, t.defaults)
	pg := printGenerator{
		pkg:      ctx.CurrentPackage,
		delegate: genieql.MultiGenerate(hg, cc, ef, cg),
	}

	if err = cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags); err != nil {
		log.Fatalln(err)
	}

	return nil
}

type insertFunctionCmd struct {
	*generateInsertConfig
	functionName        string
	scanner             string
	packageType         string
	queryer             string
	emitColumnConstant  bool
	emitExplodeFunction bool
}

func (t *insertFunctionCmd) configure(cmd *kingpin.CmdClause) *kingpin.CmdClause {
	cmd.Action(t.functionCmd)
	cmd.Flag(
		"function-name",
		"override the name of the function, by default it is the {Type}Insert or {Type}InsertWithDefaults",
	).StringVar(&t.functionName)

	cmd.Flag("column-constant", "controls column constant being output").Default("true").BoolVar(&t.emitColumnConstant)
	cmd.Flag("explode-function", "controls explode function being output").Default("true").BoolVar(&t.emitExplodeFunction)

	cmd.Flag("scanner", "name of the scanner to use using the package.Type format").StringVar(&t.scanner)
	cmd.Flag("queryer", "selector expression representing the type that will execute the queryer").StringVar(&t.queryer)
	cmd.Arg(
		"package.Type",
		"the type we want to build the scanner/query for, can optionally be prefixed with a package",
	).Required().StringVar(&t.packageType)

	cmd.Arg(
		"table",
		"table you want to build the queries for",
	).Required().StringVar(&t.table)

	return cmd
}

func (t *insertFunctionCmd) functionCmd(*kingpin.ParseContext) (err error) {
	var (
		ctx     generators.Context
		mapping genieql.MappingConfig
		columns []genieql.ColumnInfo
		fields  []*ast.Field
		queryer ast.Expr
	)

	pkgRelativePath, typName := t.extractPackageType(t.packageType)
	if ctx, err = loadGeneratorContext(build.Default, t.configName, pkgRelativePath); err != nil {
		return err
	}

	if err = ctx.Configuration.ReadMap(t.mapName, &mapping, genieql.MCOPackage(ctx.CurrentPackage), genieql.MCOType(typName)); err != nil {
		return errors.Wrap(err, "failed to load mapping context")
	}

	if columns, err = ctx.Dialect.ColumnInformationForTable(t.table); err != nil {
		return err
	}

	if queryer, err = parser.ParseExpr(stringsx.DefaultIfBlank(t.queryer, ctx.Configuration.Queryer)); err != nil {
		return errors.Wrapf(err, "%s: is not a valid expression", t.queryer)
	}

	if columns, _, err = mapping.Clone(genieql.MCOColumns(columns...)).MappedColumnInfo(ctx.Driver, ctx.Dialect, ctx.FileSet, ctx.CurrentPackage); err != nil {
		return err
	}

	onlyMap := genieql.ColumnInfoSet(columns).Filter(genieql.ColumnInfoFilterIgnore(t.defaults...))
	if fields, _, err = mapping.MapFieldsToColumns2(ctx.FileSet, ctx.CurrentPackage, onlyMap...); err != nil {
		return errors.Wrapf(err, "failed to locate fields for %s", t.packageType)
	}

	scannerName := fmt.Sprintf("New%sScannerStaticRow", stringsx.ToPublic(typName))
	scannerName = stringsx.DefaultIfBlank(t.scanner, scannerName)

	functionName := fmt.Sprintf("%sInsert", typName)
	if len(t.defaults) > 0 {
		functionName = fmt.Sprintf("%sInsertWithDefaults", typName)
	}
	functionName = stringsx.DefaultIfBlank(t.functionName, functionName)

	searcher := genieql.NewSearcher(ctx.FileSet, ctx.CurrentPackage)
	scannerFunction, err := searcher.FindFunction(func(name string) bool {
		return name == scannerName
	})
	if err != nil {
		return errors.Wrapf(err, "failed to find the scanner: %s", scannerName)
	}

	field := astutil.Field(ast.NewIdent(typName), ast.NewIdent("arg1"))

	hg := newHeaderGenerator(t.buildInfo, ctx.FileSet, t.packageType, os.Args[1:]...)
	cc := maybeColumnConstants(t.emitColumnConstant, fmt.Sprintf("%sStaticColumns", functionName), ctx.Dialect, t.defaults, columns)
	ef := maybeGenerator(
		t.emitExplodeFunction,
		generators.NewExploderFunction(
			ctx,
			astutil.Field(ast.NewIdent(typName), ast.NewIdent("arg1")),
			fields,
			generators.QFOName(fmt.Sprintf("%sExplode", functionName)),
		),
	)
	cg := generators.NewQueryFunction(
		ctx,
		generators.QFOName(functionName),
		generators.QFOBuiltinQueryFromString(ctx.Dialect.Insert(1, t.table, genieql.ColumnInfoSet(columns).ColumnNames(), t.defaults)),
		generators.QFOScanner(scannerFunction),
		generators.QFOExplodeStructParam(field, fields...),
		generators.QFOQueryer("q", queryer),
	)

	pg := printGenerator{
		pkg:      ctx.CurrentPackage,
		delegate: genieql.MultiGenerate(hg, cc, ef, cg),
	}

	return cmd.WriteStdoutOrFile(pg, t.output, cmd.DefaultWriteFlags)
}

func maybeColumnConstants(enabled bool, name string, dialect genieql.Dialect, defaults []string, columns []genieql.ColumnInfo) genieql.Generator {
	return maybeGenerator(enabled, generators.NewColumnConstants(
		name,
		genieql.ColumnValueTransformer{
			Defaults:           defaults,
			DialectTransformer: dialect.ColumnValueTransformer(),
		},
		columns,
	),
	)
}

func maybeGenerator(enabled bool, g genieql.Generator) genieql.Generator {
	if enabled {
		return g
	}
	return genieql.NewErrGenerator(nil)
}
