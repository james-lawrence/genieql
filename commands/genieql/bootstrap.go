package main

import "github.com/alecthomas/kingpin"

type bootstrap struct {
	buildInfo
}

func (t *bootstrap) configure(app *kingpin.Application) *kingpin.CmdClause {
	bootstrap := app.Command("bootstrap", "commands for bootstrapping configurations")

	(&bootstrapDatabase{}).configure(
		bootstrap.Command("database", "build a instance of qlgenie from the provided database"),
	).Default()

	return bootstrap
}
