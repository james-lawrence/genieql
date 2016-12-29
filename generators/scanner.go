package generators

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
	"go/types"
	"io"
	"strings"
	"text/template"

	"github.com/pkg/errors"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/astutil"
	"bitbucket.org/jatone/genieql/x/stringsx"
)

type mode int

func (t mode) Enabled(o mode) bool {
	return t&o != 0
}

const (
	// ModeInterface - output the scanner interface.
	ModeInterface mode = 1 << iota
	// ModeStatic - output the static scanner.
	ModeStatic
	// ModeDynamic - output the dynamic scanner.
	ModeDynamic
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
	// TODO
	return []ScannerOption{}
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

// ScannerOptionOutputMode set the output mode for the scanner.
// e.g.) generators.ModeInterface|generators.ModeStatic|generators.ModeDynamic
// each mode provided enables the given scanner type.
func ScannerOptionOutputMode(m mode) ScannerOption {
	return func(s *scanner) error {
		s.Mode = m
		return nil
	}
}

// ScannerOptionInterfaceName DEPRECATED only used for old functions.
func ScannerOptionInterfaceName(n string) ScannerOption {
	return func(s *scanner) error {
		s.interfaceName = n
		return nil
	}
}

// ScannerOptionIgnoreSet DEPRECATED only used for old functions.
// sets the selection of columns to be ignored.
func ScannerOptionIgnoreSet(n ...string) ScannerOption {
	return func(s *scanner) error {
		s.ignoreSet = n
		return nil
	}
}

// ScannerFromGenDecl creates a structure generator from  from the provided *ast.GenDecl
func ScannerFromGenDecl(decl *ast.GenDecl, providedOptions ...ScannerOption) []genieql.Generator {
	g := make([]genieql.Generator, 0, len(decl.Specs)*2)

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
	return maybeScanner(newScanner(options...))
}

func maybeScanner(s scanner, err error) genieql.Generator {
	if err != nil {
		return genieql.NewErrGenerator(err)
	}
	return s
}

func newScanner(options ...ScannerOption) (scanner, error) {
	// by default enable all modes
	s := scanner{
		Mode: ModeInterface | ModeStatic | ModeDynamic,
	}

	for _, opt := range options {
		if err := opt(&s); err != nil {
			return s, err
		}
	}

	return s, nil
}

type scanner struct {
	Name          string
	interfaceName string // DEPRECATED
	Mode          mode
	Fields        *ast.FieldList
	Package       *build.Package
	Config        genieql.Configuration
	Driver        genieql.Driver
	ignoreSet     []string
}

func (t scanner) Generate(dst io.Writer) error {
	var (
		err  error
		tmpl *template.Template
	)

	type context struct {
		Name          string
		InterfaceName string
		Parameters    []*ast.Field
		Columns       []genieql.ColumnMap
	}

	ctx := context{
		Name:          t.Name,
		InterfaceName: stringsx.ToPublic(stringsx.DefaultIfBlank(t.interfaceName, t.Name)),
		Parameters:    t.Fields.List,
	}

	for _, param := range t.Fields.List {
		var (
			columns []genieql.ColumnMap
		)

		if columns, err = t.columnMaps(param); err != nil {
			return err
		}

		ctx.Columns = append(ctx.Columns, columns...)
	}

	lookupNullableTypes := composeLookupNullableType(DefaultLookupNullableType, t.Driver.LookupNullableType)
	nullableTypes := composeNullableType(DefaultNullableTypes, t.Driver.NullableType)

	funcMap := template.FuncMap{
		"expr":       types.ExprString,
		"scan":       scan,
		"arguments":  argumentsAsPointers,
		"printAST":   astPrint,
		"nulltype":   lookupNullableTypes,
		"assignment": assignmentStmt{NullableType: nullableTypes}.assignment,
		"title":      stringsx.ToPublic,
		"private":    stringsx.ToPrivate,
	}

	if t.Mode.Enabled(ModeInterface) {
		tmpl = template.Must(template.New("interface").Funcs(funcMap).Parse(interfaceScanner))
		if err = tmpl.Execute(dst, ctx); err != nil {
			return errors.Wrap(err, "failed to generate interface scanner")
		}

		dst.Write([]byte("\n"))
	}

	if t.Mode.Enabled(ModeStatic) {
		cc := NewColumnConstantFromFieldList(
			ColumnConstantContext{
				Config:  t.Config,
				Package: t.Package,
			},
			fmt.Sprintf("%sStaticColumns", stringsx.ToPublic(t.Name)),
			genieql.NewColumnInfoNameTransformer(),
			t.Fields,
		)
		if err = cc.Generate(dst); err != nil {
			return err
		}

		tmpl = template.Must(template.New("static").Funcs(funcMap).Parse(staticScanner))
		if err = tmpl.Execute(dst, ctx); err != nil {
			return errors.Wrap(err, "failed to generate static scanner")
		}

		dst.Write([]byte("\n"))

		tmpl = template.Must(template.New("static-row").Funcs(funcMap).Parse(staticRowScanner))
		if err = tmpl.Execute(dst, ctx); err != nil {
			return errors.Wrap(err, "failed to generate static row scanner")
		}

		dst.Write([]byte("\n"))
	}

	if t.Mode.Enabled(ModeDynamic) {
		tmpl = template.Must(template.New("dynamic").Funcs(funcMap).Parse(dynamicScanner))
		if err = tmpl.Execute(dst, ctx); err != nil {
			return errors.Wrap(err, "failed to generate dynamic scanner")
		}
	}

	return nil
}

func (t scanner) columnMaps(param *ast.Field) ([]genieql.ColumnMap, error) {
	if builtinType(param.Type) {
		return builtinParam(param)
	}
	return t.mappedParam(param)
}

// mappedParam converts a *ast.Field that represents a struct into an array
// of ColumnMap.
func (t scanner) mappedParam(param *ast.Field) ([]genieql.ColumnMap, error) {
	var (
		cMap []genieql.ColumnMap
		m    genieql.MappingConfig
	)

	if err := t.Config.ReadMap(packageName(t.Package, param.Type), types.ExprString(param.Type), "default", &m); err != nil {
		return []genieql.ColumnMap{}, err
	}

	aliaser := m.Aliaser()
	columns, err := m.ColumnInfo()

	if err != nil {
		return []genieql.ColumnMap{}, err
	}

	for _, arg := range param.Names {
		for _, column := range columns {
			if stringsx.Contains(column.Name, t.ignoreSet...) {
				continue
			}

			c, err := column.MapColumn(&ast.SelectorExpr{
				Sel: ast.NewIdent(aliaser.Alias(column.Name)),
				X:   arg,
			})
			if err != nil {
				return []genieql.ColumnMap{}, err
			}
			cMap = append(cMap, c)
		}
	}

	return cMap, nil
}

// turns an array of column mappings into the inputs to the
// scan function.
func scan(columns []genieql.ColumnMap) string {
	args := []string{}

	for idx, col := range columns {
		args = append(args, fmt.Sprintf("&%s", types.ExprString(col.Local(idx))))
	}

	return strings.Join(args, ", ")
}

type assignmentStmt struct {
	genieql.NullableType
}

func (t assignmentStmt) assignment(i int, column genieql.ColumnMap) ast.Stmt {
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

const interfaceScanner = `
// {{.InterfaceName}} scanner interface.
type {{.InterfaceName}} interface {
	Scan({{ .Parameters | arguments }}) error
	Next() bool
	Close() error
	Err() error
}

type err{{.InterfaceName}} struct {
	e error
}

func (t err{{.InterfaceName}}) Scan({{ .Parameters | arguments }}) error {
	return t.e
}

func (t err{{.InterfaceName}}) Next() bool {
	return false
}

func (t err{{.InterfaceName}}) Err() error {
	return t.e
}

func (t err{{.InterfaceName}}) Close() error {
	return nil
}
`

const staticScanner = `// New{{.Name | title}}Static creates a scanner that operates on a static
// set of columns that are always returned in the same order.
func New{{.Name | title}}Static(rows *sql.Rows, err error) {{.InterfaceName}} {
	if err != nil {
		return err{{.InterfaceName}}{e: err}
	}

	return {{.Name | private}}Static{
		Rows: rows,
	}
}

type {{.Name | private}}Static struct {
	Rows *sql.Rows
}

func (t {{.Name | private}}Static) Scan({{ .Parameters | arguments }}) error {
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

func (t {{.Name | private}}Static) Err() error {
	return t.Rows.Err()
}

func (t {{.Name | private}}Static) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t {{.Name | private}}Static) Next() bool {
	return t.Rows.Next()
}
`

const staticRowScanner = `// New{{.Name | title}}StaticRow creates a scanner that operates on a static
// set of columns that are always returned in the same order, only scans a single row.
func New{{.Name | title}}StaticRow(row *sql.Row) {{.Name | title}}StaticRow {
	return {{.Name | title}}StaticRow {
		row: row,
	}
}

type {{.Name | title}}StaticRow struct {
	row *sql.Row
}

func (t {{.Name | title}}StaticRow) Scan({{ .Parameters | arguments }}) error {
	var (
		{{- range $index, $column := .Columns }}
		{{ $column.Local $index }} {{ $column.Type | nulltype | expr -}}
		{{ end }}
	)

	if err := t.row.Scan({{ .Columns | scan}}); err != nil {
		return err
	}

	{{ range $index, $column := .Columns}}
	{{ assignment $index $column | printAST }}
	{{ end }}

	return nil
}
`

const dynamicScanner = `
// New{{.Name | title}}Dynamic creates a scanner that operates on a dynamic
// set of columns that can be returned in any subset/order.
func New{{.Name | title}}Dynamic(rows *sql.Rows, err error) {{.InterfaceName}} {
	if err != nil {
		return err{{.InterfaceName}}{e: err}
	}

	return {{.Name | private}}Dynamic{
		Rows: rows,
	}
}

type {{.Name | private}}Dynamic struct {
	Rows *sql.Rows
}

func (t {{.Name | private}}Dynamic) Scan({{ .Parameters | arguments }}) error {
	const (
		{{- range $index, $column := .Columns }}
		{{$column.Name}}{{$index}} = "{{$column.Name}}"
		{{- end }}
	)
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
		case {{$column.Name}}{{$index}}:
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
		case {{$column.Name}}{{$index}}:
			{{ assignment $index $column | printAST -}}
		{{- end }}
		}
	}

	return t.Rows.Err()
}

func (t {{.Name | private}}Dynamic) Err() error {
	return t.Rows.Err()
}

func (t {{.Name | private}}Dynamic) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t {{.Name | private}}Dynamic) Next() bool {
	return t.Rows.Next()
}
`
