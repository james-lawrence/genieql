//+build genieql,generate,scanners

package scanners

import (
	"bitbucket.org/jatone/genieql/generators/internal/scanners/alternate1"
	"bitbucket.org/jatone/genieql/generators/internal/scanners/alternate2"
)

type ComboScanner func(t1 alternate1.Type1, t2 alternate2.Type1, t3 Type1)
