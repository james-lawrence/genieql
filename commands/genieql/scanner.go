package main

import "github.com/alecthomas/kingpin"

type scanners struct {
	buildInfo
}

func (t *scanners) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("scanner", "generate scanners")
	(&queryLiteral{
		scanner: scannerConfig{
			buildInfo: t.buildInfo,
		},
	}).configure(cmd.Command("query-literal", "build a scanner for the provided type/query"))
	(&staticScanner{
		scanner: scannerConfig{
			buildInfo: t.buildInfo,
		},
	}).configure(cmd.Command("static", "build a static scanner for the provided type/table"))
	(&dynamicScanner{
		scanner: scannerConfig{
			buildInfo: t.buildInfo,
		},
	}).configure(cmd.Command("dynamic", "build a dynamic scanner for the provided type/table"))
	(&defaultScanner{
		scanner: scannerConfig{
			buildInfo: t.buildInfo,
		},
	}).configure(cmd.Command("default", "build the default (which is a static scanner) for the provided type/table"))
	return cmd
}
