package genieql

import (
	"database/sql"
	"fmt"
)

var dialectMap = map[string]Dialect{}

func RegisterDialect(dialect string, imp Dialect) error {
	if _, exists := dialectMap[dialect]; exists {
		return fmt.Errorf("column query is already registered for dialect %s", dialect)
	}

	dialectMap[dialect] = imp

	return nil
}

func LookupDialect(dialect string) (Dialect, error) {
	impl, exists := dialectMap[dialect]
	if !exists {
		return nil, fmt.Errorf("no implementation for dialect %s", dialect)
	}

	return impl, nil
}

// Dialect ...
type Dialect interface {
	Insert(table string, columns, defaults []string) string
	Select(table string, columns, predicates []string) string
	Update(table string, columns, predicates []string) string
	Delete(table string, columns, predicates []string) string
	ColumnQuery(table string) string
	PrimaryKeyQuery(table string) string
}

// LookupTableDetails determines the table details for the given dialect.
func LookupTableDetails(db *sql.DB, dialect Dialect, table string) (TableDetails, error) {
	var err error
	var columns []string
	var naturalKey []string

	if columns, err = Columns(db, dialect.ColumnQuery(table)); err != nil {
		return TableDetails{}, err
	}

	if naturalKey, err = ExtractPrimaryKey(db, dialect.PrimaryKeyQuery(table)); err != nil {
		return TableDetails{}, err
	}

	return TableDetails{
		Dialect:    dialect,
		Table:      table,
		Naturalkey: naturalKey,
		Columns:    columns,
	}, nil
}
