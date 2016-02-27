package genieql

import (
	"fmt"
)

// Dialect ...
type Dialect interface {
	Insert(table string, columns, defaults []string) string
	Select(table string, columns, predicates []string) string
	Update(table string, columns, predicates []string) string
	Delete(table string, columns, predicates []string) string
	ColumnQuery(table string) string
	PrimaryKeyQuery(table string) string
}

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
