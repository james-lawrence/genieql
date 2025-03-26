//go:build genieql.duckdb

package main

import (

	// ensure the driver is registered
	"context"
	"database/sql"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/goosex"
	_ "github.com/marcboeker/go-duckdb"
	"github.com/pressly/goose/v3"
)

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
