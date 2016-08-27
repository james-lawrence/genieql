package generators

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
)

// ScannerOption option to provide the structure function.
type ScannerOption func(*scanner) error

// ScannerOptionName provide the base name for the scanners.
func ScannerOptionName(n string) ScannerOption {
	return func(s *scanner) error {
		s.Name = n
		return nil
	}
}

// ScannerOptionParameters provide the parameters for the scanner.
func ScannerOptionParameters(fields *ast.FieldList) ScannerOption {
	return func(s *scanner) error {
		s.Fields = fields
		return nil
	}
}

// ScannerOptionsFromComment extracts configuration from a comment group attached
// to the function type.
func ScannerOptionsFromComment(comment *ast.CommentGroup) []ScannerOption {
	options := make([]ScannerOption, 0, 10)

	return options
}

// ScannerOptionPackage provides the package being used to build the scanners.
func ScannerOptionPackage(p *build.Package) ScannerOption {
	return func(s *scanner) error {
		s.Package = p
		return nil
	}
}

// ScannerOptionConfiguration provides the configuration being used to build the scanners.
func ScannerOptionConfiguration(c genieql.Configuration) ScannerOption {
	return func(s *scanner) error {
		var (
			err error
		)
		s.Config = c
		s.Driver, err = genieql.LookupDriver(c.Driver)
		return err
	}
}

// ScannerFromGenDecl creates a structure generator from  from the provided *ast.GenDecl
func ScannerFromGenDecl(decl *ast.GenDecl, providedOptions ...ScannerOption) []genieql.Generator {
	g := make([]genieql.Generator, 0, len(decl.Specs))

	for _, spec := range decl.Specs {
		if ts, ok := spec.(*ast.TypeSpec); ok {
			if ft, ok := ts.Type.(*ast.FuncType); ok {
				options := append(
					providedOptions,
					ScannerOptionName(ts.Name.Name),
					ScannerOptionParameters(ft.Params),
				)

				options = append(options, ScannerOptionsFromComment(decl.Doc)...)

				g = append(g, NewScanner(options...))
			}
		}
	}

	return g
}

// NewScanner creates a new genieql.Generator from the provided scanner options.
func NewScanner(options ...ScannerOption) genieql.Generator {
	s := scanner{}

	for _, opt := range options {
		if err := opt(&s); err != nil {
			return genieql.NewErrGenerator(err)
		}
	}

	return s
}

type scanner struct {
	Name    string
	Fields  *ast.FieldList
	Package *build.Package
	Config  genieql.Configuration
	Driver  genieql.Driver
}

func (t scanner) Generate(dst io.Writer) error {
	var (
		err  error
		tmpl *template.Template
	)

	type context struct {
		Name       string
		Parameters []*ast.Field
		Columns    []genieql.ColumnMap2
	}

	ctx := context{
		Name:       t.Name,
		Parameters: t.Fields.List,
	}

	for _, param := range t.Fields.List {
		var (
			columns []genieql.ColumnMap2
		)

		if columns, err = t.columnMapper(param); err != nil {
			return err
		}

		ctx.Columns = append(ctx.Columns, columns...)
	}

	lookupNullableTypes := composeLookupNullableType(DefaultLookupNullableType, t.Driver.LookupNullableType)
	nullableTypes := composeNullableType(DefaultNullableTypes, t.Driver.NullableType)

	funcMap := template.FuncMap{
		"expr":       types.ExprString,
		"scan":       scan,
		"arguments":  arguments,
		"nulltype":   lookupNullableTypes,
		"assignment": assignmentStmt{NullableType: nullableTypes}.assignment,
		"printAST":   astPrint,
	}

	tmpl = template.Must(template.New("interface").Funcs(funcMap).Parse(interfaceScanner))
	if err = tmpl.Execute(dst, ctx); err != nil {
		return errors.Wrap(err, "failed to generate interface scanner")
	}

	dst.Write([]byte("\n"))

	tmpl = template.Must(template.New("static").Funcs(funcMap).Parse(staticScanner))
	if err = tmpl.Execute(dst, ctx); err != nil {
		return errors.Wrap(err, "failed to generate static scanner")
	}

	dst.Write([]byte("\n"))

	tmpl = template.Must(template.New("dynamic").Funcs(funcMap).Parse(dynamicScanner))
	if err = tmpl.Execute(dst, ctx); err != nil {
		return errors.Wrap(err, "failed to generate dynamic scanner")
	}

	return nil
}

