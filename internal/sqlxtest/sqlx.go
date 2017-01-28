package sqlxtest

import (
	"database/sql"
	"fmt"

	"github.com/satori/go.uuid"

	// load the postgresql driver
	_ "github.com/lib/pq"
)

// TemplateDatabaseName template database name
const TemplateDatabaseName string = "genieql_test_template"

const dbtemplate string = "dbname=%s sslmode=disable port=5432"

func generatePostgresql(name, template string) string {
	return fmt.Sprintf("CREATE DATABASE \"%s\" TEMPLATE %s", name, template)
}

func destroyPostgresql(name string) string {
	return fmt.Sprintf("DROP DATABASE \"%s\"", name)
}

func NewPostgresql(template string) (string, *sql.DB) {
	name := uuid.NewV4().String()
	psql := mustOpen(sql.Open("postgres", fmt.Sprintf(dbtemplate, "postgres")))
	defer psql.Close()
	mustExec(psql.Exec(generatePostgresql(name, template)))
	return name, mustOpen(sql.Open("postgres", fmt.Sprintf(dbtemplate, name)))
}

func DestroyPostgresql(template, name string) {
	psql := mustOpen(sql.Open("postgres", fmt.Sprintf(dbtemplate, "postgres")))
	defer psql.Close()
	mustExec(psql.Exec(destroyPostgresql(name)))
}

func mustOpen(db *sql.DB, err error) *sql.DB {
	if err != nil {
		panic(err)
	}

	return db
}

func mustExec(result sql.Result, err error) sql.Result {
	if err != nil {
		panic(err)
	}

	return result
}
