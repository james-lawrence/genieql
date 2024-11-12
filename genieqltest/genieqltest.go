package genieqltest

import (
	"go/ast"
	"go/build"
	"go/token"
	"log"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/buildx"
	"bitbucket.org/jatone/genieql/columninfo"
	"bitbucket.org/jatone/genieql/dialects"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"
	"bitbucket.org/jatone/genieql/internal/errorsx"

	_ "bitbucket.org/jatone/genieql/internal/postgresql"
)

func NewColumnMap(d genieql.Driver, typ string, local string, field string) genieql.ColumnMap {
	return genieql.ColumnMap{
		ColumnInfo: genieql.ColumnInfo{
			Definition: errorsx.Must(d.LookupType(typ)),
			Name:       field,
		},
		Dst:   astutil.SelExpr(local, field),
		Field: astutil.Field(ast.NewIdent(typ), ast.NewIdent(field)),
	}
}

func DialectConfig1(options ...genieql.ConfigurationOption) genieql.Configuration {
	const dialect = "test.dialect.1"
	err := dialects.Register(dialect, dialects.TestFactory(dialects.Test{
		Quote:             "\"",
		CValueTransformer: columninfo.NewNameTransformer(),
		QueryInsert:       "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES :gql.insert.values::gql.insert.conflict: RETURNING :gql.insert.returning:",
	}))
	if err != nil {
		log.Println("failed to register test dialect", dialect, err)
	}

	return genieql.MustConfiguration(
		genieql.Configuration{
			Dialect: dialect,
			Driver:  drivers.StandardLib,
		}.Clone(options...),
	)
}

func DialectPSQL(options ...genieql.ConfigurationOption) genieql.Configuration {
	return genieql.MustConfiguration(
		genieql.Configuration{
			Dialect: "postgres",
			Driver:  drivers.PGX,
		}.Clone(options...),
	)
}

func GeneratorContext(c genieql.Configuration) (ctx generators.Context, err error) {
	var (
		driver  genieql.Driver
		dialect genieql.Dialect
	)

	if driver, err = genieql.LookupDriver(c.Driver); err != nil {
		return ctx, err
	}

	if dialect, err = dialects.LookupDialect(c); err != nil {
		return ctx, err
	}

	pkg := &build.Package{
		Name: "example",
		Dir:  filepath.Dir(c.Location),
		GoFiles: []string{
			"example.go",
		},
	}

	bctx := buildx.Clone(
		build.Default,
		buildx.Tags(genieql.BuildTagIgnore, genieql.BuildTagGenerate),
	)

	return generators.Context{
		Build:          bctx,
		Configuration:  c,
		CurrentPackage: pkg,
		FileSet:        token.NewFileSet(),
		Dialect:        dialect,
		Driver:         driver,
	}, nil
}
