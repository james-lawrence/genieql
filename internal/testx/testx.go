package testx

import (
	"io"
	"os"

	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/iox"
)

func Fixture(fixture string) []byte {
	buf, err := os.ReadFile(fixture)
	errorsx.PanicOnError(err)
	return buf
}

// ReadString reads the entire string from a reader.
// if the reader is also a seeker it'll rewind before reading.
// will panic on error.
func ReadString(in io.Reader) (s string) {
	var (
		err error
		raw []byte
	)

	if seeker, ok := in.(io.Seeker); ok {
		if err = iox.Rewind(seeker); err != nil {
			panic(err)
		}
	}

	if raw, err = io.ReadAll(in); err != nil {
		panic(err)
	}

	return string(raw)
}
