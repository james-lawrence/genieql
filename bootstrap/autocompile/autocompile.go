// Package autocompile contains the embedded archive for bootstraping a package.
package autocompile

import "embed"

// Archive for bootstraping packages.
//
//go:embed genieql.cmd.go genieql.input.go
var Archive embed.FS
