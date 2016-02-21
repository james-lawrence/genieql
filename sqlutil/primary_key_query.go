package sqlutil

import "fmt"

var primaryKeyQuery = map[string]string{}

// RegisterPrimaryKeyQuery registers the primary key query for the given dialect.
// for use by code generators.
func RegisterPrimaryKeyQuery(dialect, query string) error {
	if _, exists := primaryKeyQuery[dialect]; exists {
		return fmt.Errorf("primary key query is already registered for dialect %s", dialect)
	}

	primaryKeyQuery[dialect] = query

	return nil
}

// LookupPrimaryKeyQuery returns the primary key query, if any, for the given dialect.
func LookupPrimaryKeyQuery(dialect, table string) (string, error) {
	tmpl, exists := primaryKeyQuery[dialect]
	if !exists {
		return "", fmt.Errorf("no primary key query implemented for dialect %s", dialect)
	}

	return fmt.Sprintf(tmpl, table), nil
}
