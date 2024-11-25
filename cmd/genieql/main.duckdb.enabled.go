//go:build genieql.duckdb

package main

import (

	// ensure the driver is registered
	"context"
	"database/sql"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/goosex"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/pressly/goose/v3"
)

func (t *duckdb) execute(*kingpin.ParseContext) (err error) {
	db, err := sql.Open("duckdb", t.database)
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
