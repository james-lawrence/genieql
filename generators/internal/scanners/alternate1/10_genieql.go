package alternate1

//go:generate genieql generate experimental structure table constants --config=generators-test.config -o postgresql.table.structs.gen.go
//go:generate genieql generate experimental scanners types --config=generators-test.config -o postgresql.scanners.gen.go
//go:generate genieql generate experimental crud --config=generators-test.config -o postgresql.crud.functions.gen.go --table=type1 --scanner=NewType1ScannerStatic --unique-scanner=NewType1ScannerStaticRow Type1
//go:generate genieql generate experimental functions types --config=generators-test.config -o postgresql.functions.gen.go
//go:generate genieql generate insert experimental batch-function --config=generators-test.config -o postgresql.insert.batch.gen.go
