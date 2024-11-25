package main

import (
	"github.com/alecthomas/kingpin"
)

type duckdb struct {
	database   string
	migrations string
}

func (t *duckdb) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("duckdb", "duckdb migrations using goose").Action(t.execute)
	cli.Flag("database", "name of the database file to create").Default("duck.db").StringVar(&t.database)
	cli.Arg("migrations", "path to the migrations directory").Required().StringVar(&t.migrations)
	return cli
}
