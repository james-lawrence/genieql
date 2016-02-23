package postgresql

import (
	"fmt"
	"strings"
)

type CRUD struct{}

func (t CRUD) InsertQuery(table string, columns []string) string {
	p, _ := placeholders(1, ",", columns...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(insertTmpl, table, columnOrder, p, columnOrder)
}

func (t CRUD) SelectQuery(table string, columns, predicates []string) string {
	clause, _ := predicate(1, " AND ", predicates...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(selectByFieldTmpl, columnOrder, table, clause)
}

func (t CRUD) UpdateQuery(table string, columns, predicates []string) string {
	offset := 1
	update, offset := predicate(offset, ", ", columns...)
	clause, _ := predicate(offset, " AND ", predicates...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(updateTmpl, table, update, clause, columnOrder)
}

func (t CRUD) DeleteQuery(table string, columns, predicates []string) string {
	clause, _ := predicate(1, " AND ", predicates...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(deleteTmpl, table, clause, columnOrder)
}

func predicate(offset int, join string, predicates ...string) (string, int) {
	clauses := make([]string, 0, len(predicates))
	for idx, predicate := range predicates {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", predicate, offset+idx))
	}

	return strings.Join(clauses, join), len(clauses) + 1
}

func placeholders(offset int, join string, columns ...string) (string, int) {
	clauses := make([]string, 0, len(columns))
	for idx := range columns {
		clauses = append(clauses, fmt.Sprintf("$%d", offset+idx))
	}

	return strings.Join(clauses, join), len(clauses)
}

const selectByFieldTmpl = "SELECT %s FROM %s WHERE %s"
const insertTmpl = "INSERT INTO %s (%s) VALUES (%s) RETURNING %s"
const updateTmpl = "UPDATE %s SET (%s) WHERE %s RETURNING %s"
const deleteTmpl = "DELETE FROM %s WHERE %s RETURNING %s"
