package astcodec_test

import (
	"bytes"
	"errors"

	"github.com/james-lawrence/genieql/astcodec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Formatting", func() {
	Describe("FormatOutput", func() {
		It("should format the code", func() {
			buffer := bytes.NewBuffer([]byte{})
			Expect(astcodec.FormatOutput(buffer, []byte(unformattedCode))).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(formattedCode))
		})

		It("should error when invalid code is provided", func() {
			buffer := bytes.NewBuffer([]byte{})
			err := errors.Unwrap(astcodec.FormatOutput(buffer, []byte(invalidCode)))
			Expect(err).To(MatchError("generated.go:2:1: expected 'package', found 'func'"))
		})
	})
})

const invalidCode = `
func HelloWorld() {
	fmt.Println("Hello World")
}
`
const unformattedCode = `
package test
import "fmt"

func HelloWorld() {
fmt.Println("Hello World")
}
`

const formattedCode = `package test

import "fmt"

func HelloWorld() {
	fmt.Println("Hello World")
}
`
