package main

import (
	"log"
	"net/url"
	"path/filepath"

	"github.com/alecthomas/kingpin"

	"bitbucket.org/jatone/genieql"
)

// data stored in qlgenie.conf - dialect, default alias strategy, map definition directory,
// default for including table prefix aliases, database connection information.
// qlgenie bootstrap database postgres://username:password@host:port/example?sslmode=disabled -> creates default.config
// qlgenie bootstrap database --ouput="someothername.qlgenie" postgres://username:password@host:port/example?sslmode=disabled -> creates someothername.qlgenie
type bootstrapDatabase struct {
	outputfilepath string
	outputfile     string
	dburi          *url.URL
	driver         string
	queryer        string
	rowtype        string
}

func (t *bootstrapDatabase) Bootstrap(ctx *kingpin.ParseContext) error {
	log.Println("bootstraping", t.dburi)
	return genieql.Bootstrap(
		genieql.ConfigurationOptionLocation(filepath.Join(t.outputfilepath, t.outputfile)),
		genieql.ConfigurationOptionDriver(t.driver),
		genieql.ConfigurationOptionDatabase(t.dburi),
		genieql.ConfigurationOptionQueryer(t.queryer),
		genieql.ConfigurationOptionRowType(t.rowtype),
	)
}

func (t *bootstrapDatabase) configure(bootstrap *kingpin.CmdClause) *kingpin.CmdClause {
	bootstrap.Flag("output-directory", "directory to place the configuration file").Default(genieql.ConfigurationDirectory()).StringVar(&t.outputfilepath)
	bootstrap.Flag("output-file", "filename of the configuration directory").Default("default.config").StringVar(&t.outputfile)
	bootstrap.Flag("driver", "name of the underlying driver for the database, usually the import url").
		Default("github.com/lib/pq").StringVar(&t.driver)
	bootstrap.Flag("queryer", "the default queryer to use").Default("*sql.DB").StringVar(&t.queryer)
	bootstrap.Flag("rowtype", "the default type to use for retrieving rows").Default("*sql.Row").StringVar(&t.rowtype)
	bootstrap.Arg("uri", "uri for the database qlgenie will work with").Required().URLVar(&t.dburi)
	bootstrap.Action(t.Bootstrap)

	return bootstrap
}
