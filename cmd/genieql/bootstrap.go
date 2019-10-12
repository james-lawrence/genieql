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

	(&bootstrapPackage{
		buildInfo: t.buildInfo,
	}).configure(
		bootstrap.Command("package", "generate the boilerplate for each package provided"),
	)

	return bootstrap
}
