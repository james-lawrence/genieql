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
			config, err := ConfigurationFromURI("github.com/lib/pq", uri)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Driver).To(Equal("github.com/lib/pq"))
			Expect(config.Dialect).To(Equal("postgres"))
			Expect(config.Database).To(Equal("databasename"))
			Expect(config.Host).To(Equal("localhost"))
			Expect(config.Port).To(Equal(5432))
			Expect(config.Username).To(Equal("soandso"))
			Expect(config.Password).To(Equal("password"))
		})

		It("should properly extract a URI without a password", func() {
			uri, err := url.Parse("postgres://soandso@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			config, err := ConfigurationFromURI("github.com/lib/pq", uri)
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
			_, err = ConfigurationFromURI("github.com/lib/pq", uri)
			Expect(err).To(MatchError(ErrRequireHostAndPort))
		})

		It("should error if port is invalid", func() {
			uri, err := url.Parse("postgres://soandso@localhost:abc1/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			_, err = ConfigurationFromURI("github.com/lib/pq", uri)
			Expect(err.Error()).To(Equal("strconv.ParseInt: parsing \"abc1\": invalid syntax"))
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
			var readConfig Configuration
			path := filepath.Join(tmpdir, "dummy.config")
			config, err := ConfigurationFromURI("github.com/lib/pq", uri)
			Expect(err).ToNot(HaveOccurred())
			Expect(WriteConfiguration(path, config)).ToNot(HaveOccurred())

			Expect(ReadConfiguration(path, &readConfig)).ToNot(HaveOccurred())
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

			err := Bootstrap(path, "github.com/lib/pq", uri)
			Expect(err).ToNot(HaveOccurred())

			raw, err := ioutil.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(raw)).To(Equal(exampleBootstrapConfiguration))
		})

		It("should error if we can't write to the directory", func() {
			Expect(os.Chmod(tmpdir, 0444)).ToNot(HaveOccurred())
			path := filepath.Join(tmpdir, "dir", "dummy.config")

			err := Bootstrap(path, "github.com/lib/pq", uri)
			Expect(err).To(MatchError(fmt.Sprintf("mkdir %s: permission denied", filepath.Dir(path))))
		})

		It("should error if uri is invalid", func() {
			path := filepath.Join(tmpdir, "dummy.config")
			uri, err := url.Parse("postgres://soandso@localhost/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			err = Bootstrap(path, "github.com/lib/pq", uri)
			Expect(err).To(MatchError(ErrRequireHostAndPort))
		})
	})
})

const exampleBootstrapConfiguration = `dialect: postgres
driver: github.com/lib/pq
connectionurl: postgres://soandso:password@localhost:5432/databasename?sslmode=disable
host: localhost
port: 5432
database: databasename
username: soandso
password: password
`
