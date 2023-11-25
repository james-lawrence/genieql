package ffierrors

import (
	"errors"
	"os"

	"bitbucket.org/jatone/genieql/internal/errorsx"
)

const (
	ErrNotImplemented = 999
	ErrUnrecoverable  = 1000
)

func Exit(cause error) {
	var (
		unrecoverable errorsx.Unrecoverable
	)

	if errors.Is(cause, &unrecoverable) {
		os.Exit(ErrUnrecoverable)
	}

	os.Exit(1)
}
