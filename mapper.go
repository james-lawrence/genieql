package genieql

import (
	"fmt"
	"go/ast"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type MappingConfig struct {
	Package              string
	Type                 string
	Table                string
	IncludeTablePrefixes bool
	NaturalKey           []string
	Transformations      []string
}

func WriteMapper(root string, configuration Configuration, m MappingConfig) error {
	d, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	path := filepath.Join(root, configuration.Database, m.Package, m.Type, m.Table)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, d, 0666)
}

func ReadMapper(root, pkg, typ, table string, configuration Configuration, m *MappingConfig) error {
	raw, err := ioutil.ReadFile(filepath.Join(root, configuration.Database, pkg, typ, table))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, m)
}

func Map(configFile string, m MappingConfig) error {
	var config Configuration

	if err := ReadConfiguration(configFile, &config); err != nil {
		return err
	}

	return WriteMapper(filepath.Dir(configFile), config, m)
}

type Mapper struct {
	Aliasers []Aliaser
}

func (t Mapper) MapColumns(argname *ast.Ident, fields []*ast.Field, columns ...string) ([]ColumnMap, error) {
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
