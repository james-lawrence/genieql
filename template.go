package genieql

const scanner = `
type {{ ScannerSuffix .Name }} interface {
	Scan({{ FunctionArgs .Arguments }}) error
}

func NewMaybe{{ .Name }}Scanner (r *sql.Rows, err error) {{ ScannerSuffix .Name }} {
  return {{ ScannerSuffix (Lowercase .Name) }} {
    err: err,
    rows: r,
  }
}

func New{{.Name}}Scanner(r *sql.Rows) {{ ScannerSuffix .Name }} {
  return {{ ScannerSuffix (Lowercase .Name) }} {
    rows: r,
  }
}

type {{ ScannerSuffix (Lowercase .Name) }} struct {
  err error
  rows *sql.Rows
}

func (t {{ ScannerSuffix (Lowercase .Name) }}) Scan({{FunctionArgs .Arguments}}) error {
  if t.err != nil {
    return t.err
  }
  {{ range .Matches }}
  var {{ LocalVariable . }} {{ .Field.Type -}}
  {{ end }}

  if err := t.rows.Scan({{ ScanArgs .Matches }}); err != nil {
    return err
  }

  {{ range .Matches }}
  {{ ArgVariable .ArgPosition}}.{{ .Field.StructField }} = {{ LocalVariable . -}}
  {{ end }}

  return t.rows.Err()
}
`
