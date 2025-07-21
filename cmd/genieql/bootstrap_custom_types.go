package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/drivers"
	"gopkg.in/yaml.v3"
)

type bootstrapCustom struct {
	configName string
}

func (t *bootstrapCustom) Bootstrap(ctx *kingpin.ParseContext) error {
	config := genieql.MustReadConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), t.configName),
		),
	)
	log.Println("bootstraping", config.Location)

	encoded, err := yaml.Marshal([]genieql.ColumnDefinition{{
		Type:       "sql.NullString",
		ColumnType: "sql.NullString",
		Native:     "string",
		Decode:     drivers.StdlibDecodeString,
		Encode:     drivers.StdlibEncodeString,
	}})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(config.Location, "driver.yml"), encoded, 0600)
}

func (t *bootstrapCustom) configure(bootstrap *kingpin.CmdClause) *kingpin.CmdClause {
	bootstrap.Flag("config", "name of the genieql configuration to use").Default(defaultConfigurationName).StringVar(&t.configName)
	bootstrap.Action(t.Bootstrap)

	return bootstrap
}
