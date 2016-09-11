package sqlite3_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSqlite3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlite3 Suite")
}
