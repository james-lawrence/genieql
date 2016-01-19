#SQLGENIE Specification
##configuration
SQLGENIEDIR defaults to $GOPATH/.sqlgenie

## SQLGenie command
sqlgenie provide the initialization of the database details for validation/warnings.
```bash
sqlgenie initialize --dialect=postgres --username=xxx --password=yyy --dial=zzz
```

## SQLMap command
sqlmap builds mappings of tables to structures for use by other commands.
output is stored in $SQLGENIEDIR/maps. sqlmap will provide subcommands for
common transformations such as snakecase and downcase, it will also provide an
escape hatch by allowing a hand tailored mapping.
```bash
# definitions:
## unmapped column: a column that appears in the code but not the database.
# custom mapping example, emits warnings for unmapped columns.
sqlmap --package=github.com/soandso/project --type=A --table=A --alias="Value1=dbcol1" --alias="Value2=dbcol2"
# lowersnakecase example) will convert structure fields to snakecase then lowercase
sqlmap --package=github.com/soandso/project --type=A --table=A --alias="FieldName1=dbcol1" snakecase lowercase
# uppersnakecase example) will convert structure fields to snakecase then uppercase
sqlmap --package=github.com/soandso/project --type=A --table=A --alias="FieldName1=dbcol1" snakecase uppercase
# lowercase example) will convert structure fields to lowercase.
sqlmap --package=github.com/soandso/project --type=A --table=A --alias="FieldName1=dbcol1" lowercase
# uppercase example) will convert structure fields to uppercase.
sqlmap --package=github.com/soandso/project --type=A --table=A --alias="FieldName1=dbcol1" uppercase
```

## SQLScanner command
A scanner converts a row into a struct.
sqlscanner will use the generated map files to build Scanners that
read rows provided from sql queries and inserts them into structures.
It requires that the structures being populated have been mapped, see sqlmap.
it will error if no mapping exists.
```bash
# will create a basic scanner for the provided type. basic scanner is a direct 1 to 1, table to struct
# scanner.
sqlscanner basic --package github.com/jatone/project --type=A

# merge will create a scanner for fields from multiple type maps.
# will fill in a type like type MergedType struct {A,B}, requires there
# is a unique alias for each field in all involved structures.
sqlscanner merge github.com/jatone/project.A github.com/jatone/project.B

# onetomany will handle n+1 queries.
# type OneToMany struct{A,[]B,[]C}
sqlscanner onetomany
```

## SQLGen command
sqlgen will generate common query patterns for the DB.
```bash
# crud pattern will generate common SELECT * FROM X WHERE col = ?
# will do each type individually.
# uses sqlscanner merge semantics for result scanning.
sqlgen crud github.com/jatone/project.A github.com/jatone/project.B

# join pattern will generate common SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, C ON A.id = C.a_id
# uses sqlscanner merge semantics for result scanning.
sqlgen join github.com/jatone/project.A.id github.com/jatone/project.B.a_id github.com/jatone/project.C.a_id

# chainjoin pattern will generate common SELECT A.*, B.*, C.* FROM A JOIN B ON A.id = B.a_id, C ON B.id = C.b_id
# uses sqlscanner merge semantics for result scanning.
sqlgen chainjoin github.com/jatone/project.A.id github.com/jatone/project.B.a_id github.com/jatone/project.C.b_id

# onetomany handle the n+1 queries.
# uses sqlscanner onetomany semantics for result scanning.
sqlgen onetomany github.com/jatone/project.A github.com/jatone/project.B github.com/jatone/project.C
```