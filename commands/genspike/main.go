package main

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

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

	// printspike("example2.go")
	// fmt.Println()
	genspike(scannerName, columnMap)
	fmt.Println()
	parseExpr("*sso.Identity")
	parseExpr("sso.Identity")
	parseExpr("t.rows.Scan()")
	parseExpr("time.Time")
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
	ColumnMaps []ColumnMap
}

type ColumnMap struct {
	Column     *ast.Ident
	Type       ast.Expr
	Assignment ast.Expr
}

func genspike(name string, mapping []ColumnMap) {
	fset := token.NewFileSet()

	var scannerTypeDecl = &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{
			&ast.TypeSpec{
				Name: &ast.Ident{
					Name: name,
					Obj: &ast.Object{
						Kind: ast.Typ,
						Name: name,
					},
				},
				Type: &ast.StructType{
					Fields: &ast.FieldList{
						List: []*ast.Field{
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "err",
										Obj: &ast.Object{
											Kind: ast.Var,
											Name: "err",
										},
									},
								},
								Type: &ast.Ident{
									Name: "error",
								},
							},
							&ast.Field{
								Names: []*ast.Ident{
									&ast.Ident{
										Name: "rows",
										Obj: &ast.Object{
											Kind: ast.Var,
											Name: "rows",
										},
									},
								},
								Type: &ast.StarExpr{
									X: &ast.SelectorExpr{
										X: &ast.Ident{
											Name: "sql",
										},
										Sel: &ast.Ident{
											Name: "Rows",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	var funcDecl = &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				&ast.Field{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: "t",
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: "t",
							},
						},
					},
					Type: &ast.Ident{
						Name: name,
					},
				},
			},
		},
		Name: &ast.Ident{
			Name: "Scan",
		},
		Type: &ast.FuncType{
			Params:  &ast.FieldList{},
			Results: &ast.FieldList{},
		}, // scan-function-ast.go
		Body: &ast.BlockStmt{
			List: []ast.Stmt{},
		}, // scannerbody
	}

	funcDecl.Type.Params.List = FuncParams(SExpr(ssoIdentity))
	funcDecl.Type.Results.List = FuncResults(&ast.Ident{Name: "error"})
	funcDecl.Body = BlockStmtBuilder{&ast.BlockStmt{}}.Append(
		errorCheckStatement,
	).Append(
		DeclarationStatements(mapping...)...,
	).Append(
		ScanStatement(AsUnaryExpr(ColumnToExpr(mapping)...)...),
	).Append(
		AssignmentStatements(mapping)...,
	).Append(
		scannerReturnStatement,
	).BlockStmt

	if err := format.Node(os.Stdout, fset, scannerTypeDecl); err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("\n\n")
	if err := format.Node(os.Stdout, fset, funcDecl); err != nil {
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

func FuncParams(parameters ...ast.Expr) []*ast.Field {
	result := make([]*ast.Field, 0, len(parameters))

	for i, expr := range parameters {
		paramName := fmt.Sprintf("arg%d", i)
		param := &ast.Field{
			Names: []*ast.Ident{
				&ast.Ident{
					Name: paramName,
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: paramName,
					},
				},
			},
			Type: expr,
		}

		result = append(result, param)
	}

	return result
}

func FuncResults(parameters ...ast.Expr) []*ast.Field {
	result := make([]*ast.Field, 0, len(parameters))

	for _, expr := range parameters {
		param := &ast.Field{
			Names: []*ast.Ident{},
			Type:  expr,
		}

		result = append(result, param)
	}

	return result
}

func SExpr(selector *ast.SelectorExpr) ast.Expr {
	return &ast.StarExpr{
		X: selector,
	}
}

type BlockStmtBuilder struct {
	*ast.BlockStmt
}

func (t BlockStmtBuilder) Append(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(t.List, statements...)
	return t
}

func (t BlockStmtBuilder) Prepend(statements ...ast.Stmt) BlockStmtBuilder {
	t.List = append(statements, t.List...)
	return t
}

func DeclarationStatements(maps ...ColumnMap) []ast.Stmt {
	results := make([]ast.Stmt, 0, len(maps))
	for _, m := range maps {
		results = append(results, DeclarationStatement(m))
	}

	return results
}

func DeclarationStatement(m ColumnMap) ast.Stmt {
	return &ast.DeclStmt{
		Decl: &ast.GenDecl{
			Tok: token.VAR,
			Specs: []ast.Spec{
				&ast.ValueSpec{
					Names: []*ast.Ident{
						&ast.Ident{
							Name: m.Column.Name,
							Obj: &ast.Object{
								Kind: ast.Var,
								Name: m.Column.Name,
							},
						},
					},
					Type: m.Type,
				},
			},
		},
	}
}

func ScanStatement(columns ...ast.Expr) ast.Stmt {
	return &ast.IfStmt{
		Init: &ast.AssignStmt{
			Lhs: []ast.Expr{
				&ast.Ident{
					Name: "err",
					Obj: &ast.Object{
						Kind: ast.Var,
						Name: "err",
					},
				},
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				&ast.CallExpr{
					Fun: &ast.SelectorExpr{
						X: &ast.SelectorExpr{
							X: &ast.Ident{
								Name: "t",
							},
							Sel: &ast.Ident{
								Name: "rows",
							},
						},
						Sel: &ast.Ident{
							Name: "Scan",
						},
					},
					Args: columns,
				},
			},
		},
		Cond: &ast.BinaryExpr{
			X: &ast.Ident{
				Name: "err",
			},
			Op: token.NEQ,
			Y: &ast.Ident{
				Name: "nil",
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.Ident{
							Name: "err",
						},
					},
				},
			},
		},
	}
}

