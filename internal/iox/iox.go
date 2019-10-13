package iox

import (
	"io"
	"io/ioutil"
)

// Rewind the io.Seeker
func Rewind(i io.Seeker) (err error) {
	_, err = i.Seek(0, io.SeekStart)
	return err
}

// Error discards the byte count and returns just the error.
func Error(_ int64, err error) error {
	return err
}

// ReadString reads the entire string from a reader.
// if the reader is also a seeker it'll rewind before reading.
func ReadString(in io.Reader) (s string, err error) {
	var (
		raw []byte
	)

	if seeker, ok := in.(io.Seeker); ok {
		if err = Rewind(seeker); err != nil {
			return "", err
		}
	}

	if raw, err = ioutil.ReadAll(in); err != nil {
		return "", err
	}

	return string(raw), nil
}
