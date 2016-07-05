// Package query example generate a scanner from a query const and a structure.
//
// database setup instructions, replace database connection information as needed.
// 	USERNAME=postgres
// 	HOST=localhost
// 	PORT=5432
// 	pushd src/bitbucket.org/jatone/genieql/examples/scanners/query-literal
// 	createdb -p $PORT -U $USERNAME genieql_examples "genieql"
// 	cat structure.sql | psql -p $PORT -U $USERNAME -d genieql_examples
// 	popd
// 	go generate bitbucket.org/jatone/genieql/examples/scanners/query
package query

import "time"

//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://$USER@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql map bitbucket.org/jatone/genieql/examples/scanners/query.example snakecase lowercase
//go:generate genieql scanner default --interface-only --output=example_scanner.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example crud
//go:generate genieql scanner query-literal --output=example_query_literal.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example bitbucket.org/jatone/genieql/examples/scanners/query.query
type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

const query = `SELECT * FROM query_literal`
