package main

import (
	"net/url"
	"path/filepath"

	"bitbucket.org/jatone/genieql"

	"gopkg.in/alecthomas/kingpin.v2"
)

// data stored in qlgenie.conf - dialect, default alias strategy, map definition directory,
// default for including table prefix aliases, database connection information.
// qlgenie bootstrap psql://host:port/example?username=123&password=456 -> creates example.qlgenie.
// qlgenie bootstrap --ouput="someothername.qlgenie" psql://host:port/example?username=123&password=456 -> creates someothername.qlgenie
type bootstrap struct {
	outputfilepath string
	dburi          *url.URL
}

func (t bootstrap) Bootstrap() error {
	return genieql.Bootstrap(t.outputfilepath, t.dburi)
}

func (t *bootstrap) configure(app *kingpin.Application) *kingpin.CmdClause {
	t.outputfilepath = filepath.Join(configurationDirectory(), "default.config")

	bootstrap := app.Command("bootstrap", "build a instance of qlgenie")
	bootstrap.Arg("uri", "uri for the database qlgenie will work with").Required().URLVar(&t.dburi)
	bootstrap.Flag("output-directory", "directory to place the configuration file").Default(t.outputfilepath).StringVar(&t.outputfilepath)

	return bootstrap
}
