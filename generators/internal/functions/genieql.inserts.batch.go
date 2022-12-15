//go:build genieql && generate && insert && batch
// +build genieql,generate,insert,batch

package functions

import "bitbucket.org/jatone/genieql/internal/sqlx"

// genieql.options: table=example4
// genieql.options: default-columns=created_at,updated_at
type example4BatchInsertFunction func(queryer sqlx.Queryer, p [5]Example4) NewExample4ScannerStatic
