package sqlxtest

import (
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	uuid "github.com/satori/go.uuid"
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
	name := uuid.NewV4().String()
	psql := mustOpen(fmt.Sprintf(dbtemplate, "postgres"))
	defer psql.Close()
	mustExec(psql.Exec(generatePostgresql(name, template)))
	return name, mustOpen(fmt.Sprintf(dbtemplate, name))
}

func DestroyPostgresql(template, name string) {
	psql := mustOpen(fmt.Sprintf(dbtemplate, "postgres"))
	defer psql.Close()
	mustExec(psql.Exec(destroyPostgresql(name)))
}

func mustOpen(cstring string) *sql.DB {
	pcfg := mustParse(pgx.ParseConfig(cstring))
	return stdlib.OpenDB(*pcfg)
}

func mustExec(result sql.Result, err error) sql.Result {
	if err != nil {
		panic(err)
	}

	return result
}

func mustParse(c *pgx.ConnConfig, err error) *pgx.ConnConfig {
	if err != nil {
		panic(err)
	}

	return c
}
