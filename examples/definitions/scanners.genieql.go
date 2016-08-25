//+build genieql,generate,scanners

//genieql.options: config=default.config
package definitions

import "time"

type ProfileScanner func(i1, i2 int, b1 bool, t1 time.Time)

type Example1Scanner func(e Example1)
