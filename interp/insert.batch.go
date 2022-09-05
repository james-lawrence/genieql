package genieql

import (
	"go/ast"
	"go/token"
	"go/types"
	"io"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/generators"
	"bitbucket.org/jatone/genieql/generators/functions"
	"bitbucket.org/jatone/genieql/generators/typespec"
	"bitbucket.org/jatone/genieql/internal/x/stringsx"
)

// InsertBatch configuration interface for generating batch inserts.
type InsertBatch interface {
	genieql.Generator              // must satisfy the generator interface
	Into(string) InsertBatch       // what table to insert into
	Ignore(...string) InsertBatch  // do not attempt to insert the specified column.
	Default(...string) InsertBatch // use the database default for the specified columns.
	Conflict(string) InsertBatch   // specify how conflicts should be handled.
	Batch(n int) InsertBatch       // specify a batch insert
}

// NewInsert instantiate a new insert generator. it uses the name of function
// that calls Define as the name of the generated function.
func NewBatchInsert(
	ctx generators.Context,
	name string,
	comment *ast.CommentGroup,
	cf *ast.Field,
	qf *ast.Field,
	tf *ast.Field,
	scanner *ast.FuncDecl,
) InsertBatch {
	return &batch{
		ctx:     ctx,
		name:    name,
		comment: comment,
		qf:      qf,
		cf:      cf,
		tf:      tf,
		scanner: scanner,
		n:       1,
	}
}

type batch struct {
	ctx      generators.Context
	n        int // number of records to support inserting
	name     string
	table    string
	conflict string
	defaults []string
	ignore   []string
	tf       *ast.Field    // type field.
	cf       *ast.Field    // context field, can be nil.
	qf       *ast.Field    // db Query field.
	scanner  *ast.FuncDecl // scanner being used for results.
	comment  *ast.CommentGroup
}

// Into specify the table the data will be inserted into.
func (t *batch) Into(s string) InsertBatch {
	t.table = s
	return t
}

// Default specify the table columns to be given their default values.
func (t *batch) Default(defaults ...string) InsertBatch {
	t.defaults = defaults
	return t
}

// Ignore specify the table columns to ignore during insert.
func (t *batch) Ignore(ignore ...string) InsertBatch {
	t.ignore = ignore
	return t
}

// Batch specify the maximum number of records to insert.
func (t *batch) Batch(size int) InsertBatch {
	t.n = size
	return t
}

func (t *batch) Conflict(s string) InsertBatch {
	t.conflict = s
	return t
}

func (t *batch) Generate(dst io.Writer) (err error) {
	t.ctx.Println("generation of", t.name, "initiated")
	defer t.ctx.Println("generation of", t.name, "completed")
	t.ctx.Debugln("batch.insert type", t.ctx.CurrentPackage.Name, t.ctx.CurrentPackage.ImportPath, types.ExprString(t.tf.Type))
	t.ctx.Debugln("batch.insert table", t.table)

	initializesig := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{
				t.cf,
				t.qf,
				astutil.Field(&ast.Ellipsis{
					Elt: t.tf.Type,
				}, t.tf.Names...),
			},
		},
		Results: t.scanner.Type.Results,
	}

	typename := stringsx.ToPrivate(t.name + "Type")
	initialize := functions.NewFn(
		astutil.Return(
			&ast.UnaryExpr{
				Op: token.AND,
				X: &ast.CompositeLit{
					Type: ast.NewIdent(typename),
					Elts: []ast.Expr{
						&ast.KeyValueExpr{
							Key:   ast.NewIdent("ctx"),
							Value: t.cf.Names[0],
						},
						&ast.KeyValueExpr{
							Key:   ast.NewIdent("q"),
							Value: t.qf.Names[0],
						},
						&ast.KeyValueExpr{
							Key:   ast.NewIdent("remaining"),
							Value: t.tf.Names[0],
						},
					},
				},
			},
		),
	)

	typedecl := typespec.NewType(typename, &ast.StructType{
		Fields: &ast.FieldList{
			List: []*ast.Field{t.tf},
		},
	})

	return genieql.NewFuncGenerator(func(dst io.Writer) (err error) {
		if err = generators.GenerateComment(generators.DefaultFunctionComment(t.name), t.comment).Generate(dst); err != nil {
			return err
		}

		if err = functions.CompileInto(dst, functions.New(t.name, initializesig), initialize); err != nil {
			return err
		}

		if err = typespec.CompileInto(dst, typedecl); err != nil {
			return err
		}

		return nil
	}).Generate(dst)
}

const batchScannerTemplate = `// New{{.QueryFunction.Name | title}} creates a scanner that inserts a batch of
// records into the database.
func New{{.QueryFunction.Name | title}}({{ .Parameters | arguments }}) {{ .ScannerType | expr }} {
	return &{{.QueryFunction.Name | private}}{
		q: {{.QueryFunction.QueryerName}},
		remaining: {{.Type | name }},
	}
}

type {{.QueryFunction.Name | private}} struct {
	q         {{.QueryFunction.Queryer | expr}}
	remaining {{ .Type.Type | array | expr }}
	scanner   {{ .ScannerType | expr }}
}

func (t *{{.QueryFunction.Name | private}}) Scan(dst *{{.Type.Type | expr}}) error {
	return t.scanner.Scan(dst)
}

func (t *{{.QueryFunction.Name | private}}) Err() error {
	if t.scanner == nil {
		return nil
	}

	return t.scanner.Err()
}

func (t *{{.QueryFunction.Name | private}}) Close() error {
	if t.scanner == nil {
		return nil
	}
	return t.scanner.Close()
}

func (t *{{.QueryFunction.Name | private}}) Next() bool {
	var (
		advanced bool
	)

	if t.scanner != nil && t.scanner.Next() {
		return true
	}

	// advance to the next check
	if len(t.remaining) > 0 && t.Close() == nil {
		t.scanner, t.remaining, advanced = t.advance(t.q, t.remaining...)
		return advanced && t.scanner.Next()
	}

	return false
}

func (t *{{.QueryFunction.Name | private}}) advance(q sqlx.Queryer, {{.Type | name}} ...{{.Type.Type | expr}}) ({{ .ScannerType | expr }}, {{ .Type.Type | array | expr }}, bool) {
	switch len({{.Type | name }}) {
	case 0:
		return nil, []{{.Type.Type | expr}}(nil), false
	{{- range $ctx := .Statements }}
	case {{ $ctx.Number }}:
		{{ $ctx.BuiltinQuery | ast }}
		{{ $ctx.Exploder | ast }}
		{{ range $_, $stmt := $ctx.Explode }}
		{{ $stmt | ast }}
		{{ end }}
		return {{ $.ScannerFunc | expr }}({{ $ctx.Queryer | expr }}), {{$.Type.Type | array | expr}}(nil), true
	{{- end }}
	default:
		{{ .DefaultStatement.BuiltinQuery | ast }}
		{{ .DefaultStatement.Exploder | ast }}
		{{ range $_, $stmt := .DefaultStatement.Explode }}
		{{ $stmt | ast }}
		{{ end }}
		return {{ .ScannerFunc | expr }}({{ .DefaultStatement.Queryer | expr }}), {{.Type | name}}[{{.DefaultStatement.Number}}:], true
	}
}
`
