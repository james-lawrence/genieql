# Roadmap
## Done
- support postgresql basic crud queries and scanner. (0.0.1)
- use the database schema to determine the primary key columns for a table. (0.0.2)

## Upcoming
these are listed in no particular order.
- be able to generate insert queries with DEFAULT values.
- support pointer fields.
- support writing the generated code into other packages, separate from where the type is located.
- support generating functions that execute particular queries and scan them into a structure.
```go
func LookupMyType(db *sql.DB, id int, dst *MyType) error {
    scanner := NewSomeScanner(db.Query(SomeQueryConstant, id))
    defer scanner.Close()

    return scanner.Scan(dst)
}
```
- support for generating join queries
```bash
genieql generate join (github.com/soandso/package.A,A.id) (github.com/soandso/package.B,B.a_id) (github.com/soandso/package.C,C.a_id)
```
```text
SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, JOIN C ON A.id = C.a_id
```
- support chain join queries.
```bash
genieql generate chainjoin (github.com/soandso/package.A,A.id) (github.com/soandso/package.B,B.a_id) (github.com/soandso/package.C,C.b_id)
```
```
SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, C ON B.id = C.b_id
```
