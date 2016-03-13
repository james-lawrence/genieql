package main

import "gopkg.in/alecthomas/kingpin.v2"

type scanners struct{}

func (t *scanners) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("scanner", "generate scanners")
	(&queryLiteral{}).configure(cmd)
	(&defaultScanner{}).configure(cmd)

	return cmd
}
