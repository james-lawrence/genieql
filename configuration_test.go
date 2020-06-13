package genieql_test

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	. "bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration", func() {
	Describe("ConfigurationFromURI", func() {
		It("should extract all fields from the URI", func() {
			uri, err := url.Parse("postgres://soandso:password@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionQueryer("sqlx.Queryer"),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Driver).To(Equal("github.com/lib/pq"))
			Expect(config.Dialect).To(Equal("postgres"))
			Expect(config.Database).To(Equal("databasename"))
			Expect(config.Host).To(Equal("localhost"))
			Expect(config.Port).To(Equal(5432))
			Expect(config.Username).To(Equal("soandso"))
			Expect(config.Password).To(Equal("password"))
			Expect(config.Queryer).To(Equal("sqlx.Queryer"))
		})

		It("should properly extract a URI without a password", func() {
			uri, err := url.Parse("postgres://soandso@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Driver).To(Equal("github.com/lib/pq"))
			Expect(config.Dialect).To(Equal("postgres"))
			Expect(config.Database).To(Equal("databasename"))
			Expect(config.Host).To(Equal("localhost"))
			Expect(config.Port).To(Equal(5432))
			Expect(config.Username).To(Equal("soandso"))
			Expect(config.Password).To(Equal(""))
		})

		It("should error if port is missing", func() {
			uri, err := url.Parse("postgres://soandso@localhost/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			_, err = NewConfiguration(
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(MatchError(ErrRequireHostAndPort))
		})

		It("should error if port is invalid", func() {
			_, err := url.Parse("postgres://soandso@localhost:abc1/databasename?sslmode=disable")
			Expect(err.Error()).To(Equal("parse \"postgres://soandso@localhost:abc1/databasename?sslmode=disable\": invalid port \":abc1\" after host"))
		})
	})

	Describe("Write and Read Configuration", func() {
		var tmpdir string
		var uri *url.URL

		BeforeEach(func() {
			var err error
			tmpdir, err = ioutil.TempDir(".", "bootstrap")
			Expect(err).ToNot(HaveOccurred())
			uri, err = url.Parse("postgres://soandso:password@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpdir)).ToNot(HaveOccurred())
		})

		It("should be able to write and read a configuration", func() {
			var (
				readConfig Configuration
			)
			readConfig, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionLocation(filepath.Join(tmpdir, "dummy.config")),
			)
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionLocation(filepath.Join(tmpdir, "dummy.config")),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(WriteConfiguration(config)).ToNot(HaveOccurred())

			Expect(ReadConfiguration(&readConfig)).ToNot(HaveOccurred())
			Expect(readConfig).To(Equal(config))
		})
	})

	Describe("Bootstrap", func() {
		var tmpdir string
		var uri *url.URL

		BeforeEach(func() {
			var err error
			tmpdir, err = ioutil.TempDir(".", "bootstrap")
			Expect(err).ToNot(HaveOccurred())
			uri, err = url.Parse("postgres://soandso:password@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpdir)).ToNot(HaveOccurred())
		})

		It("should write the config to the specified location", func() {
			path := filepath.Join(tmpdir, "dummy.config")

			err := Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionRowType("sqlx.Row"),
			)
			Expect(err).ToNot(HaveOccurred())

			raw, err := ioutil.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(raw)).To(Equal(fmt.Sprintf(exampleBootstrapConfiguration, tmpdir)))
		})

		It("should error if we can't write to the directory", func() {
			Expect(os.Chmod(tmpdir, 0444)).ToNot(HaveOccurred())
			path := filepath.Join(tmpdir, "dir", "dummy.config")

			err := Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(MatchError(fmt.Sprintf("failed to make bootstrap directory: mkdir %s: permission denied", filepath.Dir(path))))
		})

		It("should error if uri is invalid", func() {
			path := filepath.Join(tmpdir, "dummy.config")
			uri, err := url.Parse("postgres://soandso@localhost/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			err = Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/lib/pq"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(MatchError(ErrRequireHostAndPort))
		})
	})
})

const exampleBootstrapConfiguration = `location: %s
name: dummy.config
dialect: postgres
driver: github.com/lib/pq
queryer: '*sql.DB'
rowtype: sqlx.Row
connectionurl: postgres://soandso:password@localhost:5432/databasename?sslmode=disable
host: localhost
port: 5432
database: databasename
username: soandso
password: password
`
