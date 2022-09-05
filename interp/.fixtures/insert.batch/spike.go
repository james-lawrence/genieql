package example

type example1BatchInsertFunction struct {
	q         sqlx.Queryer
	remaining []Example1
	scanner   Example1Scanner
}
