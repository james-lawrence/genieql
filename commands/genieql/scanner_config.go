package main

import (
	"fmt"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

type scannerOption func(c *scannerConfig)

func defaultScannerNameFormat(format string) scannerOption {
	return func(c *scannerConfig) {
		c.defaultScannerNameFormat = format
	}
}

func defaultRowScannerNameFormat(format string) scannerOption {
	return func(c *scannerConfig) {
		c.defaultRowScannerNameFormat = format
	}
}

func defaultInterfaceNameFormat(format string) scannerOption {
	return func(c *scannerConfig) {
		c.defaultInterfaceNameFormat = format
	}
}

func defaultInterfaceRowNameFormat(format string) scannerOption {
	return func(c *scannerConfig) {
		c.defaultInterfaceRowNameFormat = format
	}
}

func defaultErrScannerNameFormat(format string) scannerOption {
	return func(c *scannerConfig) {
		c.defaultErrScannerNameFormat = format
	}
}

type scannerConfig struct {
	defaultScannerNameFormat      string
	defaultRowScannerNameFormat   string
	defaultInterfaceNameFormat    string
	defaultInterfaceRowNameFormat string
	defaultErrScannerNameFormat   string

	packageType      string
	scannerName      string
	scannerRowName   string
	interfaceName    string
	interfaceRowName string
	errScannerName   string
	configName       string
	mapName          string
	output           string
	private          bool
}

func (t *scannerConfig) configure(cmd *kingpin.CmdClause, opts ...scannerOption) {
	cmd.Flag("config", "name of configuration file to use").Default("default.config").
		StringVar(&t.configName)
	cmd.Flag("mapping", "name of the map to use").Default("default").
		StringVar(&t.mapName)
	cmd.Flag("output", "path of output file").Default("").
		StringVar(&t.output)
	cmd.Flag("internal-interface", "make the scanner internal to the package").
		Default("false").BoolVar(&t.private)
	cmd.Flag("scanner-name", "name of the scanner, defaults to type name").
		Default("").StringVar(&t.scannerName)

	for _, opt := range opts {
		opt(t)
	}

	cmd.PreAction(t.apply)
}

func (t *scannerConfig) apply(*kingpin.ParseContext) error {
	_, typName := extractPackageType(t.packageType)

	t.scannerName = defaultIfBlank(t.scannerName, fmt.Sprintf(t.defaultScannerNameFormat, typName))
	t.scannerRowName = defaultIfBlank(t.scannerRowName, fmt.Sprintf(t.defaultRowScannerNameFormat, typName))
	t.interfaceName = defaultIfBlank(t.interfaceName, fmt.Sprintf(t.defaultInterfaceNameFormat, typName))
	t.interfaceRowName = defaultIfBlank(t.interfaceRowName, fmt.Sprintf(t.defaultInterfaceRowNameFormat, typName))
	t.errScannerName = defaultIfBlank(t.errScannerName, fmt.Sprintf(t.defaultErrScannerNameFormat, typName))

	if t.private {
		t.scannerName = lowercaseFirstLetter(t.scannerName)
		t.scannerRowName = lowercaseFirstLetter(t.scannerRowName)
		t.interfaceName = lowercaseFirstLetter(t.interfaceName)
		t.interfaceRowName = lowercaseFirstLetter(t.interfaceRowName)
		t.errScannerName = lowercaseFirstLetter(t.errScannerName)
	} else {
		t.scannerName = strings.Title(t.scannerName)
		t.scannerRowName = strings.Title(t.scannerRowName)
		t.interfaceName = strings.Title(t.interfaceName)
		t.interfaceRowName = strings.Title(t.interfaceRowName)
		t.errScannerName = strings.Title(t.errScannerName)
	}

	return nil
}
