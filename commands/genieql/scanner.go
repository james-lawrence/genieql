package main

import "github.com/alecthomas/kingpin"

type scanners struct{}

func (t *scanners) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("scanner", "generate scanners")
	(&queryLiteral{}).configure(cmd.Command("query-literal", "build a scanner for the provided type/query"))
	(&staticScanner{}).configure(cmd.Command("static", "build a static scanner for the provided type/table"))
	(&dynamicScanner{}).configure(cmd.Command("dynamic", "build a dynamic scanner for the provided type/table"))
	(&defaultScanner{}).configure(cmd.Command("default", "build the default (which is a static scanner) for the provided type/table"))
	return cmd
}
