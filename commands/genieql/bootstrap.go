package main

import (
	"log"
	"net/url"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"bitbucket.org/jatone/genieql"
)

// data stored in qlgenie.conf - dialect, default alias strategy, map definition directory,
// default for including table prefix aliases, database connection information.
// qlgenie bootstrap postgres://username:password@host:port/example?sslmode=disabled -> creates default.config
// qlgenie bootstrap --ouput="someothername.qlgenie" postgres://username:password@host:port/example?sslmode=disabled -> creates someothername.qlgenie
type bootstrap struct {
	outputfilepath string
	outputfile     string
	dburi          *url.URL
	driver         string
}

func (t bootstrap) Bootstrap() error {
	log.Println("bootstraping", t.dburi)
	return genieql.Bootstrap(filepath.Join(t.outputfilepath, t.outputfile), t.driver, t.dburi)
}

func (t *bootstrap) configure(app *kingpin.Application) *kingpin.CmdClause {
	bootstrap := app.Command("bootstrap", "build a instance of qlgenie")
	bootstrap.Flag("output-directory", "directory to place the configuration file").Default(configurationDirectory()).StringVar(&t.outputfilepath)
	bootstrap.Flag("output-file", "filename of the configuration directory").Default("default.config").StringVar(&t.outputfile)
	bootstrap.Flag("driver", "name of the underlying driver for the database, usually the import url i.e.) github.com/lib/pq").Required().StringVar(&t.driver)
	bootstrap.Arg("uri", "uri for the database qlgenie will work with").Required().URLVar(&t.dburi)

	return bootstrap
}
