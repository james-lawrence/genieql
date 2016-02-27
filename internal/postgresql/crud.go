package postgresql

import (
	"fmt"
	"strings"
)

// Insert generate an insert query.
func Insert(table string, columns, defaulted []string) string {
	p, _ := placeholders(1, selectPlaceholder(columns, defaulted))
	values := strings.Join(p, ",")
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(insertTmpl, table, columnOrder, values, columnOrder)
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
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(updateTmpl, table, strings.Join(updates, ", "), strings.Join(clauses, " AND "), columnOrder)
}

// Delete generate a delete query.
func Delete(table string, columns, predicates []string) string {
	clauses, _ := predicate(1, predicates...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(deleteTmpl, table, strings.Join(clauses, " AND "), columnOrder)
}

func predicate(offset int, predicates ...string) ([]string, int) {
	clauses := make([]string, 0, len(predicates))
	for idx, predicate := range predicates {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", predicate, offset+idx))
	}

	return clauses, len(clauses) + 1
}

func placeholders(offset int, columns []placeholder) ([]string, int) {
	clauses := make([]string, 0, len(columns))
	for idx, column := range columns {
		clauses = append(clauses, column.String(offset+idx))
	}

	return clauses, len(clauses)
}

func defaults(offset int, columns ...string) ([]string, int) {
	clauses := make([]string, 0, len(columns))
	for range columns {
		clauses = append(clauses, "DEFAULT")
	}

	return clauses, len(clauses) + 1
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
	String(offset int) string
}

type defaultPlaceholder struct{}

func (t defaultPlaceholder) String(offset int) string {
	return "DEFAULT"
}

type offsetPlaceholder struct{}

func (t offsetPlaceholder) String(offset int) string {
	return fmt.Sprintf("$%d", offset)
}

const selectByFieldTmpl = "SELECT %s FROM %s WHERE %s"
const insertTmpl = "INSERT INTO %s (%s) VALUES (%s) RETURNING %s"
const updateTmpl = "UPDATE %s SET (%s) WHERE %s RETURNING %s"
const deleteTmpl = "DELETE FROM %s WHERE %s RETURNING %s"
