package testx

import (
	"io"
	"log"
	"os"

	"github.com/mattn/go-isatty"

	. "github.com/onsi/gomega"
)

// Logging enable logging if stdout terminal is a tty.
// generally this means run the ginkgo without the -p (parallel) option.
func Logging() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.LUTC)
	log.SetOutput(os.Stderr)

	if isatty.IsTerminal(os.Stdout.Fd()) {
		return
	}

	log.SetOutput(io.Discard)
}

func MaybePanic(err error) {
	Expect(err).To(Succeed())
}
