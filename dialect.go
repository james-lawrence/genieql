package genieql

import (
	"database/sql"
	"fmt"
)

// ErrMissingDialect - returned when a dialect has not been registered.
var ErrMissingDialect = fmt.Errorf("requested dialect is not registered")

// ErrDuplicateDialect - returned when a dialect gets registered twice.
var ErrDuplicateDialect = fmt.Errorf("dialect has already been registered")

var dialects = dialectRegistry{}

// RegisterDialect register a sql dialect with genieql. usually in an init function.
func RegisterDialect(dialect string, imp Dialect) error {
	return dialects.RegisterDialect(dialect, imp)
}

// LookupDialect lookup a registered dialect.
func LookupDialect(dialect string) (Dialect, error) {
	return dialects.LookupDialect(dialect)
}

// Dialect ...
type Dialect interface {
	Insert(table string, columns, defaults []string) string
	Select(table string, columns, predicates []string) string
	Update(table string, columns, predicates []string) string
	Delete(table string, columns, predicates []string) string
	ColumnInformation(db *sql.DB, table string) ([]ColumnInfo, error)
	ColumnInformationForQuery(db *sql.DB, query string) ([]ColumnInfo, error)
}

type dialectRegistry map[string]Dialect

func (t dialectRegistry) RegisterDialect(dialect string, imp Dialect) error {
	if _, exists := t[dialect]; exists {
		return ErrDuplicateDialect
	}

	t[dialect] = imp

	return nil
}

func (t dialectRegistry) LookupDialect(dialect string) (Dialect, error) {
	impl, exists := t[dialect]
	if !exists {
		return nil, ErrMissingDialect
	}

	return impl, nil
}
