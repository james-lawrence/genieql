# genieql
sql query and code generation program.
its purpose is to generate as much of the
boilerplate code for interacting with database/sql
as possible. without putting any runtime dependencies into
your codebase.

# is it production ready?
its very much in alpha code, it only supports
postgresql currently, no test coverage on the code.
mainly getting it out early to solicite feedback on the api
of the code that gets generated and feature requests.

as a result you should expect the api to change/break until around 1.0.

# communication
I've created a [google+ community](https://plus.google.com/communities/103872946940860163885) around genieql.

# documentation
release notes, FAQ, and roadmap documentation
can be found in the documentation directory.

## genieql commands
- bootstrap - saves database information to a file for other commands.
- map - writes a configuration file describing how to map a structure to a database column.
- generate - used to generate scanners and queries.

## example usage
```go
package mypackage
//go:generate genieql map --natural-key=id github.com/mypackage.MyType my_table
//go:generate genieql generate crud --output=mytype_crud_gen.go github.com/mypackage.MyType my_table
```

## genieql bootstrap command
```text
usage: qlgenie bootstrap [<flags>] <uri>

build a instance of qlgenie

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).
  --output-directory="/home/james-lawrence/development/guardian/.genieql/default.config"  
          directory to place the configuration file

Args:
  <uri>  uri for the database qlgenie will work with
```
```bash
genieql bootstrap postgres://username@localhost:5432/databasename?sslmode=disable
```
```yml
// generates this file at $GOPATH/.genieql/default.config
dialect: postgres
connectionurl: postgres://jatone@localhost:5432/sso?sslmode=disable
host: localhost
port: 5432
database: databasename
username: username
password: ""
```
## genieql map command
```text
usage: qlgenie map [<flags>] <package.type> <table> [<transformations>...]

define mapping configuration for a particular type/table combination

Flags:
  --help                     Show context-sensitive help (also try --help-long and --help-man).
  --config="default.config"  configuration to use
  --include-table-prefix-aliases  
                             generate additional aliases with the table name prefixed i.e.) my_column -> my_table_my_column
  --natural-key=id ...       natural key for this mapping, this is deprecated will be able to automatically determine in later versions

Args:
  <package.type>       location of type to work with github.com/soandso/package.MyType
  <table>              table that we are mapping
  [<transformations>]  transformations (in left to right order) to apply to structure fields to map them to column names
```
```bash
genieql map --natural-key=id github.com/soandso/project.MyType MyTable snakecase lowercase
genieql map --natural-key=col1, --natural-key=col2  github.com/soandso/project.MyType MyTable snakecase lowercase
```

## genieql generate command
```text
usage: qlgenie generate <command> [<args> ...]

generate sql queries

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Subcommands:
  generate crud [<flags>] <package.Type> <table>
    generate crud queries (INSERT, SELECT, UPDATE, DELETE)
```
```bash
genieql generate crud --output=mytype_crud_gen.go github.com/soandso/project.Type table
```
