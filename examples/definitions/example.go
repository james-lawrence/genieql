package definitions

//go:generate dropdb --if-exists -U postgres genieql_examples
//go:generate createdb -U postgres genieql_examples "genieql example database"
//go:generate psql -U postgres -d genieql_examples --file=structure.sql
//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://postgres@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql generate experimental structure table constants -o postgresql.table.structs.gen.go
//go:generate genieql generate experimental structure query constants -o postgresql.query.structs.gen.go
//go:generate genieql generate experimental scanners types -o postgresql.scanners.gen.go
//go:generate genieql generate experimental crud -o postgresql.crud.functions.gen.go --table=example1 --scanner=DynamicExample1Scanner --unique-scanner=NewStaticRowExample1Scanner bitbucket.org/jatone/genieql/examples/definitions.Example1
//go:generate genieql generate experimental functions types -o postgresql.functions.gen.go
