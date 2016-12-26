package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction5(q sqlx.Queryer, query string, uuidArgument int, camelcaseArgument int, snakecaseArgument int, uppercaseArgument int, lowercaseArgument int) ExampleRowScanner {
	return StaticExampleRowScanner(q.QueryRow(query, uuidArgument, camelcaseArgument, snakecaseArgument, uppercaseArgument, lowercaseArgument))
}
