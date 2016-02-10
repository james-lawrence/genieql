package genieql

import (
	"fmt"
	"go/ast"
)

type MapperV2 struct {
	Aliasers []Aliaser
}

func (t MapperV2) MapColumns(argname *ast.Ident, fields []*ast.Field, columns ...string) ([]ColumnMap, error) {
	matches := make([]ColumnMap, 0, len(columns))
	for idx, column := range columns {
		for _, field := range fields {
			m, matched, err := MapFieldToColumn(argname, column, idx, field, t.Aliasers...)
			if err != nil {
				return matches, err
			}

			if matched {
				matches = append(matches, m)
				break
			}
		}
	}

	return matches, nil
}

func MapFieldToColumn(argname *ast.Ident, column string, colIdx int, field *ast.Field, aliases ...Aliaser) (ColumnMap, bool, error) {
	if len(field.Names) != 1 {
		return ColumnMap{}, false, fmt.Errorf("field had more than 1 name")
	}

	fieldName := field.Names[0].Name
	for _, aliaser := range aliases {
		if column == aliaser.Alias(fieldName) {
			return ColumnMap{
				Column: &ast.Ident{
					Name: fmt.Sprintf("c%d", colIdx),
				},
				Type: field.Type,
				Assignment: &ast.SelectorExpr{
					X: argname,
					Sel: &ast.Ident{
						Name: fieldName,
					},
				},
			}, true, nil
		}
	}

	return ColumnMap{}, false, nil
}
