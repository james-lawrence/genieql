package compiler

import (
	"bytes"
	"fmt"
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

func nodeInfo(ctx Context, n ast.Node) string {
	pos := ctx.FileSet.PositionFor(n.Pos(), true).String()
	switch n := n.(type) {
	case *ast.FuncDecl:
		return fmt.Sprintf("(%s.%s - %s)", ctx.CurrentPackage.Name, n.Name, pos)
	default:
		return fmt.Sprintf("(%s.%T - %s)", ctx.CurrentPackage.Name, n, pos)
	}
}
