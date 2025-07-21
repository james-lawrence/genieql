package main

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/goosex"
	"github.com/pressly/goose/v3"
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

func (t *duckdb) execute(*kingpin.ParseContext) (err error) {
	dbpath := filepath.Join(genieql.ConfigurationDirectory(), ".duckdb", t.database)
	if err = os.MkdirAll(filepath.Dir(dbpath), 0700); err != nil {
		return err
	}

	db, err := sql.Open("duckdb", dbpath)
	if err != nil {
		return err
	}
	defer db.Close()

	mprov, err := goose.NewProvider("", db, os.DirFS(t.migrations), goose.WithStore(goosex.DuckdbStore{}))
	if err != nil {
		return errorsx.Wrap(err, "unable to build migration provider")
	}

	if _, err := mprov.Up(context.Background()); err != nil {
		return errorsx.Wrap(err, "unable to run migrations")
	}

	return nil
}
