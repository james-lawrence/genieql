package genieql

import (
	"go/ast"
	"go/build"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"

	"bitbucket.org/jatone/genieql/astutil"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
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

// TypeFields returns the fields of underlying struct of the mapping.
func (t MappingConfig) TypeFields(fset *token.FileSet, pkg *build.Package) ([]*ast.Field, error) {
	return NewSearcher(fset, pkg).FindFieldsForType(ast.NewIdent(t.Type))
}

// ColumnInfo defined by the mapping.
func (t MappingConfig) ColumnInfo(dialect Dialect) ([]ColumnInfo, error) {
	if t.CustomQuery {
		return dialect.ColumnInformationForQuery(t.TableOrQuery)
	}
	return dialect.ColumnInformationForTable(t.TableOrQuery)
}

// MappedColumnInfo returns the mapped and unmapped columns for the mapping.
func (t MappingConfig) MappedColumnInfo(dialect Dialect, fset *token.FileSet, pkg *build.Package) ([]ColumnInfo, []ColumnInfo, error) {
	var (
		err     error
		fields  []*ast.Field
		columns []ColumnInfo
	)

	if fields, err = t.TypeFields(fset, pkg); err != nil {
		return []ColumnInfo(nil), []ColumnInfo(nil), errors.Wrapf(err, "failed to lookup fields: %s.%s", t.Package, t.Type)
	}

	if columns, err = t.ColumnInfo(dialect); err != nil {
		return []ColumnInfo(nil), []ColumnInfo(nil), errors.Wrapf(err, "failed to lookup columns: %s.%s using %s", t.Package, t.Type, t.TableOrQuery)
	}

	mColumns, uColumns := mapColumns(columns, fields, t.Mapper().Aliasers...)
	return mColumns, uColumns, nil
}

// MappedFields returns the fields that are mapped to columns.
func (t MappingConfig) MappedFields(dialect Dialect, fset *token.FileSet, pkg *build.Package, ignoreColumnSet ...string) ([]*ast.Field, []*ast.Field, error) {
	var (
		err     error
		columns []ColumnInfo
	)

	if columns, err = t.ColumnInfo(dialect); err != nil {
		return []*ast.Field{}, []*ast.Field{}, errors.Wrapf(err, "failed to lookup columns: %s.%s using %s", t.Package, t.Type, t.TableOrQuery)
	}

	return t.MapFieldsToColumns(fset, pkg, ColumnInfoSet(columns).Filter(ColumnInfoFilterIgnore(ignoreColumnSet...))...)
}

func (t MappingConfig) MapFieldsToColumns(fset *token.FileSet, pkg *build.Package, columns ...ColumnInfo) ([]*ast.Field, []*ast.Field, error) {
	var (
		err    error
		fields []*ast.Field
	)

	if fields, err = t.TypeFields(fset, pkg); err != nil {
		return []*ast.Field{}, []*ast.Field{}, errors.Wrapf(err, "failed to lookup fields: %s.%s", t.Package, t.Type)
	}

	mFields, uFields := mapFields(columns, fields, t.Mapper().Aliasers...)
	return mFields, uFields, nil
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
			matched *ast.Field
		)

		for _, field := range fields {
			if matched = MapFieldToColumn(column, idx, field, t.Aliasers...); matched != nil {
				break
			}
		}

		if matched == nil {
			matches = append(matches, column)
			break
		}
	}

	return matches, nil
}

// MapFieldToColumn maps a column to a field based on the provided aliases.
func MapFieldToColumn(column string, colIdx int, field *ast.Field, aliases ...Aliaser) *ast.Field {
	for _, fieldName := range field.Names {
		for _, aliaser := range aliases {
			if aliaser.Alias(column) == fieldName.Name {
				return astutil.Field(field.Type, fieldName)
			}
		}
	}
	return nil
}
