package columninfo_test

import (
	"unicode"

	"bitbucket.org/jatone/genieql"
	. "bitbucket.org/jatone/genieql/columninfo"
	"bitbucket.org/jatone/genieql/internal/transformx"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/serenize/snaker"
)

var _ = Describe("Rename", func() {
	DescribeTable("Examples",
		func(c genieql.ColumnInfo, m transform.Transformer, expected string) {
			r, err := Rename(c, m)
			Expect(err).To(Succeed())
			Expect(r).To(Equal(expected))
		},
		Entry(
			"example 1 - lowercase",
			genieql.ColumnInfo{
				Name: "FOOBAR",
			},
			runes.Map(unicode.ToLower),
			"foobar",
		),
		Entry(
			"example 2 - uppercase",
			genieql.ColumnInfo{
				Name: "foobar",
			},
			runes.Map(unicode.ToUpper),
			"FOOBAR",
		),
		Entry(
			"example 3 - snakecase",
			genieql.ColumnInfo{
				Name: "FooBar",
			},
			transformx.Full(snaker.CamelToSnake),
			"foo_bar",
		),
		Entry(
			"example 4 - camelcase",
			genieql.ColumnInfo{
				Name: "foo_bar",
			},
			transformx.Full(snaker.SnakeToCamel),
			"FooBar",
		),
		Entry(
			"example 5 - quoted",
			genieql.ColumnInfo{
				Name: "foo",
			},
			transformx.Wrap("\""),
			"\"foo\"",
		),
		Entry(
			"example 6 - prefixed",
			genieql.ColumnInfo{
				Name: "bar",
			},
			transformx.Prefix("foo."),
			"foo.bar",
		),
		Entry(
			"example 7 - camel cased, quoted, prefixed",
			genieql.ColumnInfo{
				Name: "foo_bar",
			},
			transform.Chain(
				transformx.Full(snaker.SnakeToCamel),
				transformx.Wrap("\""),
				transformx.Prefix("\"prefix\"."),
			),
			"\"prefix\".\"FooBar\"",
		),
		Entry(
			"example 8 - camelcase",
			genieql.ColumnInfo{
				Name: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			transformx.Full(snaker.SnakeToCamel),
			"Aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		),
	)
})
