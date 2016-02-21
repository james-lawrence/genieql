package commands

import (
	"io"
	"log"
	"os"
)

// WriteStdoutOrFile writes to stdout if fpath is empty.
func WriteStdoutOrFile(fpath string, flags int, src io.Reader) error {
	var err error
	var dst io.WriteCloser = os.Stdout
	if len(fpath) > 0 {
		log.Println("Writing Results to", fpath)
		if dst, err = os.OpenFile(fpath, flags, 0666); err != nil {
			return err
		}
		defer dst.Close()
	}

	_, err = io.Copy(dst, src)
	return err
}
