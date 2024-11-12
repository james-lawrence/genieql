package sqlxtest

import (
	"database/sql"
	"fmt"

	"bitbucket.org/jatone/genieql/internal/errorsx"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
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
