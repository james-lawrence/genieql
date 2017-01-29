//+build genieql,generate,insert,batch

package functions

import "bitbucket.org/jatone/genieql/internal/sqlx"

//genieql.options: table=example1
//genieql.options: default-columns=id,created_at,updated_at
type example1BatchInsertFunction func(queryer sqlx.Queryer, p [5]Example1) NewExample1ScannerStatic
