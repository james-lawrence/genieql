package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"bitbucket.org/jatone/genieql"

	_ "github.com/lib/pq"
)

type database struct {
	Driver         string
	User           string
	Name           string
	SSLMode        string
	Port           int
	MaxConnections int
}

type Identity struct {
	ID      string
	Email   string
	Created time.Time
	Updated time.Time
}

func main() {
	dbconf := database{
		Driver:         "postgres",
		User:           "jatone",
		Name:           "sso",
		SSLMode:        "disable",
		Port:           5432,
		MaxConnections: 10,
	}

	d, err := sql.Open(dbconf.Driver, fmt.Sprintf("user=%s dbname=%s sslmode=%s port=%d", dbconf.User, dbconf.Name, dbconf.SSLMode, dbconf.Port))
	if err != nil {
		log.Fatal(err)
	}

	mapper := Mapper{
		Type: reflect.TypeOf(Identity{}),
		Aliases: []Alias{
			{StructField: "ID", Aliases: []string{"id", "identity_id"}},
			{StructField: "Email", Aliases: []string{"email", "identity_email"}},
			{StructField: "Created", Aliases: []string{"created", "identity_created"}},
			{StructField: "Updated", Aliases: []string{"updated", "identity_updated"}},
		},
	}

	printIfErr(ScannerBuilder([]Mapper{mapper}).BuildScanner(d, "SELECT * FROM identity"))
	printIfErr(AmbiguityCheckFromQuery(d, "SELECT * FROM identity"))
	printIfErr(AmbiguityCheckFromQuery(d, "SELECT * FROM identity JOIN identity_google ON identity.id = identity_google.identity_id JOIN identity_username ON identity.id = identity_username.identity_id"))
}

func AmbiguityCheckFromQuery(db *sql.DB, query string, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	return genieql.AmbiguityCheck(columns...)
}

type Mapper struct {
	Type    reflect.Type
	Aliases []Alias
}

type Alias struct {
	StructField string
	Aliases     []string
}

func (t Alias) IsAlias(s string) bool {
	for _, alias := range t.Aliases {
		if alias == s {
			return true
		}
	}

	return false
}

type Match struct {
	Type        reflect.Type
	StructField string
	Position    int
}

type ScannerBuilder []Mapper

func (t ScannerBuilder) BuildScanner(db *sql.DB, query string, args ...interface{}) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	if err := genieql.AmbiguityCheck(columns...); err != nil {
		return err
	}

	fmt.Println("Columns:", columns)
	types := []reflect.Type{}
	matches := make([]Match, 0, len(columns))

	for idx, column := range columns {
	nextColumn:
		for _, mapper := range t {
			for _, alias := range mapper.Aliases {
				// if isNotAmbiguous(column) && alias.IsAlias(column) {
				if alias.IsAlias(column) {
					matches = append(matches, Match{
						Type:        mapper.Type,
						StructField: alias.StructField,
						Position:    idx,
					})
					fmt.Printf("%s is an alias %s.%s: position %d\n", column, mapper.Type.Name(), alias.StructField, idx)
					continue nextColumn
				}
			}
		}
	}

	for _, ttype := range types {
		fmt.Printf("var %s %s.%s\n", strings.ToLower(ttype.Name()), ttype.PkgPath(), ttype.Name())
	}

	localVariables := []string{}
	for _, match := range matches {
		if v, ok := match.Type.FieldByName(match.StructField); ok {
			typeName := typeToName(v.Type)
			localVariable := localVariable("c", match.Position)
			fmt.Printf("var %s %s\n", localVariable, typeName)
			localVariables = append(localVariables, localVariable)
		} else {
			panic(fmt.Sprintf("field %s not found on %s.%s", match.StructField, match.Type.PkgPath(), match.Type.Name()))
		}
	}

	fmt.Println("rows.Scan(&" + strings.Join(localVariables, ", &") + ")")

	for _, mapper := range t {
		fmt.Printf("var %s %s\n", strings.ToLower(mapper.Type.Name()), typeToName(mapper.Type))
	}

	for _, match := range matches {
		fmt.Printf("%s.%s = %s\n", strings.ToLower(match.Type.Name()), match.StructField, localVariable("c", match.Position))
	}

	return nil
}

func typeToName(t reflect.Type) (typeName string) {
	typeName = t.PkgPath()
	if len(typeName) == 0 {
		typeName = t.Name()
	} else {
		typeName = typeName + "." + t.Name()
	}

	return
}

func localVariable(prefix string, position int) string {
	return fmt.Sprintf("%s%d", prefix, position)
}

func printIfErr(err error) {
	if err != nil {
		log.Println(err)
	} else {
		log.Println("No Issues Detected")
	}
}

type IdentityScanner interface {
	Scan(*Identity) error
	// Uses cap([]Identity) to determine how many items to scan.
	ScanN([]Identity) error
	ScanAll([]Identity) error
}

func NewIdentityQueryScanner(r *sql.Rows, err error) IdentityScanner {
	if err != nil {
		return errIdentityScanner{err}
	}

	return identityScanner{}
}

type errIdentityScanner struct {
	err error
}

func (t errIdentityScanner) Scan(_ *Identity) error {
	return t.err
}

func (t errIdentityScanner) ScanN(_ []Identity) error {
	return t.err
}

func (t errIdentityScanner) ScanAll(_ []Identity) error {
	return t.err
}

type identityScanner struct {
}

func (t identityScanner) Scan(identity *Identity) error {
	return fmt.Errorf("Not Implemented")
}

func (t identityScanner) ScanN(_ []Identity) error {
	return fmt.Errorf("Not Implemented")
}

func (t identityScanner) ScanAll(_ []Identity) error {
	return fmt.Errorf("Not Implemented")
}

// func queryAndPrint(db *sql.DB, q string) {
// 	rows, err := db.Query(q)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
//
// 	log.Println(rows.Columns())
// }
//
// func scanAndPrint(db *sql.DB, q string) {
// 	rows, err := db.Query(q)
// 	if err != nil {
// 		log.Fatalln(err)
// 	}
// 	results := []Identity{}
// 	scanner := NewIdentityScanner(rows.Columns())
// 	log.Println("Scan All Error", scanner.ScanAll(rows, &results))
// 	for _, ident := range results {
// 		log.Println(ident)
// 	}
// }
