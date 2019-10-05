package errorsx

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Compact returns the first error in the set, if any.
func Compact(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

// NewErrRatelimit creates a new rate limit error with the provided backoff.
func NewErrRatelimit(err error, backoff time.Duration) ErrRatelimit {
	return ratelimit{error: err, backoff: backoff}
}

// ErrRatelimit ...
type ErrRatelimit interface {
	error
	Backoff() time.Duration
}

type ratelimit struct {
	backoff time.Duration
	error
}

func (t ratelimit) Backoff() time.Duration {
	return t.backoff
}

// String representing an error, useful for declaring string constants as errors.
type String string

func (t String) Error() string {
	return string(t)
}

// StackChecksum computes a checksum of the given error
// using its stack trace.
func StackChecksum(err error) string {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}

	hash := md5.New()
	if err != nil {
		if failure, ok := err.(stackTracer); ok {
			for _, frame := range failure.StackTrace() {
				hash.Write([]byte(fmt.Sprint(frame)))
			}
		}
	}

	sum := hash.Sum(nil)
	return hex.EncodeToString(sum[:])
}
