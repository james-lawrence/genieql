package testx

import (
	"bytes"
	"crypto/md5"
	"io"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TempDir() string {
	return ginkgo.GinkgoT().TempDir()
}

func Fixture(fixture string) []byte {
	return Must(os.ReadFile(fixture))
}

// Read a file at the given path.
func Read(path ...string) io.Reader {
	return bytes.NewReader(errorsx.Must(os.ReadFile(filepath.Join(path...))))
}

// ReadString from the given file.
func ReadString(path ...string) string {
	return string(errorsx.Must(os.ReadFile(filepath.Join(path...))))
}

// Must is a small language extension for panicing on the common
// value, error return pattern. only used in tests.
func Must[T any](v T, err error) T {
	gomega.Expect(err).To(gomega.BeNil())
	return v
}

// Tempenvvar temporarily set the environment variable.
func Tempenvvar(k, v string, do func()) {
	o := os.Getenv(k)
	defer os.Setenv(k, o)
	if err := os.Setenv(k, v); err != nil {
		panic(err)
	}
	do()
}

func IOMD5(in io.Reader) string {
	digester := md5.New()
	errorsx.Must(io.Copy(digester, in))
	return uuid.FromBytesOrNil(digester.Sum(nil)).String()
}

func IOString(in io.Reader) string {
	return string(errorsx.Must(io.ReadAll(in)))
}

func IOBytes(in io.Reader) []byte {
	return errorsx.Must(io.ReadAll(in))
}
