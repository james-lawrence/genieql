package genieql

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Dialect       string
	ConnectionURL string
	Host          string
	Port          int
	Database      string
	Username      string
	Password      string
}

func Bootstrap(path string, uri *url.URL) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	config, err := ConfigurationFromURI(uri)
	if err != nil {
		return err
	}

	return WriteConfiguration(path, config)
}

func WriteConfiguration(path string, configuration Configuration) error {
	d, err := yaml.Marshal(configuration)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, d, 0666)
}

func ReadConfiguration(path string, config *Configuration) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(raw, config)
}

func ConfigurationFromURI(uri *url.URL) (Configuration, error) {
	var password string
	splits := strings.Split(uri.Host, ":")
	if len(splits) != 2 {
		return Configuration{}, fmt.Errorf("invalid host/port combination")
	}

	host, portString := splits[0], splits[1]

	port, err := strconv.Atoi(portString)
	if err != nil {
		return Configuration{}, err
	}

	password, _ = uri.User.Password()

	return Configuration{
		ConnectionURL: uri.String(),
		Dialect:       uri.Scheme,
		Host:          host,
		Port:          port,
		Database:      strings.Trim(uri.Path, "/"),
		Username:      uri.User.Username(),
		Password:      password,
	}, nil
}
