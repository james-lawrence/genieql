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
	return generatorFunc(func(dst io.Writer) error {
		names := genieql.ColumnInfoSet(t.Columns).ColumnNames()
		query := t.Dialect.Insert(t.Table, names, defaults)
		return emit(dst, name, query)
	})
}

type Select genieql.TableDetails

func (t Select) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer) error {
		names := genieql.ColumnInfoSet(t.Columns).ColumnNames()
		query := t.Dialect.Select(t.Table, names, predicates)
		return emit(dst, name, query)
	})
}

type Update genieql.TableDetails

func (t Update) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer) error {
		names := genieql.ColumnInfoSet(t.Columns).ColumnNames()
		query := t.Dialect.Update(t.Table, names, predicates)
		return emit(dst, name, query)
	})
}

type Delete genieql.TableDetails

func (t Delete) Build(name string, predicates []string) genieql.Generator {
	return generatorFunc(func(dst io.Writer) error {
		names := genieql.ColumnInfoSet(t.Columns).ColumnNames()
		query := t.Dialect.Delete(t.Table, names, predicates)
		return emit(dst, name, query)
	})
}

type generatorFunc func(dst io.Writer) error

func (t generatorFunc) Generate(dst io.Writer) error {
	return t(dst)
}

func emit(dst io.Writer, constName, query string) error {
	if err := format.Node(dst, token.NewFileSet(), genieql.QueryLiteral(constName, query)); err != nil {
		return err
	}
	_, err := fmt.Fprintf(dst, "\n")
	return err
}
