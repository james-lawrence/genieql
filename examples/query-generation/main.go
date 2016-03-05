// Package crud example generate a scanner and basic queries from a structure and a table.
// database setup instructions, replace database connection information as needed.
// USERNAME=postgres
// HOST=localhost
// PORT=5432
// pushd src/bitbucket.org/jatone/genieql/examples/query-generation
// createdb -p $PORT -U $USERNAME genieql_examples "genieql"
// cat structure.sql | psql -p $PORT -U $USERNAME -d genieql_examples
// genieql bootstrap postgres://$USERNAME@$HOST:$PORT/genieql_examples?sslmode=disable
// popd
// go generate bitbucket.org/jatone/genieql/examples/query-generation
package main

import "time"

//go:generate genieql map bitbucket.org/jatone/genieql/examples/query-generation.example snakecase lowercase
//go:generate genieql scanner default --output=example_default_scanner_gen.go bitbucket.org/jatone/genieql/examples/query-generation.example crud
//go:generate genieql generate crud --output=example_crud_queries_gen.go bitbucket.org/jatone/genieql/examples/query-generation.example crud
//go:generate genieql generate insert --output=example_insert_queries_gen.go bitbucket.org/jatone/genieql/examples/query-generation.example crud --suffix=WithDefaults --default=updated --default=created
type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

func main() {

}
