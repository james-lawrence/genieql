package generators_test

import (
	"bytes"

	"bitbucket.org/jatone/genieql"

	. "bitbucket.org/jatone/genieql/generators"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ColumnConstants", func() {
	DescribeTable("should create a constant based on the table details",
		func(name string, columns []genieql.ColumnInfo, trans genieql.ColumnTransformer, result string) {
			dst := bytes.NewBuffer([]byte{})
			Expect(NewColumnConstants(name, trans, columns).Generate(dst)).ToNot(HaveOccurred())
			Expect(dst.String()).To(Equal(result))
		},
		Entry(
			"simple columns",
			"constant",
			[]genieql.ColumnInfo{
				genieql.ColumnInfo{
					Name: "col1",
				},
				genieql.ColumnInfo{
					Name: "col2",
				},
			},
			genieql.NewColumnInfoNameTransformer(""),
			"const constant = `col1,col2`\n",
		),
	)
})
