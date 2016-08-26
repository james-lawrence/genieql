package generators_test

import (
	"bytes"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"

	"bitbucket.org/jatone/genieql"
	_ "bitbucket.org/jatone/genieql/internal/drivers"

	. "bitbucket.org/jatone/genieql/generators"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scanner", func() {
	config := genieql.MustConfiguration(
		genieql.ConfigurationOptionLocation(
			filepath.Join(genieql.ConfigurationDirectory(), "scanner-test.config"),
		),
	)
	genieql.RegisterDriver(config.Driver, noopDriver{})

	DescribeTable("should build a scanner for builtin types",
		func(definition, fixture string) {
			buffer := bytes.NewBuffer([]byte{})
			formatted := bytes.NewBuffer([]byte{})
			fset := token.NewFileSet()

			node, err := parser.ParseFile(fset, "example", definition, 0)
			Expect(err).ToNot(HaveOccurred())

			buffer.WriteString("package example\n\n")
			for _, d := range genieql.SelectFuncType(genieql.FindTypes(node)...) {
				for _, g := range ScannerFromGenDecl(d, ScannerOptionConfiguration(config)) {
					Expect(g.Generate(buffer)).ToNot(HaveOccurred())
					buffer.WriteString("\n")
				}
			}
			expected, err := ioutil.ReadFile(fixture)
			Expect(err).ToNot(HaveOccurred())
			Expect(genieql.FormatOutput(formatted, buffer.Bytes())).ToNot(HaveOccurred())
			Expect(formatted.Bytes()).To(Equal(expected))
		},
		Entry("scan int", `package example; type ExampleInt func(arg int)`, "_test_fixtures/int_scanner.go"),
		Entry("scan bool", `package example; type ExampleBool func(arg bool)`, "_test_fixtures/bool_scanner.go"),
		Entry("scan time.Time", `package example; type ExampleTime func(arg time.Time)`, "_test_fixtures/time_scanner.go"),
		Entry("scan multipleParams", `package example; type ExampleMultipleParam func(arg1, arg2 int, arg3 bool, arg4 string)`, "_test_fixtures/multiple_params_scanner.go"),
	)
})

const intScanner = `
`

const boolScanner = `package example

import "database/sql"

type example struct {
	Rows *sql.Rows
}

func (t example) Scan(arg *bool) error {
	var (
		c0 sql.NullBool
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	if c0.Valid {
		tmp := c0.Bool
		arg = &tmp
	}

	return t.Rows.Err()
}

func (t example) Err() error {
	return t.Rows.Err()
}

func (t example) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t example) Next() bool {
	return t.Rows.Next()
}
`

const timeScanner = `package example

import (
	"database/sql"
	"time"
)

type example struct {
	Rows *sql.Rows
}

func (t example) Scan(arg *time.Time) error {
	var (
		c0 time.Time
	)

	if err := t.Rows.Scan(&c0); err != nil {
		return err
	}

	arg = &c0

	return t.Rows.Err()
}

func (t example) Err() error {
	return t.Rows.Err()
}

func (t example) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t example) Next() bool {
	return t.Rows.Next()
}
`

const multiParamScanner = `package example

import "database/sql"

type example struct {
	Rows *sql.Rows
}

func (t example) Scan(arg1, arg2 *int, arg3 *bool, arg4 *string) error {
	var (
		c0 sql.NullInt64
		c1 sql.NullInt64
		c2 sql.NullBool
		c3 sql.NullString
	)

	if err := t.Rows.Scan(&c0, &c1, &c2, &c3); err != nil {
		return err
	}

	if c0.Valid {
		tmp := int(c0.Int64)
		arg1 = &tmp
	}

	if c1.Valid {
		tmp := int(c1.Int64)
		arg2 = &tmp
	}

	if c2.Valid {
		tmp := c2.Bool
		arg3 = &tmp
	}

	if c3.Valid {
		tmp := c3.String
		arg4 = &tmp
	}

	return t.Rows.Err()
}

func (t example) Err() error {
	return t.Rows.Err()
}

func (t example) Close() error {
	if t.Rows == nil {
		return nil
	}
	return t.Rows.Close()
}

func (t example) Next() bool {
	return t.Rows.Next()
}
`
