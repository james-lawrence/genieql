package postgresql_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPostgresql(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgresql Suite")
}
