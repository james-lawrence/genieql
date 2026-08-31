package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
	extensions []string
}

func (t *duckdb) configure(app *kingpin.Application) *kingpin.CmdClause {
	cli := app.Command("duckdb", "duckdb migrations using goose").Action(t.execute)
	cli.Flag("database", "name of the database file to create").Default("duck.db").StringVar(&t.database)
	cli.Flag("extension", "duckdb extension to install and load prior to migrations").StringsVar(&t.extensions)
	cli.Arg("migrations", "path to the migrations directory").Required().StringVar(&t.migrations)
	return cli
}

// duckdbExtensionState reports whether the named extension is installed and/or
// loaded in the given database, using duckdb's own extension introspection.
func duckdbExtensionState(db *sql.DB, ext string) (installed bool, loaded bool, err error) {
	row := db.QueryRowContext(
		context.Background(),
		"SELECT installed, loaded FROM duckdb_extensions() WHERE extension_name = ?",
		ext,
	)

	if err := row.Scan(&installed, &loaded); err != nil {
		if err == sql.ErrNoRows {
			return false, false, nil
		}
		return false, false, err
	}

	return installed, loaded, nil
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

	for _, ext := range t.extensions {
		installed, loaded, err := duckdbExtensionState(db, ext)
		if err != nil {
			return errorsx.Wrapf(err, "failed to determine state of '%s' extension", ext)
		}

		if loaded {
			log.Println("extension already loaded", ext)
			continue
		}

		if !installed {
			log.Println("installing extension", ext)
			if _, err := db.ExecContext(context.Background(), fmt.Sprintf("INSTALL %s;", ext)); err != nil {
				// DuckDB may fail to reach the extension repository even though the
				// extension is already present locally from a prior install; only
				// treat this as fatal when the extension truly isn't installed.
				stillInstalled, _, serr := duckdbExtensionState(db, ext)
				if serr != nil || !stillInstalled {
					return errorsx.Wrapf(err, "failed to install '%s' extension", ext)
				}
				log.Println("warning: failed to refresh extension, using existing install", ext, err)
			}
		}

		log.Println("loading extension", ext)
		if _, err := db.ExecContext(context.Background(), fmt.Sprintf("LOAD %s;", ext)); err != nil {
			return errorsx.Wrapf(err, "failed to load '%s' extension", ext)
		}
	}

	mprov, err := goose.NewProvider("", db, os.DirFS(t.migrations), goose.WithStore(goosex.DuckdbStore{}))
	if err != nil {
		return errorsx.Wrap(err, "unable to build migration provider")
	}

	if _, err := mprov.Up(context.Background()); err != nil {
		return errorsx.Wrap(err, "unable to run migrations")
	}

	return nil
}
