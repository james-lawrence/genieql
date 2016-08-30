// Package basic example generate a scanner and basic queries from a structure and a table.
// database setup instructions, replace database connection information as needed.
package basic

import "time"

//go:generate dropdb --if-exists -U postgres genieql_examples
//go:generate createdb -U postgres genieql_examples "genieql example database"
//go:generate psql -U postgres -d genieql_examples --file=structure.sql
//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://$USER@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql map bitbucket.org/jatone/genieql/examples/basic.example camelcase
//go:generate genieql scanner default --output=example_default_scanner.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud
//go:generate genieql scanner dynamic --output=example_dynamic_scanner.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud
//go:generate genieql generate crud --output=example_crud_queries.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud
//go:generate genieql generate insert --output=example_insert_queries.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud --suffix=WithDefaults --default=updated --default=created
type example struct {
	ID      int
	Email   *string
	Created time.Time
	Updated time.Time
}
