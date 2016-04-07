package genieql

import (
	"strings"

	"github.com/serenize/snaker"
)

// Aliaser implementations auto generate aliases
type Aliaser interface {
	Alias(string) string
}

// AliaserFunc pure functional implementations of the Aliaser
type AliaserFunc func(string) string

// Alias see Aliaser
func (t AliaserFunc) Alias(name string) string {
	return t(name)
}

// AliaserChain TODO ...
func AliaserChain(aliasers ...Aliaser) Aliaser {
	return AliaserFunc(func(name string) string {
		for _, aliaser := range aliasers {
			name = aliaser.Alias(name)
		}

		return name
	})
}

// MultiAliaser TODO ...
func MultiAliaser(name string, aliasers ...Aliaser) []string {
	result := make([]string, 0, len(aliasers))
	for _, aliaser := range aliasers {
		result = append(result, aliaser.Alias(name))
	}

	return result
}

// AliaserBuilder looks up transformations by name, if any of transformations
// do not exist returns nil.
func AliaserBuilder(names ...string) Aliaser {
	aliaserSet := make([]Aliaser, 0, len(names))
	for _, name := range names {
		aliaser := AliaserSelect(name)
		if aliaser == nil {
			return nil
		}
		aliaserSet = append(aliaserSet, aliaser)
	}

	return AliaserChain(aliaserSet...)
}

// AliaserSelect predefines common transformations for Aliases
func AliaserSelect(aliasername string) Aliaser {
	switch strings.ToLower(aliasername) {
	case "lowercase":
		return AliasStrategyLowercase
	case "uppercase":
		return AliasStrategyUppercase
	case "snakecase":
		return AliasStrategySnakecase
	case "camelcase":
		return AliasStrategyCamelcase
	default:
		return nil
	}
}

// AliasStrategyLowercase strategy for lowercasing field names to match result fields.
var AliasStrategyLowercase Aliaser = AliaserFunc(strings.ToLower)

// AliasStrategyUppercase strategy for uppercasing field names to match result fields.
var AliasStrategyUppercase Aliaser = AliaserFunc(strings.ToUpper)

// AliasStrategySnakecase strategy for snake casing field names to match result fields.
var AliasStrategySnakecase Aliaser = AliaserFunc(snaker.CamelToSnake)

// AliasStrategySnakecase strategy for camel casing field names to match result fields.
var AliasStrategyCamelcase Aliaser = AliaserFunc(snaker.SnakeToCamel)
