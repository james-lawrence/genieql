// Package crud Create, Read, Update, Delete example
// setup: see README.md
package main

import "time"

//go:generate genieql map --natural-key=id bitbucket.org/jatone/genieql/examples/crud.example crud snakecase lowercase
//go:generate genieql generate crud --output=example_crud_gen.go bitbucket.org/jatone/genieql/examples/crud.example crud

type example struct {
	ID      int
	Email   string
	Created time.Time
	Updated time.Time
}

func main() {

}
