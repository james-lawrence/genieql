package example

func explodeFunction1(arg1 *Foo) []interface{} {
	return []interface{}{arg1.field1, arg1.field2, arg1.field3}
}
