package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"bitbucket.org/jatone/genieql"
	_ "github.com/lib/pq"
)

type database struct {
	Driver         string
	User           string
	Name           string
	SSLMode        string
	Port           int
	MaxConnections int
}

func main() {
	dbconf := database{
		Driver:         "postgres",
		User:           "jatone",
		Name:           "sso",
		SSLMode:        "disable",
		Port:           5432,
		MaxConnections: 10,
	}

	db, err := sql.Open(dbconf.Driver, fmt.Sprintf("user=%s dbname=%s sslmode=%s port=%d", dbconf.User, dbconf.Name, dbconf.SSLMode, dbconf.Port))
	if err != nil {
		log.Fatal(err)
	}

	columns, err := genieql.ExtractColumns(db, "SELECT * FROM identity")
	if err != nil {
		log.Fatalln(err)
	}

	if err := genieql.AmbiguityCheck(columns...); err != nil {
		log.Fatalln(err)
	}

	maps := []genieql.Mapping{m}
	matches := genieql.Matcher{maps}.Match(columns...)
	builder := genieql.ScannerBuilder{
		Name:      "Identity",
		Arguments: maps,
		Matches:   matches,
	}

	if err := builder.WriteScanner(os.Stdout); err != nil {
		log.Println(err)
		return
	}
}

var m = genieql.Mapping{
	Package: "sso",
	Type:    "Identity",
	Fields: []genieql.Field{
		{
			Type:        "string",
			StructField: "ID",
			Aliases: []string{
				"id",
				"identity_id",
			},
		},
		{
			Type:        "string",
			StructField: "Email",
			Aliases: []string{
				"email",
				"identity_email",
			},
		},
		{
			Type:        "time.Time",
			StructField: "Created",
			Aliases: []string{
				"created",
				"identity_created",
			},
		},
	},
}
