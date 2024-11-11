# genieql - sql query and code generation.
its purpose is to generate a decent amount of the
boilerplate code for interacting with database/sql
as possible without putting any runtime dependencies
into your codebase. primary areas of focus are:
1. data scanners (hydrating structures from queries)
2. make support and maintaince for simple queries a breeze.
3. integrate well with the broader ecosystem. aka: scanners should play well
with query builders.

# is it production ready?
its nearing production ready, essentially beta code, but it is in use on few production applications already.

- it only supports postgresql currently.
- sqlite support is otw.
- adding additional support is very straight forward, just implement the Dialect interface. see the postgresql implementation as the example.
- test coverage is getting added. (a good chunk already exists but working on a better test harness for integration tests)

mainly getting it out early to solicite feedback on the api
of the code that gets generated and feature requests.

as a result you should expect the api to change/break until around 1.0.

# documentation
release notes, and roadmap documentation
can be found in the documentation directory.
everything else will be found in godoc.

## genieql commands
- genieql bootstrap - saves database information for generation.
- genieql auto - runs the gql scripts to generate database code.

## examples
see the examples directory.