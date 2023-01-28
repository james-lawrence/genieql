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

-test coverage is getting added.

mainly getting it out early to solicite feedback on the api
of the code that gets generated and feature requests.

as a result you should expect the api to change/break until around 1.0.

# documentation
release notes, and roadmap documentation
can be found in the documentation directory.
everything else will be found in godoc.

## genieql commands
- bootstrap - saves database information to a file for other commands.
- map - writes a configuration file describing how to map a structure to a database column.
- generate - used to generate queries. main use case is to bootstrap a project quickly.
- scanner - used to create scanners.

## examples
see the examples directory.

failed to build code generator: Example1: failed to compile source: 4:2:
import "bitbucket.org/jatone/genieql/interp"
error: /home/james.lawrence/development/genieql/interp/functions.go:4:2:import "go/ast"
error: /usr/lib/go/src/go/ast/ast.go:10:2: import "go/token"
error: /usr/lib/go/src/go/token/position.go:8:2: import "fmt"
error: /usr/lib/go/src/fmt/errors.go:7:8: import "errors"
error: /usr/lib/go/src/errors/wrap.go:8:2: import "internal/reflectlite"
error: /usr/lib/go/src/internal/reflectlite/swapper.go:8:2: import "internal/goarch"
error: /usr/lib/go/src/internal/goarch/gengoarch.go:10:2: import "bytes"
error: /usr/lib/go/src/bytes/buffer.go:10:2: import "errors"
error: import cycle not allowed imports errors)