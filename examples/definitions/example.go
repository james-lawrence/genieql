package definitions

//go:generate dropdb --if-exists -U postgres genieql_examples
//go:generate createdb -U postgres genieql_examples "genieql example database"
//go:generate psql -U postgres -d genieql_examples --file=structure.sql
//go:generate genieql bootstrap --driver=github.com/lib/pq postgres://postgres@localhost:5432/genieql_examples?sslmode=disable
//go:generate genieql generate experimental structure table constants -o postgresql.table.structs.gen.go
//go:generate genieql generate experimental structure query constants -o postgresql.query.structs.gen.go
//go:generate genieql generate experimental scanners types -o postgresql.scanners.gen.go
//go:generate genieql generate experimental crud -o postgresql.crud.functions.gen.go --table=example1 --scanner=NewExample1ScannerDynamic --unique-scanner=NewExample1ScannerStaticRow bitbucket.org/jatone/genieql/examples/definitions.Example1
//go:generate genieql generate experimental functions types -o postgresql.functions.gen.go

const query1 = `SELECT * FROM example1 WHERE id = $1 || id = $2 || id = $3`
