package duckdb

//go:generate rm -rf duck.db
//go:generate genieql duckdb ../../.migrations/duckdb
//go:generate genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/marcboeker/go-duckdb --output-file=duckdb.example.config duckdb://localhost/duck.db
//go:generate genieql auto --config "duckdb.example.config" -o "genieql.gen.go"
