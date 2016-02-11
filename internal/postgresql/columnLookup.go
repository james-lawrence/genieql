package postgresql

import "bitbucket.org/jatone/genieql/sqlutil"

func init() {
	sqlutil.RegisterColumnQuery("postgres", "SELECT * FROM %s LIMIT 1")
}
