package cmd

import (
	"bytes"
	"io"
	"log"
	"os"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/iox"
)

type errWriter struct {
	err error
}

func (t errWriter) Write([]byte) (int, error) {
	return 0, t.err
}

func (t errWriter) Close() error {
	return t.err
}

// DefaultWriteFlags default write flags for WriteStdoutOrFile
const DefaultWriteFlags = os.O_CREATE | os.O_TRUNC | os.O_RDWR

// WriteStdoutOrFile writes to stdout if fpath is empty.
func WriteStdoutOrFile(g genieql.Generator, fpath string, flags int) error {
	var (
		err error
		dst io.WriteCloser = os.Stdout
		buf bytes.Buffer
	)

	if err = g.Generate(&buf); err != nil {
		log.Printf("%s: failed to generate: %+v\n", genieql.PrintDebug(), err)
		return err
	}

	if len(fpath) > 0 {
		log.Println("writing results to", fpath)
		if dst, err = os.OpenFile(fpath, flags, 0666); err != nil {
			dst = errWriter{err: errorsx.Wrap(err, "")}
		}
		defer dst.Close()
	}

	if _, err = io.Copy(dst, &buf); err != nil {
		return err
	}

	return err
}

// StdoutOrFile returns a writer for the path or stdout
func StdoutOrFile(fpath string, flags int) (dst io.WriteCloser, err error) {
	if len(fpath) > 0 {
		log.Println("writing results to", fpath)
		return os.OpenFile(fpath, flags, 0666)
	}

	return iox.NoopWriteCloser{WriteCloser: os.Stdout}, nil
}
