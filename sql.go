package genieql

import (
	"fmt"
	"go/ast"
	"go/parser"
	"sort"
)

type ColumnInfo struct {
	Name       string
	Nullable   bool
	PrimaryKey bool
	Type       string
}

func (t ColumnInfo) MapColumn(x ast.Expr) (ColumnMap, error) {
	typ, err := parser.ParseExpr(t.Type)
	return ColumnMap{
		Name:   t.Name,
		Dst:    x,
		PtrDst: t.Nullable,
		Type:   typ,
	}, err
}

type ColumnInfoSet []ColumnInfo

// ColumnNames returns the column names inside the ColumnInfoSet.
func (t ColumnInfoSet) ColumnNames() []string {
	var columns []string

	for _, column := range t {
		columns = append(columns, column.Name)
	}

	return columns
}

// AmbiguityCheck checks the provided columns for duplicated values.
func (t ColumnInfoSet) AmbiguityCheck() error {
	var (
		columnNames = t.ColumnNames()
	)

	sort.Strings(columnNames)

	ambiguousColumns := []string{}

	if len(columnNames) > 0 {
		previous, tail := columnNames[0], columnNames[1:]
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

// LookupTableDetails determines the table details for the given dialect.
func LookupTableDetails(dialect Dialect, table string) (TableDetails, error) {
	var (
		err         error
		columnNames []string
		naturalKey  []string
		columns     []ColumnInfo
	)

	if columns, err = dialect.ColumnInformationForTable(table); err != nil {
		return TableDetails{}, err
	}

	for _, column := range columns {
		columnNames = append(columnNames, column.Name)
		if column.PrimaryKey {
			naturalKey = append(naturalKey, column.Name)
		}
	}

	return TableDetails{
		Dialect:    dialect,
		Table:      table,
		Naturalkey: naturalKey,
		Columns:    columnNames,
	}, nil
}
