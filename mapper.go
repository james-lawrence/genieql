package genieql

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/token"
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
	TableOrQuery         string
	CustomQuery          bool
	dialect              Dialect
}

// Mapper ...
func (t MappingConfig) Mapper() Mapper {
	return Mapper{Aliasers: []Aliaser{AliaserBuilder(t.Transformations...)}}
}

// Aliaser ...
func (t MappingConfig) Aliaser() Aliaser {
	return AliaserBuilder(t.Transformations...)
}

// TypeFields ...
func (t MappingConfig) TypeFields(fset *token.FileSet, context build.Context, filter func(*build.Package) bool) ([]*ast.Field, error) {
	pkg, err := LocatePackage(t.Package, context, filter)
	if err != nil {
		return nil, err
	}

	typ, err := NewUtils(fset).FindUniqueType(FilterName(t.Type), pkg)
	if err != nil {
		return nil, err
	}

	return ExtractFields(typ).List, nil
}

// ColumnInfo defined by the mapping.
func (t MappingConfig) ColumnInfo() ([]ColumnInfo, error) {
	if t.CustomQuery {
		return t.dialect.ColumnInformationForQuery(t.TableOrQuery)
	}

	return t.dialect.ColumnInformationForTable(t.TableOrQuery)
}

// WriteMapper persists the structure -> result row mapping to disk.
func WriteMapper(config Configuration, name string, m MappingConfig) error {
	d, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	path := filepath.Join(config.Location, config.Database, m.Package, m.Type, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, d, 0666)
}

// ReadMapper loads the structure -> result row mapping from disk.
func ReadMapper(config Configuration, pkg, typ, name string, m *MappingConfig) error {
	var (
		err error
	)

	if m.dialect, err = LookupDialect(config); err != nil {
		return err
	}

	raw, err := ioutil.ReadFile(filepath.Join(config.Location, config.Database, pkg, typ, name))
	if err != nil {
		return err
	}

	return yaml.Unmarshal(raw, m)
}

// Map TODO...
func Map(configFile, name string, m MappingConfig) error {
	var config = Configuration{
		Location: filepath.Dir(configFile),
		Name:     filepath.Base(configFile),
	}

	if err := ReadConfiguration(&config); err != nil {
		return err
	}

	return WriteMapper(config, name, m)
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
