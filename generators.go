package genieql

import (
	"fmt"
	"go/format"
	"go/token"
	"io"

	"bitbucket.org/jatone/genieql/internal/postgresql"

	"github.com/serenize/snaker"
)

var PostgresqlCRUDGen = postgresql.CRUD{}

type Generator interface {
	Generate() (io.Reader, error)
}

type CrudWriter interface {
	Write(dialect CrudGenerator, fset *token.FileSet) error
}

type CrudGenerator interface {
	InsertQuery(table string, columns []string) string
	SelectQuery(table string, columns, predicates []string) string
	UpdateQuery(table string, columns, predicates []string) string
	DeleteQuery(table string, columns, predicates []string) string
}

func NewCRUDWriter(out io.Writer, prefix, table string, naturalkey []string, columns []string) CrudWriter {
	return crudWriter{
		out:        out,
		prefix:     prefix,
		table:      table,
		naturalkey: naturalkey,
		columns:    columns,
	}
}

type crudWriter struct {
	out        io.Writer
	prefix     string
	table      string
	naturalkey []string
	columns    []string
}

func (t crudWriter) Write(dialect CrudGenerator, fset *token.FileSet) error {
	constName := fmt.Sprintf("%sInsert", t.prefix)
	query := dialect.InsertQuery(t.table, t.columns)
	if err := format.Node(t.out, fset, QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	for i, column := range t.columns {
		constName := fmt.Sprintf("%sFindBy%s", t.prefix, snaker.SnakeToCamel(column))
		query := dialect.SelectQuery(t.table, t.columns, t.columns[i:i+1])
		if err := format.Node(t.out, fset, QueryLiteral(constName, query)); err != nil {
			return err
		}
		fmt.Fprintf(t.out, "\n")
	}

	constName = fmt.Sprintf("%sUpdateByID", t.prefix)
	query = dialect.UpdateQuery(t.table, t.columns, t.naturalkey)
	if err := format.Node(t.out, fset, QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	constName = fmt.Sprintf("%sDeleteByID", t.prefix)
	query = dialect.DeleteQuery(t.table, t.columns, t.naturalkey)
	if err := format.Node(t.out, fset, QueryLiteral(constName, query)); err != nil {
		return err
	}
	fmt.Fprintf(t.out, "\n")

	return nil
}
