package sqlxtest

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/pressly/goose/v3"
)

// TemplateDatabaseName template database name
const TemplateDatabaseName string = "genieql_test_template"

const dbtemplate string = "dbname=%s sslmode=disable"

func generatePostgresql(name, template string) string {
	return fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE %s", name, template)
}

func destroyPostgresql(name string) string {
	return fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", name)
}

func NewPostgresql(template string) (string, *sql.DB) {
	name := uuid.Must(uuid.NewV4()).String()
	psql := mustOpen(fmt.Sprintf(dbtemplate, "postgres"))
	defer psql.Close()
	errorsx.Must(psql.Exec(generatePostgresql(name, template)))
	return name, mustOpen(fmt.Sprintf(dbtemplate, name))
}

func DestroyPostgresql(template, name string) {
	psql := mustOpen(fmt.Sprintf(dbtemplate, "postgres"))
	defer psql.Close()
	errorsx.Must(psql.Exec(destroyPostgresql(name)))
}

func mustOpen(cstring string) *sql.DB {
	pcfg := errorsx.Must(pgx.ParseConfig(cstring))
	return stdlib.OpenDB(*pcfg)
}

func generateDuckDB(name, template string) error {
	src, err := os.Open(template)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(name)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func Migrate(ctx context.Context, db *sql.DB, migrations fs.FS, options ...goose.ProviderOption) error {
	mprov, err := goose.NewProvider("", db, migrations, options...)
	if err != nil {
		return errorsx.Wrap(err, "unable to build migration provider")
	}

	if _, err := mprov.Up(ctx); err != nil {
		return errorsx.Wrap(err, "unable to run migrations")
	}

	return nil
}

func NewDuckDB() *sql.DB {
	return errorsx.Must(sql.Open("duckdb", ""))
}
