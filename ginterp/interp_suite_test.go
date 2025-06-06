package ginterp_test

import (
	"log"
	"testing"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/columninfo"
	"github.com/james-lawrence/genieql/dialects"
	"github.com/james-lawrence/genieql/internal/drivers"
	"github.com/james-lawrence/genieql/internal/testx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInterp(t *testing.T) {
	testx.Logging()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Interp Suite")
}

func DialectConfig1() genieql.Configuration {
	const dialect = "test.dialect.1"
	err := dialects.Register(dialect, dialects.TestFactory(dialects.Test{
		Quote:             "\"",
		CValueTransformer: columninfo.NewNameTransformer(),
		QueryInsert:       "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES :gql.insert.values::gql.insert.conflict: RETURNING :gql.insert.returning:",
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
