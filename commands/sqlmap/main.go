package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"
)

type typeDeclaration2 struct {
	PackageName string
	TypeIdent   string
	Fields      *ast.FieldList
}

func (t typeDeclaration2) TypeString() string {
	return t.PackageName + "." + t.TypeIdent
}

func (t typeDeclaration2) Aliases(aliasers ...genieql.Aliaser) ([]genieql.Field, []error) {
	fields := make([]genieql.Field, 0, t.Fields.NumFields())
	errors := make([]error, 0, t.Fields.NumFields())

	for _, field := range t.Fields.List {
		r, err := ASTFieldToGenieqlField(field, aliasers...)
		fields = append(fields, r)
		if err != nil {
			errors = append(errors, err)
		}
	}

	return fields, errors
}

type typeDeclaration struct {
	GenDecl *ast.GenDecl
	Package *ast.Package
}

func main() {
	var packageName string
	var typeName string
	var tableName string
	var strategyArray []string
	var customAliases []string

	app := kingpin.New("sqlaliaser", "sqlaliaser generate field to column aliases from types")
	app.Flag("package", "package where type is located").Required().StringVar(&packageName)
	app.Flag("type", "type to generate field aliases").Required().StringVar(&typeName)
	app.Flag("table", "table the type maps to").Required().StringVar(&tableName)
	app.Flag("alias", "custom aliases").StringsVar(&customAliases)
	app.Arg("strategy", "alias generation strategy").StringsVar(&strategyArray)

	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.Println("Package", packageName)
	log.Println("Type", typeName)
	log.Println("Table", tableName)
	log.Println("Strategy Array", strategyArray)
	log.Println("Custom Aliases", customAliases)

	fset := token.NewFileSet()
	packages := []*ast.Package{}

	for _, srcDir := range build.Default.SrcDirs() {
		directory := filepath.Join(srcDir, packageName)
		// todo debug.
		// log.Println("Importing", directory)
		pkg, err := build.Default.ImportDir(directory, build.FindOnly)
		if err != nil {
			log.Fatalln("Default.ImportDir", err)
		}

		pkgs, err := parser.ParseDir(fset, pkg.Dir, nil, 0)
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			log.Fatalln(err)
		}

		for _, astPkg := range pkgs {
			packages = append(packages, astPkg)
		}
	}

	decls := FilterDeclarations(filterType(typeName), packages...)

	switch len(decls) {
	case 1:
	// happy case, fallthrough
	case 0:
		log.Fatalln("Failed to locate", packageName, typeName)
	default:
		log.Fatalln("Ambiguous type, located multiple matches", decls)
	}

	typeDecl := typeDeclaration2{
		TypeIdent:   typeName,
		PackageName: decls[0].Package.Name,
		Fields:      ExtractFields(decls[0].GenDecl),
	}

	aliaser := genieql.AliaserBuilder(strategyArray...)
	if aliaser == nil {
		log.Fatalln("unknown alias strategy")
	}

	f, errors := typeDecl.Aliases(aliaser, genieql.AliasStrategyTablePrefix(tableName, aliaser))
	if len(errors) != 0 {
		log.Fatalln(errors)
	}
	m := genieql.Mapping{
		Type:   typeDecl.TypeString(),
		Fields: f,
	}

	d, err := yaml.Marshal(&m)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))
}

// ASTFieldToGenieqlField Number of potential paths ast.SelectorExpr, ast.Ident
func ASTFieldToGenieqlField(field *ast.Field, aliasers ...genieql.Aliaser) (result genieql.Field, err error) {
	if field.Names == nil {
		return result, fmt.Errorf("Anonymous types not supported")
	}

	var name string
	var fieldType string
	var aliases []string

	switch t := field.Type.(type) {
	case *ast.Ident:
		fieldType = t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			fieldType = ident.Name + "." + t.Sel.Name
		} else {
			return result, fmt.Errorf("unsupported *ast.SelectorExpr, Expression was not an *ast.Ident, if you get this error please open an issue")
		}
	default:
		return result, fmt.Errorf("unsupported ast.Node")
	}
	name = field.Names[0].Name
	aliases = genieql.MultiAliaser(name, aliasers...)

	result = genieql.Field{
		Type:        fieldType,
		StructField: name,
		Aliases:     aliases,
	}

	return result, nil
}

func ExtractFields(decl *ast.GenDecl) (list *ast.FieldList) {
	ast.Inspect(decl, func(n ast.Node) bool {
		if fields, ok := n.(*ast.FieldList); ok {
			list = fields
			return false
		}
		return true
	})
	return
}

// FilterDeclarations filter out any type declarations that do not match the filter.
func FilterDeclarations(f ast.Filter, packageSet ...*ast.Package) []typeDeclaration {
	results := []typeDeclaration{}
	for _, pkg := range packageSet {
		ast.Inspect(pkg, func(n ast.Node) bool {
			decl, ok := n.(*ast.GenDecl)
			if ok && ast.FilterDecl(decl, f) {
				results = append(results, typeDeclaration{Package: pkg, GenDecl: decl})
			}
			return true
		})
	}
	return results
}

func filterType(typeName string) ast.Filter {
	return func(in string) bool {
		return typeName == in
	}
}
