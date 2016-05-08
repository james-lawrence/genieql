# Roadmap
## Done
- support postgresql basic crud queries and scanner. (0.0.1)
- use the database schema to determine the primary key columns for a table. (0.0.2)
- be able to generate insert queries with DEFAULT values. (0.0.3)
- support pointer fields. (0.0.4)
- support driver specific null types. (0.0.5)
- support dynamic field scanner. (0.0.5)

## Upcoming
### these are listed in no particular order.
- support writing the generated code into other packages, separate from where the type is located.
- support generating functions that execute particular queries and scan them into a structure. postponed until I determine what to do about sql.DB/sql.Tx

```go
func LookupMyType(db *sql.DB, id int, dst *MyType) error {
    scanner := NewSomeScanner(db.Query(SomeQueryConstant, id))
    defer scanner.Close()

    return scanner.Scan(dst)
}
```
