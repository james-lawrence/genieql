package main

import (
	"fmt"

	"bitbucket.org/jatone/genieql/x/stringsx"

	"github.com/alecthomas/kingpin"
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

	t.scannerName = stringsx.DefaultIfBlank(t.scannerName, fmt.Sprintf(t.defaultScannerNameFormat, typName))
	t.scannerRowName = stringsx.DefaultIfBlank(t.scannerRowName, fmt.Sprintf(t.defaultRowScannerNameFormat, typName))
	t.interfaceName = stringsx.DefaultIfBlank(t.interfaceName, fmt.Sprintf(t.defaultInterfaceNameFormat, typName))
	t.interfaceRowName = stringsx.DefaultIfBlank(t.interfaceRowName, fmt.Sprintf(t.defaultInterfaceRowNameFormat, typName))
	t.errScannerName = stringsx.DefaultIfBlank(t.errScannerName, fmt.Sprintf(t.defaultErrScannerNameFormat, typName))

	if t.private {
		t.scannerName = stringsx.ToPrivate(t.scannerName)
		t.scannerRowName = stringsx.ToPrivate(t.scannerRowName)
		t.interfaceName = stringsx.ToPrivate(t.interfaceName)
		t.interfaceRowName = stringsx.ToPrivate(t.interfaceRowName)
		t.errScannerName = stringsx.ToPrivate(t.errScannerName)
	} else {
		t.scannerName = stringsx.ToPublic(t.scannerName)
		t.scannerRowName = stringsx.ToPublic(t.scannerRowName)
		t.interfaceName = stringsx.ToPublic(t.interfaceName)
		t.interfaceRowName = stringsx.ToPublic(t.interfaceRowName)
		t.errScannerName = stringsx.ToPublic(t.errScannerName)
	}

	return nil
}
