package postgresql

import (
	"fmt"
	"strings"
)

// Insert generate an insert query.
func Insert(table string, columns, defaulted []string) string {
	p, offset := placeholders(1, columns...)
	d, _ := defaults(offset, defaulted...)
	values := strings.Join(append(p, d...), ",")
	columnOrder := strings.Join(append(columns, defaulted...), ",")
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

func placeholders(offset int, columns ...string) ([]string, int) {
	clauses := make([]string, 0, len(columns))
	for idx := range columns {
		clauses = append(clauses, fmt.Sprintf("$%d", offset+idx))
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

const selectByFieldTmpl = "SELECT %s FROM %s WHERE %s"
const insertTmpl = "INSERT INTO %s (%s) VALUES (%s) RETURNING %s"
const updateTmpl = "UPDATE %s SET (%s) WHERE %s RETURNING %s"
const deleteTmpl = "DELETE FROM %s WHERE %s RETURNING %s"
