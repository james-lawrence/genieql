// Package basic example generate a scanner and basic queries from a structure and a table.
// database setup instructions, replace database connection information as needed.
// USERNAME=postgres
// HOST=localhost
// PORT=5432
// pushd src/bitbucket.org/jatone/genieql/examples/query-generation
// createdb -p $PORT -U $USERNAME genieql_examples "genieql"
// cat structure.sql | psql -p $PORT -U $USERNAME -d genieql_examples
// popd
// go generate bitbucket.org/jatone/genieql/examples/basic
package basic

import "time"

//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://$USER@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql map bitbucket.org/jatone/genieql/examples/basic.example snakecase lowercase
//go:generate genieql scanner default --output=example_default_scanner.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud
//go:generate genieql generate crud --output=example_crud_queries.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud
//go:generate genieql generate insert --output=example_insert_queries.gen.go bitbucket.org/jatone/genieql/examples/basic.example crud --suffix=WithDefaults --default=updated --default=created
type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}
