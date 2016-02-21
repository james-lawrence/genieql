package postgresql

import "bitbucket.org/jatone/genieql/sqlutil"

// Dialect constant representing the dialect name.
const Dialect = "postgres"

func init() {
	maybePanic := func(err error) {
		if err != nil {
			panic(err)
		}
	}

	maybePanic(sqlutil.RegisterColumnQuery(Dialect, columnQuery))
	maybePanic(sqlutil.RegisterPrimaryKeyQuery(Dialect, primaryKeyQuery))
}

const primaryKeyQuery = `SELECT a.attname	FROM pg_index i
JOIN pg_attribute a ON a.attrelid = i.indrelid
AND a.attnum = ANY(i.indkey)
WHERE  i.indrelid = '%s'::regclass
AND    i.indisprimary`

const columnQuery = `SELECT * FROM %s LIMIT 1`
