package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type generate struct {
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")

	(&generateCrud{}).configure(cmd)
	(&generateInsert{}).configure(cmd)

	return cmd
}
