package genieql

import (
	"fmt"
	"go/ast"
	"go/parser"
	"sort"

	"bitbucket.org/jatone/genieql/x/stringsx"
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

type lesser func(i, j ColumnInfo) bool

// SortColumnInfo ...
func SortColumnInfo(input []ColumnInfo) func(c lesser) []ColumnInfo {
	return func(c lesser) []ColumnInfo {
		sort.Sort(sortableColumnInfo{columns: input, lesser: c})
		return input
	}
}

type sortableColumnInfo struct {
	lesser  lesser
	columns []ColumnInfo
}

func (t sortableColumnInfo) Len() int {
	return len(t.columns)
}

func (t sortableColumnInfo) Swap(i, j int) {
	t.columns[i], t.columns[j] = t.columns[j], t.columns[i]
}

func (t sortableColumnInfo) Less(i, j int) bool {
	return t.lesser(t.columns[i], t.columns[j])
}

func ByName(i, j ColumnInfo) bool {
	return i.Name < j.Name
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

func NewColumnInfoNameTransformer(aliasers ...Aliaser) ColumnInfoNameTransformer {
	return ColumnInfoNameTransformer{Aliaser: AliaserChain(aliasers...)}
}

type ColumnInfoNameTransformer struct {
	Aliaser
}

func (t ColumnInfoNameTransformer) Transform(column ColumnInfo) string {
	return t.Aliaser.Alias(column.Name)
}

type ColumnValueTransformer struct {
	Defaults           []string
	DialectTransformer ColumnTransformer
}

func (t ColumnValueTransformer) Transform(column ColumnInfo) string {
	const defaultValue = "DEFAULT"
	if stringsx.Contains(column.Name, t.Defaults...) {
		return defaultValue
	}
	return t.DialectTransformer.Transform(column)
}

// ColumnTransformer transforms a ColumnInfo into a string for the constant.
type ColumnTransformer interface {
	Transform(ColumnInfo) string
}

// TableDetails provides information about the table.
type TableDetails struct {
	Dialect         Dialect
	Table           string
	Naturalkey      []ColumnInfo
	Columns         []ColumnInfo
	UnmappedColumns []ColumnInfo
}

// OnlyMappedColumns filters out columns from the current TableDetails that do not
// exist in the destination structure. Mainly used for generating queries.
func (t TableDetails) OnlyMappedColumns(fields []*ast.Field, aliases ...Aliaser) TableDetails {
	dup := t

	if len(fields) == 0 {
		dup.Columns = []ColumnInfo{}
		dup.UnmappedColumns = append(dup.UnmappedColumns, t.Columns...)
		return dup
	}

	dup.Columns = make([]ColumnInfo, 0, len(t.Columns))
	dup.UnmappedColumns = make([]ColumnInfo, 0, len(t.Columns))

	for _, column := range t.Columns {
		var mapped bool
		for _, field := range fields {
			if matched, _ := MapFieldToColumn(column.Name, 0, field, aliases...); matched {
				mapped = true
				dup.Columns = append(dup.Columns, column)
			}
		}
		if !mapped {
			dup.UnmappedColumns = append(dup.UnmappedColumns, column)
		}
	}

	return dup
}

// LookupTableDetails determines the table details for the given dialect.
func LookupTableDetails(dialect Dialect, table string) (TableDetails, error) {
	var (
		err        error
		naturalKey []ColumnInfo
		columns    []ColumnInfo
	)

	if columns, err = dialect.ColumnInformationForTable(table); err != nil {
		return TableDetails{}, err
	}

	for _, column := range columns {
		if column.PrimaryKey {
			naturalKey = append(naturalKey, column)
		}
	}

	return TableDetails{
		Dialect:    dialect,
		Table:      table,
		Naturalkey: naturalKey,
		Columns:    columns,
	}, nil
}
