package drivers

import (
	"bitbucket.org/jatone/genieql"
)

// implements the lib/pq driver https://github.com/lib/pq
func init() {
	genieql.RegisterDriver(PQ, NewDriver("github.com/jackc/pgtype", pgxexports(), pgx...))
}

// PQ - driver for github.com/lib/pq
const PQ = "github.com/lib/pq"

// TODO
// var libpq = []genieql.ColumnDefinition{}
