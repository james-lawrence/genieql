package postgresql

import (
	"fmt"
	"strings"
)

type CRUD struct{}

func (t CRUD) InsertQuery(table string, columns []string) string {
	p, _ := placeholders(0, ",", columns...)
	columnOrder := strings.Join(columns, ",")
	return fmt.Sprintf(insertTmpl, table, columnOrder, p)
}

func (t CRUD) SelectQuery(table string, predicates []string) string {
	clause, _ := predicate(0, " AND ", predicates...)
	return fmt.Sprintf(selectByFieldTmpl, table, clause)
}

func (t CRUD) UpdateQuery(table string, columns, predicates []string) string {
	offset := 0
	update, offset := predicate(offset, ", ", columns...)
	clause, _ := predicate(offset, " AND ", predicates...)
	return fmt.Sprintf(updateTmpl, table, update, clause)
}

func (t CRUD) DeleteQuery(table string, predicates []string) string {
	clause, _ := predicate(0, " AND ", predicates...)
	return fmt.Sprintf(deleteTmpl, table, clause)
}

func predicate(offset int, join string, predicates ...string) (string, int) {
	clauses := make([]string, 0, len(predicates))
	for idx, predicate := range predicates {
		clauses = append(clauses, fmt.Sprintf("%s = $%d", predicate, offset+idx))
	}

	return strings.Join(clauses, join), len(clauses)
}

func placeholders(offset int, join string, columns ...string) (string, int) {
	clauses := make([]string, 0, len(columns))
	for idx := range columns {
		clauses = append(clauses, fmt.Sprintf("$%d", offset+idx))
	}

	return strings.Join(clauses, join), len(clauses)
}

const selectByFieldTmpl = "SELECT * FROM %s WHERE %s"
const insertTmpl = "INSERT INTO %s (%s) VALUES (%s) RETURNING *"
const updateTmpl = "UPDATE %s SET (%s) WHERE %s RETURNING *"
const deleteTmpl = "DELETE FROM %s WHERE %s RETURNING *"
