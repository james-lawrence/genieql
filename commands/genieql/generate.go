package main

import "gopkg.in/alecthomas/kingpin.v2"

type generate struct {
}

func (t *generate) configure(app *kingpin.Application) *kingpin.CmdClause {
	cmd := app.Command("generate", "generate sql queries")
	x := cmd.Command("experimental", "experimental generation commands")
	(&generateCrud{}).configure(cmd)
	(&generateInsert{}).configure(cmd)
	(&GenerateStructure{}).configure(x)
	(&GenerateScanner{}).configure(x)
	(&generateCRUDFunctions{}).configure(x)
	(&generateFunctionTypes{}).configure(x)

	return cmd
}
