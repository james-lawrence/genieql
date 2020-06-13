package scanners

import "time"

//go:generate genieql map --config=generators-test.config Type1
//go:generate genieql generate experimental scanners types --config=generators-test.config -o postgresql.scanners.gen.go
//go:generate genieql scanner default --config=generators-test.config --output=type1_scanner.gen.go Type1 type1
//go:generate genieql scanner dynamic --config=generators-test.config --output=type1_dynamic_scanner.gen.go Type1 type1
//go:generate genieql generate crud --config=generators-test.config --output=type1_queries.gen.go Type1 type1

// Type1 for testing
type Type1 struct {
	Field1 string
	Field2 *string
	Field3 bool
	Field4 *bool
	Field5 int
	Field6 *int
	Field7 time.Time
	Field8 *time.Time
}
