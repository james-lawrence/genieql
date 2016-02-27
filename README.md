# genieql
sql query and code generation program.
its purpose is to generate as much of the
boilerplate code for interacting with database/sql
as possible. without putting any runtime dependencies into
your codebase.

# is it production ready?
its very much in alpha code, but it is in use on few production applications already.

-it only supports postgresql currently.

-adding additional support is very straight forward, just implement the Dialect interface. see the postgresql implementation as the example.

-currently minimal test coverage.

mainly getting it out early to solicite feedback on the api
of the code that gets generated and feature requests.

as a result you should expect the api to change/break until around 1.0.

# communication
I've created a [google+ community](https://plus.google.com/communities/103872946940860163885) around genieql.

# documentation
release notes, and roadmap documentation
can be found in the documentation directory.
everything else will be found in godoc.

## genieql commands
- bootstrap - saves database information to a file for other commands.
- map - writes a configuration file describing how to map a structure to a database column.
- generate - used to generate scanners and queries.
  - scanner part will go away in future release and moved entirely to the scanner cli
- scanner - used to create scanners
## example usage
```go
package mypackage
//go:generate genieql map github.com/mypackage.MyType
//go:generate genieql generate crud --output=mytype_crud_gen.go github.com/mypackage.MyType my_table
```

## genieql bootstrap command
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
```bash
qlgenie map github.com/soandso/project.MyType snakecase lowercase
```

## genieql generate command
```bash
genieql generate crud --output=mytype_crud_gen.go github.com/soandso/project.Type table
```