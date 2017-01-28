package example

import "database/sql"

type ExampleScanner interface{}
type ExampleRowScanner interface{}

func StaticExampleScanner(rows *sql.Rows, err error) ExampleScanner {
	return struct{}{}
}

func StaticExampleRowScanner(row *sql.Row) ExampleRowScanner {
	return struct{}{}
}

type StructA struct {
	A, B, C int
	D, E, F bool
}

type StructB struct {
	A, B, C int
	D, E, F bool
}
