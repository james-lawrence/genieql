// Package query example generate a scanner from a query const and a structure.
package query

import "time"

//go:generate genieql map bitbucket.org/jatone/genieql/examples/scanners/query.example snakecase lowercase
//go:generate genieql scanner default --interface-only --output=example_scanner.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example example3
//go:generate genieql scanner query-literal --output=example_example3.gen.go bitbucket.org/jatone/genieql/examples/scanners/query.example bitbucket.org/jatone/genieql/examples/scanners/query.query

type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

const query = `SELECT id,created,updated FROM example3`
