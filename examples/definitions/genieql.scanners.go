//go:build genieql && generate && scanners
// +build genieql,generate,scanners

// genieql.options: config=default.config
package definitions

import "time"

type ProfileScanner func(i1, i2 int, b1 bool, t1 time.Time)

type Example1Scanner func(e Example1)

type ComboScanner func(e1 Example1, e2 Example2)
