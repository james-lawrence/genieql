package crud

import (
	"fmt"
	"go/format"
	"go/token"
	"io"

	"bitbucket.org/jatone/genieql"
)

// Insert generate an insert query for the table.
type Insert genieql.TableDetails

func (t Insert) Build(name string, defaults []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer, fset *token.FileSet) error {
		query := t.Dialect.Insert(t.Table, t.Columns, defaults)
		return emit(dst, fset, name, query)
	})
}

type Select genieql.TableDetails

func (t Select) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer, fset *token.FileSet) error {
		query := t.Dialect.Select(t.Table, t.Columns, predicates)
		return emit(dst, fset, name, query)
	})
}

type Update genieql.TableDetails

func (t Update) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer, fset *token.FileSet) error {
		query := t.Dialect.Update(t.Table, t.Columns, predicates)
		return emit(dst, fset, name, query)
	})
}

type Delete genieql.TableDetails

func (t Delete) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer, fset *token.FileSet) error {
		query := t.Dialect.Delete(t.Table, t.Columns, predicates)
		return emit(dst, fset, name, query)
	})
}

type generatorFunc func(dst io.Writer, fset *token.FileSet) error

func (t generatorFunc) Generate(dst io.Writer, fset *token.FileSet) error {
	return t(dst, fset)
}

func emit(dst io.Writer, fset *token.FileSet, constName, query string) error {
	if err := format.Node(dst, fset, genieql.QueryLiteral(constName, query)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(dst, "\n")
	return err
}
