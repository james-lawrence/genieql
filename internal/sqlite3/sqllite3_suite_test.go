package sqlite3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSqlite3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sqlite3 Suite")
}
