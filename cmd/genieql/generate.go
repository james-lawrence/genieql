package main

import (
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	"bitbucket.org/jatone/genieql/dialects"
)

func loadContext(config string) (configuration genieql.Configuration, dialect genieql.Dialect, err error) {
	configuration = genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), config),
		),
	)

	if dialect, err = dialects.LookupDialect(configuration); err != nil {
		return configuration, dialect, err
	}

	return configuration, dialect, err
}
