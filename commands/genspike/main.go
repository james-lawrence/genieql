package main

import (
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"bitbucket.org/jatone/genieql"

	"gopkg.in/alecthomas/kingpin.v2"
)

// data stored in qlgenie.conf - dialect, default alias strategy, map definition directory,
// default for including table prefix aliases, database connection information.
// qlgenie bootstrap psql://host:port/example?username=x&password=y -> creates example.qlgenie.
// qlgenie bootstrap --ouput="someothername.qlgenie" psql://host:port/example?username=x&password=y -> creates someothername.qlgenie
// step 1) define your mappings, will be placed in a yaml definition file, only 1 allowed per type/table pairing.
// TODO: automatically determine natural key from database schema.
// natural-key defaults to "id"
// qlgenie map --config="example.glgenie" --include-table-prefix-aliases=false --natural-key="composite,column,names" {Package}.{Type} {TableName} snakecase lowercase
// qlgenie map display --config="example.qlgenie" {Package}.{Type} {TableName} // displays file location, and contents to stdout as yml.
// step 2) autogenerate your crud functions SELECT, UPDATE, DELETE using the natural-key defined by the Package.Type and Table pairing.
// go:generate qlgenie generate crud {Package}.{Type} {TableName}
// step 3) build a scanner for a custom query.
// go:generate qlgenie scanner fromfile --name="MyScanner" --package="github.com/jatone/project" --output="my_scanner" {query_file} github.com/jatone/project.TypeA github.com/jatone/project.TypeB
// go:generate qlgenie scanner fromconstant --name="MyScanner" --package="github.com/jatone/project" --output="my_scanner" {package}.{name} github.com/jatone/project.TypeA github.com/jatone/project.TypeB
func main() {
	var packageName string
	var outputFilename string
	var scannerName string
	var types []string

	app := kingpin.New("qlscanner", "qlscanner generates scanner methods for the provided types and query")
	app.Flag("package", "name of the package the scanner is to be placed").Required().StringVar(&packageName)
	app.Flag("output", "output file for the scanner type").Default("").StringVar(&outputFilename)
	app.Flag("name", "name of the scanner type to create").Required().StringVar(&scannerName)
	app.Arg("types", "types that will be filled in by the scanner").StringsVar(&types)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.Println("Package", packageName)
	log.Println("outputfile", outputFilename)
	log.Println("scanner name", scannerName)
	log.Println("types", types)
	log.Println("len(types)", len(types))
	for _, s := range types {
		p, t := extractPackageType(s)
		log.Printf("Package: %s, Type: %s\n", p, t)
	}

	// printspike("example1.go")
	// printspike("example2.go")
	printspike("example3.go")
	// fmt.Println()
	// genspike(scannerName, columnMap)
	// fmt.Println()
	// parseExpr("*sso.Identity")
	// parseExpr("sso.Identity")
	// parseExpr("t.rows.Scan()")
	// parseExpr("time.Time")
}

func printspike(filename string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, 0)

	if err != nil {
		log.Fatalln(err)
	}

	ast.Print(fset, f)
}

func parseExpr(s string) {
	r, err := parser.ParseExpr(s)
	if err != nil {
		log.Println("err parsing expression", err)
		return
	}
	log.Printf("%#v\n", r)
}

func extractPackageType(s string) (string, string) {
	i := strings.LastIndex(s, ".")
	return s[:i], s[i+1:]
}

type Destination struct {
	Package    string
	Ident      string
	ColumnMaps []genieql.ColumnMap
}

func genspike(name string, mapping []genieql.ColumnMap) {
	fset := token.NewFileSet()

	scanner := genieql.Scanner{Name: name}.Build(mapping, ssoIdentity)

	if err := format.Node(os.Stdout, fset, scanner); err != nil {
		log.Fatalln(err)
	}
}

var ssoIdentity = &ast.SelectorExpr{
	X: &ast.Ident{
		Name: "sso",
	},
	Sel: &ast.Ident{
		Name: "Identity",
	},
}

var columnMap = []genieql.ColumnMap{
	{
		Column: &ast.Ident{
			Name: "c0",
		},
		Type: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "time",
			},
			Sel: &ast.Ident{
				Name: "Time",
			},
		},
		Assignment: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "arg0",
			},
			Sel: &ast.Ident{
				Name: "Created",
			},
		},
	},
	{
		Column: &ast.Ident{
			Name: "c1",
		},
		Type: &ast.Ident{
			Name: "string",
		},
		Assignment: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "arg0",
			},
			Sel: &ast.Ident{
				Name: "Email",
			},
		},
	},
	{
		Column: &ast.Ident{
			Name: "c2",
		},
		Type: &ast.Ident{
			Name: "string",
		},
		Assignment: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "arg0",
			},
			Sel: &ast.Ident{
				Name: "ID",
			},
		},
	},
}
