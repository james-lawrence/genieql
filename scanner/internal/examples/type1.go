package examples

import "time"

//go:generate dropdb --if-exists -U postgres genieql_test_template
//go:generate createdb -U postgres genieql_test_template
//go:generate psql -1 -f structure.sql genieql_test_template
//go:generate genieql bootstrap --driver=github.com/lib/pq --output-file=scanner-test.config postgres://$USER@localhost:5432/genieql_test_template?sslmode=disable
//go:generate genieql map --config=scanner-test.config bitbucket.org/jatone/genieql/scanner/internal/examples.Type1
//go:generate genieql scanner default --config=scanner-test.config --output=type1_scanner.gen.go bitbucket.org/jatone/genieql/scanner/internal/examples.Type1 type1
//go:generate genieql scanner dynamic --config=scanner-test.config --output=type1_dynamic_scanner.gen.go bitbucket.org/jatone/genieql/scanner/internal/examples.Type1 type1
//go:generate genieql generate crud --config=scanner-test.config --output=type1_queries.gen.go bitbucket.org/jatone/genieql/scanner/internal/examples.Type1 type1

// Type1 just a type for testing
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
