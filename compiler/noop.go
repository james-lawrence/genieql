package compiler

// // Noop matcher - matches anything.
// func Noop(ctx Context, i *interp.Interpreter, src *ast.File, fn *ast.FuncDecl) (r Result, err error) {
// 	return Result{
// 		Generator: genieql.NewFuncGenerator(func(dst io.Writer) error { return nil }),
// 		Priority:  PriorityFunctions,
// 	}, nil
// }
