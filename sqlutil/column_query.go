package sqlutil

import "fmt"

var columnQuery = map[string]string{}

func RegisterColumnQuery(dialect, query string) error {
	if _, exists := columnQuery[dialect]; exists {
		return fmt.Errorf("column query is already registered for dialect %s", dialect)
	}

	columnQuery[dialect] = query

	return nil
}

func LookupColumnQuery(dialect, table string) (string, error) {
	tmpl, exists := columnQuery[dialect]
	if !exists {
		return "", fmt.Errorf("no column query implemented for dialect %s", dialect)
	}

	return fmt.Sprintf(tmpl, table), nil
}