func (t scanner) packageName(x ast.Expr) string {
	switch x := x.(type) {
	case *ast.SelectorExpr:
		// TODO
		log.Println("imports", x.Sel.Name, t.Package.Imports)
		return ""
	default:
		return t.Package.ImportPath
	}
}

func (t scanner) columnMapper(param *ast.Field) ([]genieql.ColumnMap2, error) {
	if builtinType(param.Type) {
		return builtinParam(param)
	}
	return t.mappedParam(param)
}

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnMap2.
func (t scanner) mappedParam(param *ast.Field) ([]genieql.ColumnMap2, error) {
	var (
		cMap []genieql.ColumnMap2
		m    genieql.MappingConfig
	)

	if err := t.Config.ReadMap(t.packageName(param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return []genieql.ColumnMap2{}, err
	}

	aliaser := m.Aliaser()
	columns, err := m.ColumnInfo()

	if err != nil {
		return []genieql.ColumnMap2{}, err
	}

	for _, arg := range param.Names {
		for _, column := range columns {
			c, err := column.MapColumn(&ast.SelectorExpr{
				Sel: ast.NewIdent(aliaser.Alias(column.Name)),
				X:   arg,
			})
			if err != nil {
				return []genieql.ColumnMap2{}, err
			}
			cMap = append(cMap, c)
		}
	}

	return cMap, nil
}

func builtinType(x ast.Expr) bool {
	name := types.ExprString(x)
	for _, t := range types.Typ {
		if name == t.Name() {
			return true
		}
	}

	switch name {
	case "time.Time":
		return true
	default:
		return false
	}
}

// builtinParam converts a *ast.Field that represents a builtin type
// (time.Time, int,float,bool, etc) into an array of ColumnMap2.
func builtinParam(param *ast.Field) ([]genieql.ColumnMap2, error) {
	columns := make([]genieql.ColumnMap2, 0, len(param.Names))
	for _, name := range param.Names {
		columns = append(columns, genieql.ColumnMap2{
			Name:   name.Name,
			Type:   &ast.StarExpr{X: param.Type},
			Dst:    &ast.StarExpr{X: name},
			PtrDst: false,
		})
	}
	return columns, nil
}

func arguments(fields []*ast.Field) string {
	result := []string{}
	for _, field := range fields {
		result = append(result,
			strings.Join(
				astutil.MapExprToString(astutil.MapIdentToExpr(field.Names...)...),
				", ",
			)+" "+types.ExprString(&ast.StarExpr{X: field.Type}))
	}
	return strings.Join(result, ", ")
}

// turns an array of column mappings into the inputs into the inputs to the
// scan function.
func scan(columns []genieql.ColumnMap2) string {
	args := []string{}

	for idx, col := range columns {
		args = append(args, fmt.Sprintf("&%s", types.ExprString(col.Local(idx))))
	}

	return strings.Join(args, ", ")
}

type assignmentStmt struct {
	genieql.NullableType
}

func (t assignmentStmt) assignment(i int, column genieql.ColumnMap2) ast.Stmt {
	var (
		local         = column.Local(i)
		nullExpresion ast.Expr
		nullable      bool
	)

	if nullExpresion, nullable = t.NullableType(column.Type, local); !nullable {
		assignVal := astutil.ExprList(types.ExprString(local))
		if column.PtrDst {
			assignVal = astutil.ExprList("&" + types.ExprString(local))
		}

		return astutil.Assign(
			[]ast.Expr{column.Dst},
			token.ASSIGN,
			assignVal,
		)
	}

	tmpVar := astutil.ExprList("tmp")
	assignVal := tmpVar
	if column.PtrDst {
		assignVal = astutil.ExprList("&tmp")
	}

	return &ast.IfStmt{
		Cond: &ast.SelectorExpr{
			X:   local,
			Sel: ast.NewIdent("Valid"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				astutil.Assign(tmpVar, token.DEFINE, []ast.Expr{nullExpresion}),
				astutil.Assign([]ast.Expr{column.Dst}, token.ASSIGN, assignVal),
			},
		},
	}
}

func astPrint(n ast.Node) (string, error) {
	dst := bytes.NewBuffer([]byte{})
	fset := token.NewFileSet()
	if err := printer.Fprint(dst, fset, n); err != nil {
		log.Println("failure to print ast", err)
		return "", err
	}
	return dst.String(), nil
}

const interfaceScanner = `// {{.Name}} scanner interface.
type {{.Name}} interface {
	Scan({{ .Parameters | arguments }}) error
	Next() bool
	Close() error
	Err() error
}

type err{{.Name}} struct {
	e error
}

func (t err{{.Name}}) Scan({{ .Parameters | arguments }}) error {
	return t.e
}

func (t err{{.Name}}) Next() bool {
	return false
}

func (t err{{.Name}}) Err() error {
	return t.e
}

func (t err{{.Name}}) Close() error {
	return nil
}
`

const staticScanner = `// Static{{.Name}} creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func Static{{.Name}}(rows *sql.Rows, err error) {{.Name}} {
	if err != nil {
		return err{{.Name}}{e: err}
	}

	return static{{.Name}}{
		Rows: rows,
	}
}

type static{{.Name}} struct {
	Rows *sql.Rows
}

func (t static{{.Name}}) Scan({{ .Parameters | arguments }}) error {
	var (
		{{- range $index, $column := .Columns }}
		{{ $column.Local $index }} {{ $column.Type | nulltype | expr -}}
		{{ end }}
	)

	if err := t.Rows.Scan({{ .Columns | scan}}); err != nil {
		return err
	}

	{{ range $index, $column := .Columns}}
	{{ assignment $index $column | printAST }}
	{{ end }}

	return t.Rows.Err()
}

func (t static{{.Name}}) Err() error {
	return t.Rows.Err()
}

func (t static{{.Name}}) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t static{{.Name}}) Next() bool {
	return t.Rows.Next()
}
`

const dynamicScanner = `
// Dynamic{{.Name}} creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func Dynamic{{.Name}}(rows *sql.Rows, err error) {{.Name}} {
	if err != nil {
		return err{{.Name}}{e: err}
	}

	return dynamic{{.Name}}{
		Rows: rows,
	}
}

type dynamic{{.Name}} struct {
	Rows *sql.Rows
}

func (t dynamic{{.Name}}) Scan({{ .Parameters | arguments }}) error {
	var (
		ignored sql.RawBytes
		err     error
		columns []string
		dst     []interface{}
		{{- range $index, $column := .Columns }}
		{{ $column.Local $index }} {{ $column.Type | nulltype | expr -}}
		{{ end }}
	)

	if columns, err = t.Rows.Columns(); err != nil {
		return err
	}

	dst = make([]interface{}, 0, len(columns))

	for _, column := range columns {
		switch column {
		{{- range $index, $column := .Columns }}
		case "{{$column.Name}}":
			dst = append(dst, &{{ $column.Local $index -}})
		{{- end }}
		default:
			dst = append(dst, &ignored)
		}
	}

	if err := t.Rows.Scan(dst...); err != nil {
		return err
	}

	for _, column := range columns {
		switch column {
		{{- range $index, $column := .Columns}}
		case "{{$column.Name}}":
			{{ assignment $index $column | printAST -}}
		{{- end }}
		}
	}

	return t.Rows.Err()
}

func (t dynamic{{.Name}}) Err() error {
	return t.Rows.Err()
}

func (t dynamic{{.Name}}) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t dynamic{{.Name}}) Next() bool {
	return t.Rows.Next()
}
`
