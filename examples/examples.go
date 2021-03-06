package examples

//go:generate dropdb --if-exists -U postgres genieql_examples
//go:generate createdb -U postgres genieql_examples "genieql example database"
//go:generate psql -X -U postgres -d genieql_examples --file=structure.sql
//go:generate genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/lib/pq postgres://postgres@localhost:5432/genieql_examples?sslmode=disable
