// Package bootstrap provides functions for bootstrapping genieql and reducing boilerplate
// that needs to be written by the user.
//
// TODO: use the definition files in the example directory as the source to reduce code duplication.
package bootstrap

import (
	"fmt"
	"go/build"
	"io"
	"strings"
	"text/template"
)

// NewTableStructure builds a file for defining new structures from tables.
func NewTableStructure(pkg *build.Package) Package {
	const example = `// the genieql.options lines allow for customizing
// the output for the given table(s).
// [rename.columns] section: allows use of a kv mapping to rename columns explicitly.
//genieql.options: [rename.columns] c1=f1
const (
	Table1 = "table1"
	Table2 = "table2"
)`

	return NewSource(
		pkg,
		SourceOptionTags("genieql", "generate", "structure", "table"),
		SourceOptionExample(example),
	)
}

// NewScanners builds a file for defining new scanners from functions.
func NewScanners(pkg *build.Package) Package {
	const example = `// Use builtin types.
type Scanner1 func(i1, i2 int, b1 bool, t1 time.Time)
// Use a data structure, for example from a table mapping.
// type Scanner2 func(e MyType)
// Mix and Match. Note: using two data structures types is only partially supported currently. It only works if column names do not overlap.
// type Scanner3 func(mt MyType, i1 int, i2 int)
`
	return NewSource(
		pkg,
		SourceOptionTags("genieql", "generate", "scanners"),
		SourceOptionExample(example),
	)
}

// NewFunctions builds a file for defining new query functions from functions definitions.
func NewFunctions(pkg *build.Package) Package {
	const example = `type customQueryFunction func(queryer *sql.DB, i1, i2 int, b1 bool, t1 time.Time) Scanner1

func customQueryFunction2(queryer *sql.DB, i1, i2 int, b1 bool, t1 time.Time) Scanner1 {
	const query = "SELECT i1, i2, b1, t1 FROM my_table"
	return nil
}`
	return NewSource(pkg,
		SourceOptionTags("genieql", "generate", "functions"),
		SourceOptionExample(example),
	)
}

// NewInsertBatch builds a file for defining new batch inserts from function definitions.
func NewInsertBatch(pkg *build.Package) Package {
	const example = `// builds a scanner that inserts multiple records into the database.
// the table option must be provided at this time.
// the function parameters must follow the form:
// a queryer,
// an array with the maximum number of records to insert in a single query.
// The return type must be a scanner.
//genieql.options: table=table1
//genieql.options: default-columns=created_at,updated_at
type example1BatchInsertFunction func(queryer *sql.DB, p [5]Table1) NewTable1ScannerStatic`
	return NewSource(pkg,
		SourceOptionTags("genieql", "generate", "insert", "batch"),
		SourceOptionExample(example),
	)
}

// NewGoGenerateDefinitions ...
func NewGoGenerateDefinitions(pkg *build.Package) Package {
	const example = `//go:generate genieql generate experimental structure table constants -o postgresql.table.structs.gen.go
//go:generate genieql generate experimental scanners types -o postgresql.scanners.gen.go
//go:generate genieql generate experimental crud -o postgresql.crud.functions.gen.go --table=example1 --scanner=NewExample1ScannerDynamic --unique-scanner=NewExample1ScannerStaticRow Example1
//go:generate genieql generate experimental functions types -o postgresql.functions.gen.go
//go:generate genieql generate insert experimental batch-function -o postgresql.insert.batch.gen.go`
	return NewSource(
		pkg,
		SourceOptionExample(example),
	)
}

// SourceOption ...
type SourceOption func(*Package)

// SourceOptionTags provide the tags for the source file.
func SourceOptionTags(tags ...string) SourceOption {
	return func(src *Package) {
		src.BuildTags = tags
	}
}

// SourceOptionExample provide an example of the code.
func SourceOptionExample(example string) SourceOption {
	return func(src *Package) {
		src.Example = example
	}
}

// NewSource - generates a source file from the given tags
func NewSource(pkg *build.Package, options ...SourceOption) Package {
	src := Package{
		Package: pkg,
	}

	for _, opt := range options {
		opt(&src)
	}

	return src
}

// Package - used to generate definition files.
type Package struct {
	Package   *build.Package
	Example   string
	BuildTags []string
}

// Generate - writes the definition file to the provided destination.
func (t Package) Generate(dst io.Writer) error {
	return packageTemplate.Execute(dst, t)
}

func flattenTags(tags []string) string {
	if len(tags) > 0 {
		return fmt.Sprintf("//+build %s", strings.Join(tags, ","))
	}
	return ""
}

const _packageTemplate = `{{.BuildTags | flattenTags}}

package {{.Package.Name}}

{{.Example}}
`

var packageTemplate = template.Must(template.New("").Funcs(defaultFuncsMap).Parse(_packageTemplate))
var defaultFuncsMap = template.FuncMap{
	"flattenTags": flattenTags,
}
