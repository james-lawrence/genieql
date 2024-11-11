package sqlite3

import (
	"fmt"
	"strings"

	"github.com/james-lawrence/genieql/internal/stringsx"
)

// Insert generate an insert query.
func Insert(n int, offset int, table, conflict string, columns, defaulted []string) string {
	const (
		insertTmpl = "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES (:gql.insert.values:):gql.insert.conflict:"
	)
	offset = offset + 1
	p, _ := placeholders(offset, selectPlaceholder(columns, defaulted))
	values := strings.Join(p, ",")
	columnOrder := strings.Join(columns, ",")

	replacements := strings.NewReplacer(
		":gql.insert.tablename:", table,
		":gql.insert.columns:", columnOrder,
		":gql.insert.values:", values,
		":gql.insert.conflict:", stringsx.DefaultIfBlank(" "+conflict, ""),
		":gql.insert.returning:", columnOrder,
	)

	return replacements.Replace(insertTmpl)
}

// Select generate a select query.
func Select(table string, columns, predicates []string) string {
	clauses, _ := predicate(1, predicates...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(selectByFieldTmpl, columnOrder, table, strings.Join(clauses, " AND "))
}

// Update generate an update query.
func Update(table string, columns, predicates []string) string {
	updates, offset := predicate(1, columns...)
	clauses, _ := predicate(offset, predicates...)
	return fmt.Sprintf(updateTmpl, table, strings.Join(updates, ", "), strings.Join(clauses, " AND "))
}

// Delete generate a delete query.
func Delete(table string, columns, predicates []string) string {
	clauses, _ := predicate(1, predicates...)
	return fmt.Sprintf(deleteTmpl, table, strings.Join(clauses, " AND "))
}

func predicate(offset int, predicates ...string) ([]string, int) {
	clauses := make([]string, 0, len(predicates))
	for idx, predicate := range predicates {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", predicate, offset+idx))
	}

	if len(clauses) == 0 {
		clauses = append(clauses, matchAllClause)
	}

	return clauses, len(predicates) + 1
}

func placeholders(offset int, columns []placeholder) ([]string, int) {
	clauses := make([]string, 0, len(columns))
	idx := offset
	for _, column := range columns {
		var ph string
		ph, idx = column.String(idx)
		clauses = append(clauses, ph)
	}

	return clauses, len(clauses)
}

func selectPlaceholder(columns, defaults []string) []placeholder {
	placeholders := make([]placeholder, 0, len(columns))
	for _, column := range columns {
		var placeholder placeholder = offsetPlaceholder{}
		// todo turn into a set.
		for _, cut := range defaults {
			if cut == column {
				placeholder = defaultPlaceholder{}
				break
			}
		}
		placeholders = append(placeholders, placeholder)
	}

	return placeholders
}

type placeholder interface {
	String(offset int) (string, int)
}

type defaultPlaceholder struct{}

func (t defaultPlaceholder) String(offset int) (string, int) {
	return "DEFAULT", offset
}

type offsetPlaceholder struct{}

func (t offsetPlaceholder) String(offset int) (string, int) {
	return fmt.Sprintf("$%d", offset), offset + 1
}

const selectByFieldTmpl = "SELECT %s FROM %s WHERE %s"
const updateTmpl = "UPDATE %s SET %s WHERE %s"
const deleteTmpl = "DELETE FROM %s WHERE %s"
const matchAllClause = "'t'"
