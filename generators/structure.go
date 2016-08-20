package generators

import (
	"go/ast"
	"go/types"
	"html/template"
	"io"
	"log"

	"bitbucket.org/jatone/genieql"
)

// StructOption option to provide the structure function.
type StructOption func(*structure)

// StructOptionName provide the name of the struct to the structure.
func StructOptionName(n string) StructOption {
	return func(s *structure) {
		s.Name = n
	}
}

// StructOptionFields convience function for wrapping an array of genieql.ColumnInfo
// within a function to act as a fieldsDelegate
func StructOptionFields(c ...genieql.ColumnInfo) StructOption {
	return func(s *structure) {
		s.ColumnsDelegate = func() ([]genieql.ColumnInfo, error) {
			return c, nil
		}
	}
}

// StructOptionFieldsDelegate provides the fields delegate function for lookuping
// up the fields.
func StructOptionFieldsDelegate(delegate FieldsDelegate) StructOption {
	return func(s *structure) {
		s.ColumnsDelegate = delegate
	}
}

func StructOptionConfigurationComment(cg *ast.CommentGroup) StructOption {
	return func(s *structure) {
		s.config = cg
	}
}

type FieldsDelegate func() ([]genieql.ColumnInfo, error)

// NewStructure creates a Generator that builds structures from column information.
func NewStructure(opts ...StructOption) genieql.Generator {
	s := structure{}

	for _, opt := range opts {
		opt(&s)
	}

	return s
}

// StructureFromGenDecl creates a structure generator from  from the provided *ast.GenDecl
func StructureFromGenDecl(decl *ast.GenDecl, fields func(string) FieldsDelegate) []genieql.Generator {
	g := make([]genieql.Generator, 0, len(decl.Specs))
	for _, spec := range decl.Specs {
		if vs, ok := spec.(*ast.ValueSpec); ok {
			for idx := range vs.Names {
				value := types.ExprString(vs.Values[idx])
				s := NewStructure(
					StructOptionName(
						vs.Names[idx].Name,
					),
					StructOptionFieldsDelegate(fields(value)),
					StructOptionConfigurationComment(decl.Doc),
				)
				g = append(g, s)
			}
		}
	}
	return g
}

type structure struct {
	Name            string
	ColumnsDelegate FieldsDelegate
	config          *ast.CommentGroup
}

func (t structure) Generate(dst io.Writer) error {
	type context struct {
		Name    string
		Columns []genieql.ColumnInfo
	}
	const tmpl = `type {{.Name}} struct {
	{{- range $column := .Columns }}
	{{ $column.Name }} {{ if $column.Nullable }}*{{ end }}{{ $column.Type -}}
	{{ end }}
}`

	if t.config != nil {
		log.Println("config text:", t.config.Text())
	}

	columns, err := t.ColumnsDelegate()
	if err != nil {
		return err
	}

	ctx := context{
		Name:    t.Name,
		Columns: columns,
	}
	return template.Must(template.New("").Parse(tmpl)).Execute(dst, ctx)
}
