//+build genieql,generate,structure,table

package definitions

// the genieql.options lines allow for customizing
// the output for the given table(s).
// [general] section:
// option alias: set the strategy for renaming the columns. valid strategies are:
// camelcase (default), snakecase, lowercase, uppercase.
//
// [rename.columns] section: allows use of a kv mapping to rename columns explicitly.
//genieql.options: [general] alias=camelcase
//genieql.options: [rename.columns] c1=f1
const Example1 = "example1"

// example2 uses the default configuration.
const Example2 = "example2"