func AsUnaryExpr(expressions ...ast.Expr) []ast.Expr {
	results := make([]ast.Expr, 0, len(expressions))
	for _, expr := range expressions {
		unary := &ast.UnaryExpr{
			Op: token.AND,
			X:  expr,
		}
		results = append(results, unary)
	}

	return results
}

func ColumnToExpr(columns []ColumnMap) []ast.Expr {
	result := make([]ast.Expr, 0, len(columns))
	for _, m := range columns {
		result = append(result, m.Column)
	}

	return result
}

func AssignmentStatements(columns []ColumnMap) []ast.Stmt {
	result := make([]ast.Stmt, 0, len(columns))

	for _, m := range columns {
		assignmentStmt := &ast.AssignStmt{
			Lhs: []ast.Expr{
				m.Assignment,
			},
			Tok: token.ASSIGN,
			Rhs: []ast.Expr{
				m.Column,
			},
		}
		result = append(result, assignmentStmt)
	}

	return result
}

var columnMap = []ColumnMap{
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

var scannerReturnStatement = &ast.ReturnStmt{
	Results: []ast.Expr{
		&ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X: &ast.SelectorExpr{
					X: &ast.Ident{
						Name: "t",
					},
					Sel: &ast.Ident{
						Name: "rows",
					},
				},
				Sel: &ast.Ident{
					Name: "Err",
				},
			},
		},
	},
}

var errorCheckStatement = &ast.IfStmt{
	Cond: &ast.BinaryExpr{
		X: &ast.SelectorExpr{
			X: &ast.Ident{
				Name: "t",
			},
			Sel: &ast.Ident{
				Name: "err",
			},
		},
		Op: token.NEQ,
		Y: &ast.Ident{
			Name: "nil",
		},
	},
	Body: &ast.BlockStmt{
		List: []ast.Stmt{
			&ast.ReturnStmt{
				Results: []ast.Expr{
					&ast.SelectorExpr{
						X: &ast.Ident{
							Name: "t",
						},
						Sel: &ast.Ident{
							Name: "err",
						},
					},
				},
			},
		},
	},
}
