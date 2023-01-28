package astcodec

import (
	"go/format"
	"io"
	"os"

	"bitbucket.org/jatone/genieql/internal/iox"
	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
)

// FormatOutput formats and resolves imports for the raw bytes representing a go
// source file and writes them into the dst.
func FormatOutput(dst io.Writer, raw []byte) (err error) {
	if raw, err = imports.Process("generated.go", raw, nil); err != nil {
		return errors.Wrap(err, "failed to add required imports")
	}

	if raw, err = format.Source(raw); err != nil {
		return errors.Wrap(err, "failed to format source")
	}

	_, err = dst.Write(raw)

	return errors.Wrap(err, "failed to write to completed code to destination")
}

// ReformatFile a file
func ReformatFile(in *os.File) (err error) {
	var (
		raw []byte
	)

	// ensure we're at the start of the file.
	if err = iox.Rewind(in); err != nil {
		return err
	}

	if raw, err = io.ReadAll(in); err != nil {
		return err
	}

	if raw, err = imports.Process("generated.go", []byte(string(raw)), nil); err != nil {
		return errors.Wrap(err, "failed to add required imports")
	}

	// ensure we're at the start of the file.
	if err = iox.Rewind(in); err != nil {
		return err
	}

	if err = in.Truncate(0); err != nil {
		return errors.Wrap(err, "failed to truncate file")
	}

	if _, err = in.Write(raw); err != nil {
		return errors.Wrap(err, "failed to write formatted content")
	}

	return nil
}

// Format arbitrary source fragment.
func Format(s string) (_ string, err error) {
	var (
		raw []byte
	)

	if raw, err = imports.Process("generated.go", []byte(s), &imports.Options{Fragment: true, Comments: true, TabIndent: true, TabWidth: 8}); err != nil {
		return "", errors.Wrap(err, "failed to add required imports")
	}

	if raw, err = format.Source(raw); err != nil {
		return "", errors.Wrap(err, "failed to format source")
	}

	return string(raw), nil
}
