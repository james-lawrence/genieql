// Package genieql - cli for dealing with database access in go.
//
// bootstrap command - defines how to talk to a database.
//  genieql bootstrap --help
//   usage: qlgenie bootstrap [<flags>] <uri>
//
//   build a instance of qlgenie
//
//   Flags:
//     --help  Show context-sensitive help (also try --help-long and --help-man).
//     --output-directory="/home/james-lawrence/development/guardian/.genieql/default.config"
//             directory to place the configuration file
//
//   Args:
//     <uri>  uri for the database qlgenie will work with
// map command - define how a structure should be mapped.
// 	genieql map --help
//	 usage: qlgenie map [<flags>] <package.type> [<transformations>...]
//
//	 define mapping configuration for a particular type/table combination
//
//	 Flags:
//	   --help                     Show context-sensitive help (also try --help-long and --help-man).
//	   --config="default.config"  configuration to use
//	   --include-table-prefix-aliases
//	                              generate additional aliases with the table name prefixed i.e.) my_column -> my_table_my_column
//	   --mapping="default"        name to give the mapping
//
//	 Args:
//	   <package.type>       location of type to work with github.com/soandso/package.MyType
//	   [<transformations>]  transformations (in left to right order) to apply to structure fields to map them to column names
// generate command - generate sql queries.
//  genieql generate --help
//	 usage: qlgenie generate <command> [<args> ...]
//
//	 generate sql queries
//
//	 Flags:
//	   --help  Show context-sensitive help (also try --help-long and --help-man).
//
//	 Subcommands:
//	   generate crud [<flags>] <package.Type> <table>
//	     generate crud queries (INSERT, SELECT, UPDATE, DELETE)
// scanner command - provides an api for building scanners
// 	genieql scanner query-literal --help
//	usage: qlgenie scanner query-literal [<flags>] <scanner-name> <package.Type> <package.Query>
//
//	 build a scanner for the provided type/query
//
//	 Flags:
//	   --help                     Show context-sensitive help (also try --help-long and --help-man).
//	   --config="default.config"  name of configuration file to use
//	   --mapping="default"        name of the map to use
//	   --output=""                path of output file
//
//	 Args:
//	   <scanner-name>   name of the scanner
//	   <package.Type>   package prefixed structure we want a scanner for
//	   <package.Query>  package prefixed constant we want to use the query
package main
