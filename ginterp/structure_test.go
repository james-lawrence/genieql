package ginterp_test

import (
	"bytes"
	"io"

	"github.com/james-lawrence/genieql/astcodec"
	"github.com/james-lawrence/genieql/genieqltest"
	. "github.com/james-lawrence/genieql/ginterp"
	"github.com/james-lawrence/genieql/internal/errorsx"
	"github.com/james-lawrence/genieql/internal/membufx"
	"github.com/james-lawrence/genieql/internal/testx"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Structure", func() {
	config := DialectConfig1()
	ctx, err := genieqltest.GeneratorContext(config)
	errorsx.MaybePanic(err)

	DescribeTable(
		"examples",
		func(in Structure, out io.Reader) {
			var (
				b         = bytes.NewBufferString("package example\n")
				formatted = bytes.NewBufferString("")
			)

			Expect(in.Generate(b)).To(Succeed())
			Expect(astcodec.FormatOutput(formatted, b.Bytes())).To(Succeed())
			Expect(formatted.String()).To(Equal(testx.IOString(out)))
		},
		Entry(
			"example 1 - basic structure",
			func() Structure {
				s := NewStructure(ctx, "StructureExample1", nil)
				s.From(s.Table("struct_a"))
				return s
			}(),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/structures/example.1.go"))),
		),
		Entry(
			"example 2 - ignored fields",
			func() Structure {
				s := NewStructure(ctx, "StructureExample2", nil)
				s.Ignore("b", "e")
				s.From(s.Table("struct_a"))
				return s
			}(),
			io.Reader(membufx.NewMemBuffer(testx.Fixture(".fixtures/structures/example.2.go"))),
		),
	)
})
