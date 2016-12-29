package genieql

import "fmt"

// ErrMissingDialect - returned when a dialect has not been registered.
type ErrMissingDialect interface {
	MissingDialect() string
}

// IsMissingDialectErr determines if the given error is a missing dialect error.
func IsMissingDialectErr(err error) bool {
	_, ok := err.(ErrMissingDialect)
	return ok
}

type errMissingDialect struct {
	dialect string
}

func (t errMissingDialect) MissingDialect() string {
	return t.dialect
}

func (t errMissingDialect) Error() string {
	return fmt.Sprintf("dialect (%s) is not registered", t.dialect)
}

// ErrDuplicateDialect - returned when a dialect gets registered twice.
var ErrDuplicateDialect = fmt.Errorf("dialect has already been registered")

var dialects = dialectRegistry{}

// RegisterDialect register a sql dialect with genieql. usually in an init function.
func RegisterDialect(dialect string, imp DialectFactory) error {
	return dialects.RegisterDialect(dialect, imp)
}

// LookupDialect lookup a registered dialect.
func LookupDialect(config Configuration) (Dialect, error) {
	var (
		err     error
		factory DialectFactory
	)

	if factory, err = dialects.LookupDialect(config.Dialect); err != nil {
		return nil, err
	}

	return factory.Connect(config)
}

// DialectFactory ...
type DialectFactory interface {
	Connect(Configuration) (Dialect, error)
}

// Dialect ...
type Dialect interface {
	Insert(table string, columns, defaults []string) string
	Select(table string, columns, predicates []string) string
	Update(table string, columns, predicates []string) string
	Delete(table string, columns, predicates []string) string
	ColumnValueTransformer() ColumnTransformer
	ColumnInformationForTable(table string) ([]ColumnInfo, error)
	ColumnInformationForQuery(query string) ([]ColumnInfo, error)
}

type dialectRegistry map[string]DialectFactory

func (t dialectRegistry) RegisterDialect(dialect string, imp DialectFactory) error {
	if _, exists := t[dialect]; exists {
		return ErrDuplicateDialect
	}

	t[dialect] = imp

	return nil
}

func (t dialectRegistry) LookupDialect(dialect string) (DialectFactory, error) {
	impl, exists := t[dialect]
	if !exists {
		return nil, errMissingDialect{dialect: dialect}
	}

	return impl, nil
}
