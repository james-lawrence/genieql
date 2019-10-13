package generators

import (
	"fmt"
	"go/build"
	"go/token"
	"log"

	"bitbucket.org/jatone/genieql"
)

// Generators generate schema and configuration for testing.
//go:generate dropdb --if-exists -U postgres genieql_test_template
//go:generate createdb -U postgres genieql_test_template
//go:generate psql -1 -f structure.sql genieql_test_template
//go:generate genieql bootstrap --queryer=sqlx.Queryer --driver=github.com/lib/pq --output-file=generators-test.config postgres://$USER@localhost:5432/genieql_test_template?sslmode=disable

// Logging levels
const (
	VerbosityError = iota
	VerbosityWarn
	VerbosityInfo
	VerbosityDebug
	VerbosityTrace
)

// Context - context for generators
type Context struct {
	Build          build.Context
	CurrentPackage *build.Package
	FileSet        *token.FileSet
	Configuration  genieql.Configuration
	Dialect        genieql.Dialect
	Verbosity      int
}

// Println ...
func (t Context) Println(args ...interface{}) {
	if t.Verbosity < VerbosityInfo {
		return
	}

	log.Output(2, fmt.Sprintln(args...))
}

// Printf ...
func (t Context) Printf(format string, args ...interface{}) {
	if t.Verbosity < VerbosityInfo {
		return
	}

	log.Output(2, fmt.Sprintf(format, args...))
}

// Debug logs
func (t Context) Debug(args ...interface{}) {
	if t.Verbosity < VerbosityDebug {
		return
	}

	log.Output(2, fmt.Sprint(args...))
}

// Debugf logs
func (t Context) Debugf(format string, args ...interface{}) {
	if t.Verbosity < VerbosityDebug {
		return
	}

	log.Output(2, fmt.Sprintf(format, args...))
}

// Debugln logs
func (t Context) Debugln(args ...interface{}) {
	if t.Verbosity < VerbosityDebug {
		return
	}

	log.Output(2, fmt.Sprintln(args...))
}

// Traceln detailed logging
func (t Context) Traceln(args ...interface{}) {
	if t.Verbosity < VerbosityTrace {
		return
	}

	log.Output(2, fmt.Sprintln(args...))
}

func reserved(s string) bool {
	switch s {
	case "type":
		return true
	case "func":
		return true
	case "default":
		return true
	default:
		return false
	}
}
