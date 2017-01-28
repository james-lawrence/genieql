package genieql_test

import (
	"path/filepath"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	localdirectory string
	localfile      string
)

func TestGenieql(t *testing.T) {
	var (
		file string
		ok   bool
	)

	if _, file, _, ok = runtime.Caller(0); !ok {
		t.Error("failed to resolve file")
		t.FailNow()
	}
	localdirectory = filepath.Dir(file)
	localfile = filepath.Join(localdirectory, "foo.go")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Genieql Suite")
}
