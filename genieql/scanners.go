package genieql

import "bitbucket.org/jatone/genieql"

// Scanner - configuration interface for generating scanners.
type Scanner interface {
	genieql.Generator // must satisfy the generator interface
}
