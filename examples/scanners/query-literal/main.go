// Package query-literal example generate a scanner from a query const and a structure.
//
// database setup instructions, replace database connection information as needed.
// 	USERNAME=postgres
// 	HOST=localhost
// 	PORT=5432
// 	pushd src/bitbucket.org/jatone/genieql/examples/scanners/query-literal
// 	createdb -p $PORT -U $USERNAME genieql_examples "genieql"
// 	cat structure.sql | psql -p $PORT -U $USERNAME -d genieql_examples
// 	genieql bootstrap postgres://$USERNAME@$HOST:$PORT/genieql_examples?sslmode=disable
// 	popd
// 	go generate bitbucket.org/jatone/genieql/examples/scanners/query-literal
package main

import "time"

//go:generate genieql map bitbucket.org/jatone/genieql/examples/scanners/query-literal.example snakecase lowercase
//go:generate genieql scanner query-literal --output=example_query_literal_gen.go CrudScanner bitbucket.org/jatone/genieql/examples/scanners/query-literal.example bitbucket.org/jatone/genieql/examples/scanners/query-literal.myQuery
type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

const myQuery = `SELECT * FROM query_literal`

func main() {

}