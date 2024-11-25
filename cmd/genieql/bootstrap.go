package main

import (
	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql"
)

type bootstrap struct {
	genieql.BuildInfo
}

func (t *bootstrap) configure(app *kingpin.Application) *kingpin.CmdClause {
	bootstrap := app.Command("bootstrap", "commands for bootstrapping configurations")

	(&bootstrapDatabase{}).configure(
		bootstrap.Command("database", "build a instance of genieql from the provided database"),
	).Default()

	(&bootstrapPackage{
		BuildInfo: t.BuildInfo,
	}).configure(
		bootstrap.Command("package", "generate the boilerplate for each package provided"),
	)

	return bootstrap
}
