package genieql

import (
	"fmt"
	"go/ast"
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// MappingConfig TODO...
type MappingConfig struct {
	Package              string
	Type                 string
	IncludeTablePrefixes bool
	Transformations      []string
}

// Mapper TODO...
func (t MappingConfig) Mapper() Mapper {
	return Mapper{Aliasers: []Aliaser{AliaserBuilder(t.Transformations...)}}
}

// TypeFields TODO...
func (t MappingConfig) TypeFields(context build.Context, filter func(*ast.Package) bool) ([]*ast.Field, error) {
	pkg, err := LocatePackage(t.Package, context, filter)
	if err != nil {
		return nil, err
	}

	typ, err := FindUniqueType(FilterName(t.Type), pkg)
	if err != nil {
		return nil, err
	}

	return ExtractFields(typ).List, nil
}

// WriteMapper persists the structure -> result row mapping to disk.
func WriteMapper(root string, configuration Configuration, name string, m MappingConfig) error {
	d, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	path := filepath.Join(root, configuration.Database, m.Package, m.Type, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, d, 0666)
}

// ReadMapper loads the structure -> result row mapping from disk.
func ReadMapper(root, pkg, typ, name string, configuration Configuration, m *MappingConfig) error {
	raw, err := ioutil.ReadFile(filepath.Join(root, configuration.Database, pkg, typ, name))
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, m)
}

// Map TODO...
func Map(configFile, name string, m MappingConfig) error {
	var config Configuration

	if err := ReadConfiguration(configFile, &config); err != nil {
		return err
	}

	return WriteMapper(filepath.Dir(configFile), config, name, m)
}

// Mapper responsible for mapping a result row to a structure.
type Mapper struct {
	Aliasers []Aliaser
}

// MapColumns TODO...
func (t Mapper) MapColumns(fields []*ast.Field, columns ...string) ([]ColumnMap, error) {
	matches := make([]ColumnMap, 0, len(columns))
	for idx, column := range columns {
		for _, field := range fields {
			m, matched, err := MapFieldToColumn(column, idx, field, t.Aliasers...)
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

// MapFieldToColumn maps a column to a field based on the provided aliases.
func MapFieldToColumn(column string, colIdx int, field *ast.Field, aliases ...Aliaser) (ColumnMap, bool, error) {
	if len(field.Names) != 1 {
		return ColumnMap{}, false, fmt.Errorf("field had more than 1 name")
	}

	fieldName := field.Names[0].Name
	for _, aliaser := range aliases {
		if column == aliaser.Alias(fieldName) {
			return ColumnMap{
				ColumnName:   column,
				FieldName:    fieldName,
				ColumnOffset: colIdx,
				Type:         field.Type,
			}, true, nil
		}
	}

	return ColumnMap{}, false, nil
}
