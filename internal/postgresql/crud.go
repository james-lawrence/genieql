package postgresql

import (
	"fmt"
	"strings"
)

type CRUD struct{}

func (t CRUD) InsertQuery(table string, columns []string) string {
	placeholders := make([]string, 0, len(columns))
	for _ = range columns {
		placeholders = append(placeholders, "?")
	}
	return fmt.Sprintf(insertTmpl, table, strings.Join(columns, ","), strings.Join(placeholders, ","))
}

func (t CRUD) SelectQuery(table string, predicates []string) string {
	clause := strings.Join(predicates, " = ? AND ") + " = ?"
	return fmt.Sprintf(selectByFieldTmpl, table, clause)
}

func (t CRUD) UpdateQuery(table string, columns, predicates []string) string {
	clause := strings.Join(predicates, " = ? AND ") + " = ?"
	update := strings.Join(columns, " = ?, ") + " = ?"
	return fmt.Sprintf(updateTmpl, table, update, clause)
}

func (t CRUD) DeleteQuery(table string, predicates []string) string {
	clause := strings.Join(predicates, " = ? AND ") + " = ?"
	return fmt.Sprintf(deleteTmpl, table, clause)
}

const selectByFieldTmpl = "SELECT * FROM %s WHERE %s"
const insertTmpl = "INSERT INTO %s (%s) VALUES (%s) RETURNING *"
const updateTmpl = "UPDATE %s SET (%s) WHERE %s RETURNING *"
const deleteTmpl = "DELETE FROM %s WHERE %s"
