package drivers

import (
	"bitbucket.org/jatone/genieql"
)

// implements the lib/pq driver https://github.com/lib/pq
func init() {
	genieql.RegisterDriver(PQ, NewDriver(libpq...))
}

// PQ - driver for github.com/lib/pq
const PQ = "github.com/lib/pq"

var libpq = []NullableType{
	{Type: timeExprString, NullType: "pq.NullTime", NullField: "Time"},
}
