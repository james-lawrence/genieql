package genieql

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Mapper struct {
	Package              string
	Type                 string
	Table                string
	IncludeTablePrefixes bool
	NaturalKey           []string
	Transformations      []string
}

func WriteMapper(root string, configuration Configuration, m Mapper) error {
	d, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	path := filepath.Join(root, configuration.Database, m.Package, m.Type, m.Table)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return ioutil.WriteFile(path, d, 0666)
}

func Map(configFile string, m Mapper) error {
	var config Configuration

	if err := ReadConfiguration(configFile, &config); err != nil {
		return err
	}

	return WriteMapper(filepath.Dir(configFile), config, m)
}
