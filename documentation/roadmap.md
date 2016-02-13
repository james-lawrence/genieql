# Roadmap
## Done
- support postgresql basic crud queries and scanner.

## Upcoming
these are listed in no particular order.
- use the database schema to determine the primary key columns
for a table.
- support writing the generated code into other packages, separate from where the type is located.
- support generating functions that execute particular queries and scan them into a structure.
```go
func LookupMyType(exec *sql.DB, id int, dst *MyType) error {
    rows, err := exec.Query(SomeQueryConstant, id)
    if err == nil {
      defer rows.Close()
    }

    return NewSomeScanner(rows, erro).Scan(dst)
}
```
- support for generating join queries
```bash
genieql generate join (github.com/soandso/package.A.id,A) (github.com/soandso/package.B.a_id,B) (github.com/soandso/package.C.a_id,C)
```
```text
SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, JOIN C ON A.id = C.a_id
```
- support chain join queries.
```bash
genieql generate chainjoin (github.com/soandso/package.A.id,A) (github.com/soandso/package.B.a_id,B) (github.com/soandso/package.C.b_id,C)
```
```
SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, C ON B.id = C.b_id
```
