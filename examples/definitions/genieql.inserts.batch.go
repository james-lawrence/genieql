//+build genieql,generate,insert,batch

package definitions

import "bitbucket.org/jatone/genieql/sqlx"

type example1BatchInsertFunction func(queryer sqlx.Queryer, p [5]Example1) NewExample1ScannerStatic
