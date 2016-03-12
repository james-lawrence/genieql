package integration_tests

// DO NOT MODIFY: This File was auto generated by the following command:
// genieql generate crud --config=scanner-test.config --output=type1_queries_gen.go bitbucket.org/jatone/genieql/scanner/internal/integration_tests.Type1 type1

const Type1Insert = `INSERT INTO type1 (field1,field2,field3,field4,field5,field6) VALUES ($1,$2,$3,$4,$5,$6) RETURNING field1,field2,field3,field4,field5,field6`
const Type1FindByField1 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field1 = $1`
const Type1FindByField2 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field2 = $1`
const Type1FindByField3 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field3 = $1`
const Type1FindByField4 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field4 = $1`
const Type1FindByField5 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field5 = $1`
const Type1FindByField6 = `SELECT field1,field2,field3,field4,field5,field6 FROM type1 WHERE field6 = $1`
const Type1UpdateByID = `UPDATE type1 SET (field1 = $1, field2 = $2, field3 = $3, field4 = $4, field5 = $5, field6 = $6) WHERE  RETURNING field1,field2,field3,field4,field5,field6`
const Type1DeleteByID = `DELETE FROM type1 WHERE  RETURNING field1,field2,field3,field4,field5,field6`