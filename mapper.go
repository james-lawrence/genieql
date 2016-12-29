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

// MappingConfigOption (MCO) options for building MappingConfigs.
type MappingConfigOption func(*MappingConfig)

func MCOPackage(p string) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.Package = p
	}
}

func MCOColumnInfo(q string) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.TableOrQuery = q
	}
}

func MCOCustom(custom bool) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.CustomQuery = custom
	}
}

func MCODialect(d Dialect) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.dialect = d
	}
}

func MCOTransformations(t ...string) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.Transformations = t
	}
}

func MCORenameMap(m map[string]string) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.RenameMap = m
	}
}

func MCOType(t string) MappingConfigOption {
	return func(mc *MappingConfig) {
		mc.Type = t
	}
}

func NewMappingConfig(options ...MappingConfigOption) MappingConfig {
	mc := MappingConfig{}

	(&mc).Apply(options...)

	return mc
}

// MappingConfig TODO...
type MappingConfig struct {
	Package              string
	Type                 string
	IncludeTablePrefixes bool // deprecated
	Transformations      []string
	RenameMap            map[string]string
	TableOrQuery         string
	CustomQuery          bool
	dialect              Dialect
}

// Apply the options to the current MappingConfig
func (t *MappingConfig) Apply(options ...MappingConfigOption) {
	for _, opt := range options {
		opt(t)
	}
}

// Mapper ...
func (t MappingConfig) Mapper() Mapper {
	return Mapper{Aliasers: []Aliaser{AliaserBuilder(t.Transformations...)}}
}

// Aliaser ...
func (t MappingConfig) Aliaser() Aliaser {
	alias := AliaserBuilder(t.Transformations...)
	return AliaserFunc(func(name string) string {
		// if the configuration explicitly renames
		// a column use that value do not try to
		// transform it.
		if v, ok := t.RenameMap[name]; ok {
			return v
		}

		return alias.Alias(name)
	})
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

// UnmappedColumns returns the columns that do not map to a field.
func (t Mapper) UnmappedColumns(fields []*ast.Field, columns ...string) ([]string, error) {
	matches := make([]string, 0, len(columns))
	for idx, column := range columns {
		var (
			matched bool
			err     error
		)

		for _, field := range fields {
			matched, err = MapFieldToColumn(column, idx, field, t.Aliasers...)
			if err != nil {
				return matches, err
			}

			if matched {
				break
			}
		}

		if !matched {
			matches = append(matches, column)
			break
		}
	}

	return matches, nil
}

// MapFieldToColumn maps a column to a field based on the provided aliases.
func MapFieldToColumn(column string, colIdx int, field *ast.Field, aliases ...Aliaser) (bool, error) {
	if len(field.Names) != 1 {
		return false, fmt.Errorf("field had more than 1 name")
	}

	fieldName := field.Names[0].Name
	for _, aliaser := range aliases {
		if aliaser.Alias(column) == fieldName {
			return true, nil
		}
	}

	return false, nil
}
