package example

import "bitbucket.org/jatone/genieql/sqlx"

func queryFunction5(q sqlx.Queryer, query string, camelcaseArgument int, snakecaseArgument int, uppercaseArgument int, lowercaseArgument int) ExampleScanner {
	return StaticExampleScanner(q.QueryRow(query, camelcaseArgument, snakecaseArgument, uppercaseArgument, lowercaseArgument))
}
