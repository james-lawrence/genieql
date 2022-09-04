package transformx_test

import (
	"log"
	"unicode"

	. "bitbucket.org/jatone/genieql/internal/transformx"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/serenize/snaker"
)

var _ = DescribeTable("Wrap",
	func(with, in, out string) {
		r, _, err := transform.String(Wrap(with), in)
		Expect(err).To(Succeed())
		Expect(r).To(Equal(out))
	},
	Entry("example 1", "'", "hello", "'hello'"),
)

var _ = DescribeTable("Full",
	func(with func(string) string, in, out string) {
		log.Println("input", len(in))
		r, _, err := transform.String(Full(with), in)
		Expect(err).To(Succeed())
		Expect(r).To(Equal(out))
	},
	Entry("example 1", snaker.CamelToSnake, "HelloWorld", "hello_world"),
	Entry(
		"example 2 - long input",
		runes.Map(unicode.ToUpper).String,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	),
)
