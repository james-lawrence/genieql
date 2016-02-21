package genieql

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
)

// ConnectDB connects to a database based on the configuration.
func ConnectDB(config Configuration) (*sql.DB, error) {
	log.Printf("connection %s\n", config.ConnectionURL)
	return sql.Open(config.Dialect, config.ConnectionURL)
}

// Columns convience function, executes ExtractColumns followed by AmbiguityCheck.
func Columns(db *sql.DB, query string, args ...interface{}) ([]string, error) {
	var columns []string
	var err error
	if columns, err = ExtractColumns(db, query); err != nil {
		return columns, err
	}

	return columns, AmbiguityCheck(columns...)
}

// ExtractColumns executes a query and extracts the resulting set of columns from
// the database.
func ExtractColumns(db *sql.DB, query string, args ...interface{}) (columns []string, err error) {
	var rows *sql.Rows
	rows, err = db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()

	columns, err = rows.Columns()
	return
}

// ExtractPrimaryKey executes the query to determine the primary keys of a table.
func ExtractPrimaryKey(db *sql.DB, query string, args ...interface{}) ([]string, error) {
	var rows *sql.Rows
	var columns []string
	var err error

	rows, err = db.Query(query, args...)
	if err != nil {
		return columns, err
	}
	defer rows.Close()

	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			return columns, err
		}
		columns = append(columns, column)
	}

	return columns, nil
}

// AmbiguityCheck checks the provided columns for duplicated values.
func AmbiguityCheck(columns ...string) error {
	sort.Strings(columns)

	ambiguousColumns := []string{}

	if len(columns) > 0 {
		previous, tail := columns[0], columns[1:]
		lastMatch := ""
		for _, current := range tail {
			if previous == current && lastMatch != current {
				ambiguousColumns = append(ambiguousColumns, current)
				lastMatch = current
			}
			previous = current
		}
	}

	if len(ambiguousColumns) > 0 {
		return fmt.Errorf("ambiguous columns in results %v", ambiguousColumns)
	}

	return nil
}
