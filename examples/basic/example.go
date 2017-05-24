// Package basic example generate a scanner and basic queries from a structure and a table.
// database setup instructions, replace database connection information as needed.
package basic

import "time"

//go:generate genieql map example camelcase
//go:generate genieql scanner default --output=example_default_scanner.gen.go example example3
//go:generate genieql scanner dynamic --output=example_dynamic_scanner.gen.go example example3
//go:generate genieql generate crud --output=example_crud_queries.gen.go example example3
//go:generate genieql generate insert --output=example_insert_queries.gen.go example example3 --suffix=WithDefaults --default=updated --default=created
//go:generate genieql generate experimental crud --output=example_crud_functions.gen.go --table=example3 --queryer-type=sqlx.Queryer --unique-scanner=NewExampleScannerStaticRow --scanner=NewExampleScannerStatic example

type example struct {
	ID      int
	Email   *string
	Created time.Time
	Updated time.Time
}
