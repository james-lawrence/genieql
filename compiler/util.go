package compiler

import (
	"bytes"
	"go/ast"
	"go/format"

	"github.com/pkg/errors"
)

func formatSource(ctx Context, src *ast.File) (_ string, err error) {
	var (
		buf bytes.Buffer
	)

	if err = format.Node(&buf, ctx.FileSet, src); err != nil {
		return "", errors.Wrap(err, "failed to format")
	}

	return buf.String(), nil
}
