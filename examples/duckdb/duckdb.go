package duckdb

//go:generate rm -rf duck.db
//go:generate genieql duckdb ../../.migrations/duckdb
// hack fix for ci/cd permissions issues.
//go:generate chmod 0770 duck.db
//go:generate genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/marcboeker/go-duckdb --output-file=duckdb.example.config duckdb://localhost/duck.db
//go:generate genieql auto --config "duckdb.example.config" -o "genieql.gen.go"
