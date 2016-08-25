// Package query example generate a scanner from a query const and a structure.
package query

import "time"

//go:generate dropdb --if-exists -U postgres genieql_examples
//go:generate createdb -U postgres genieql_examples "genieql example database"
//go:generate psql -U postgres -d genieql_examples --file=structure.sql
//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://postgres@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql map bitbucket.org/jatone/genieql/examples/scanners/query.example snakecase lowercase
//go:generate genieql scanner default --interface-only --output=example_scanner.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example query_literal
//go:generate genieql scanner query-literal --output=example_query_literal.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example bitbucket.org/jatone/genieql/examples/scanners/query.query

type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

const query = `SELECT id,created,updated FROM query_literal`
