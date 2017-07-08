package functions

//go:generate genieql generate experimental structure table constants -o postgresql.table.structs.gen.go
//go:generate genieql generate experimental scanners types -o postgresql.scanners.gen.go
//go:generate genieql generate experimental crud -o postgresql.crud.functions.gen.go --table=example1 --scanner=NewExample1ScannerDynamic --unique-scanner=NewExample1ScannerStaticRow Example1
//go:generate genieql generate experimental functions types -o postgresql.functions.gen.go
//go:generate genieql generate insert --suffix=WithDefaults -o postgresql.insert.gen.go --default=id --default=text_field --default=created_at --default=updated_at Example1 example1
//go:generate genieql generate insert experimental batch-function -o postgresql.insert.batch.gen.go
