package genieql_test

import (
	"go/build"
	"go/token"
	"io"
	"log"
	"path/filepath"
	"testing"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/columninfo"
	"bitbucket.org/jatone/genieql/dialects"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/internal/drivers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterp(t *testing.T) {
	log.SetOutput(io.Discard)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interp Suite")
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

	return generators.Context{
		Configuration:  c,
		CurrentPackage: pkg,
		FileSet:        token.NewFileSet(),
		Dialect:        dialect,
		Driver:         driver,
	}, nil
}

func DialectConfig1() genieql.Configuration {
	const dialect = "test.dialect.1"
	err := dialects.Register(dialect, dialects.TestFactory(dialects.Test{
		Quote:             "\"",
		CValueTransformer: columninfo.NewNameTransformer(),
		QueryInsert:       "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES :gql.insert.values::gql.insert.conflict:",
	}))
	if err != nil {
		log.Println("failed to register test dialect", dialect, err)
	}
	return genieql.Configuration{
		Location: ".fixtures/.genieql",
		Dialect:  dialect,
		Driver:   drivers.StandardLib,
	}
}
