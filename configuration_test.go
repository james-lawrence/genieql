package genieql_test

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	. "github.com/james-lawrence/genieql"
	"github.com/james-lawrence/genieql/internal/testx"
	"github.com/james-lawrence/genieql/internal/userx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration", func() {
	Describe("ConfigurationFromURI", func() {
		It("should extract all fields from the URI", func() {
			uri, err := url.Parse("postgres://soandso:password@localhost:5432/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionQueryer("sqlx.Queryer"),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Driver).To(Equal("github.com/jackc/pgx"))
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
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(config.Driver).To(Equal("github.com/jackc/pgx"))
			Expect(config.Dialect).To(Equal("postgres"))
			Expect(config.Database).To(Equal("databasename"))
			Expect(config.Host).To(Equal("localhost"))
			Expect(config.Port).To(Equal(5432))
			Expect(config.Username).To(Equal("soandso"))
			Expect(config.Password).To(Equal(""))
		})

		It("should handle missing ports", func() {
			uri, err := url.Parse("postgres://soandso@localhost/databasename?sslmode=disable")
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(Succeed())
			Expect(config.Driver).To(Equal("github.com/jackc/pgx"))
			Expect(config.Dialect).To(Equal("postgres"))
			Expect(config.Database).To(Equal("databasename"))
			Expect(config.Host).To(Equal("localhost"))
			Expect(config.Port).To(Equal(0))
			Expect(config.Username).To(Equal("soandso"))
			Expect(config.Password).To(Equal(""))
		})
	})

	Describe("Write and Read Configuration", func() {
		var tmpdir string
		var uri *url.URL

		BeforeEach(func() {
			var err error
			tmpdir, err = os.MkdirTemp(".", "bootstrap")
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
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionLocation(filepath.Join(tmpdir, "dummy.config")),
			)
			Expect(err).ToNot(HaveOccurred())
			config, err := NewConfiguration(
				ConfigurationOptionDriver("github.com/jackc/pgx"),
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
			tmpdir = testx.TempDir()
			uri = testx.Must(url.Parse("postgres://soandso:password@localhost:5432/databasename?sslmode=disable"))
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpdir)).ToNot(HaveOccurred())
		})

		It("should write the config to the specified location", func() {
			path := filepath.Join(tmpdir, "dummy.config")

			err := Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
				ConfigurationOptionRowType("sqlx.Row"),
			)
			Expect(err).ToNot(HaveOccurred())

			raw, err := os.ReadFile(path)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(raw)).To(Equal(exampleBootstrapConfiguration))
		})

		It("should error if we can't write to the directory", func() {
			// disable this test when running as root temporarily. need to fix ci/cd.
			// essentially this test will not work if root because permissions are ignored.
			if u := userx.CurrentUserOrDefault(userx.Root()); u.Username == "root" {
				return
			}
			Expect(os.Chmod(tmpdir, 0444)).To(Succeed())
			path := filepath.Join(tmpdir, "dir", "dummy.config")

			err := Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(MatchError(fmt.Sprintf("failed to make bootstrap directory: mkdir %s: permission denied", filepath.Dir(path))))
		})

		It("should error if uri is invalid", func() {
			path := filepath.Join(tmpdir, "dummy.config")
			uri := &url.URL{
				Scheme: "postgresq",
				Host:   "localhost:abc123",
				Path:   "databasename",
			}

			err := Bootstrap(
				ConfigurationOptionLocation(path),
				ConfigurationOptionDriver("github.com/jackc/pgx"),
				ConfigurationOptionDatabase(uri),
			)
			Expect(err).To(MatchError("strconv.Atoi: parsing \"abc123\": invalid syntax"))
		})
	})
})

const exampleBootstrapConfiguration = `name: dummy.config
dialect: postgres
driver: github.com/jackc/pgx
queryer: '*sql.DB'
rowtype: sqlx.Row
connectionurl: postgres://soandso:password@localhost:5432/databasename?sslmode=disable
host: localhost
port: 5432
database: databasename
username: soandso
password: password
`
