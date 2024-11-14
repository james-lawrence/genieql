package duckdb

import (
	"fmt"
	"strings"

	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/stringsx"
)

type columnValueTransformer struct {
	offset int
}

func (t *columnValueTransformer) Transform(c genieql.ColumnInfo) string {
	t.offset++
	return fmt.Sprintf("?%d", t.offset)
}

// Insert generates an insert query for DuckDB.
func Insert(n int, offset int, table, conflict string, columns, projection, defaulted []string) string {
	const insertTmpl = "INSERT INTO :gql.insert.tablename: (:gql.insert.columns:) VALUES :gql.insert.values::gql.insert.conflict: RETURNING :gql.insert.returning:"

	columnOrder := strings.Join(quotedColumns(projection...), ",")
	insertions := strings.Join(quotedColumns(columns...), ",")
	offset++
	values := make([]string, 0, n)
	for i := 0; i < n; i++ {
		p, newOffset := placeholders(offset, selectPlaceholder(columns, defaulted))
		offset = newOffset
		values = append(values, fmt.Sprintf("(%s)", strings.Join(p, ",")))
	}

	replacements := strings.NewReplacer(
		":gql.insert.tablename:", quotedString(table),
		":gql.insert.columns:", insertions,
		":gql.insert.values:", strings.Join(values, ","),
		":gql.insert.conflict:", stringsx.DefaultIfBlank(" "+conflict, ""),
		":gql.insert.returning:", columnOrder,
	)

	return replacements.Replace(insertTmpl)
}

// Update generates an update query.
func Update(table string, columns, predicates, returning []string) string {
	const updateTmpl = "UPDATE `%s` SET %s WHERE %s RETURNING %s"
	updates, offset := predicate(1, columns...)
	clauses, _ := predicate(offset, predicates...)
	return fmt.Sprintf(updateTmpl, table, strings.Join(updates, ", "), strings.Join(clauses, " AND "),
		strings.Join(quotedColumns(returning...), ","))
}

// Select generates a select query.
func Select(table string, columns, predicates []string) string {
	clauses, _ := predicate(1, predicates...)
	columnOrder := strings.Join(quotedColumns(columns...), ",")
	return fmt.Sprintf(selectByFieldTmpl, columnOrder, table, strings.Join(clauses, " AND "))
}

// Delete generates a delete query.
func Delete(table string, columns, predicates []string) string {
	clauses, _ := predicate(1, predicates...)
	return fmt.Sprintf(deleteTmpl, table, strings.Join(clauses, " AND "))
}

// predicate formats WHERE clauses with placeholders.
func predicate(offset int, predicates ...string) ([]string, int) {
	clauses := make([]string, 0, len(predicates))
	for idx, predicate := range quotedColumns(predicates...) {
		clauses = append(clauses, fmt.Sprintf("%s = ?%d", predicate, offset+idx))
	}

	if len(clauses) == 0 {
		clauses = append(clauses, matchAllClause)
	}

	return clauses, offset + len(predicates)
}

// placeholders formats values with positional parameters.
func placeholders(offset int, columns []placeholder) ([]string, int) {
	clauses := make([]string, 0, len(columns))
	for _, column := range columns {
		var ph string
		ph, offset = column.String(offset)
		clauses = append(clauses, ph)
	}

	return clauses, offset
}

// selectPlaceholder builds a list of placeholders for SELECT statements.
func selectPlaceholder(columns, defaults []string) []placeholder {
	placeholders := make([]placeholder, len(columns))
	defaulted := make(map[string]struct{}, len(defaults))
	for _, d := range defaults {
		defaulted[d] = struct{}{}
	}

	for i, column := range columns {
		if _, ok := defaulted[column]; ok {
			placeholders[i] = defaultPlaceholder{}
		} else {
			placeholders[i] = offsetPlaceholder{}
		}
	}

	return placeholders
}

func quotedString(s string) string {
	return fmt.Sprintf("`%s`", s)
}

func quotedColumns(columns ...string) []string {
	results := make([]string, len(columns))
	for i, c := range columns {
		results[i] = quotedString(c)
	}
	return results
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
	return fmt.Sprintf("?%d", offset), offset + 1
}

const selectByFieldTmpl = "SELECT %s FROM %s WHERE %s"
const updateTmpl = "UPDATE %s SET %s WHERE %s RETURNING %s"
const deleteTmpl = "DELETE FROM %s WHERE %s"
const matchAllClause = "TRUE"
