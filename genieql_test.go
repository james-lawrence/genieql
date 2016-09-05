package genieql_test

import (
	"bytes"
	"go/ast"

	"github.com/pkg/errors"

	. "bitbucket.org/jatone/genieql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Genieql", func() {
	Describe("FormatOutput", func() {
		It("should format the code", func() {
			buffer := bytes.NewBuffer([]byte{})
			Expect(FormatOutput(buffer, []byte(unformattedCode))).ToNot(HaveOccurred())
			Expect(buffer.String()).To(Equal(formattedCode))
		})

		It("should error when invalid code is provided", func() {
			buffer := bytes.NewBuffer([]byte{})
			err := errors.Cause(FormatOutput(buffer, []byte(invalidCode)))
			Expect(err).To(MatchError("2:1: expected 'package', found 'func'"))
		})
	})

	Describe("TableDetails", func() {
		It("should filter out columns that do not match the provided fields", func() {
			details := TableDetails{
				Table:           "table",
				Naturalkey:      []ColumnInfo{{Name: "column1"}},
				Columns:         []ColumnInfo{{Name: "column1"}, {Name: "column2"}, {Name: "column3"}},
				UnmappedColumns: []ColumnInfo{},
			}

			filteredDetails := details.OnlyMappedColumns([]*ast.Field{}, AliasStrategyLowercase)
			Expect(filteredDetails.Columns).To(BeEmpty())
			Expect(filteredDetails.UnmappedColumns).To(Equal(details.Columns))

			fields := []*ast.Field{
				{Names: []*ast.Ident{&ast.Ident{Name: "column1"}}},
			}

			filteredDetails = details.OnlyMappedColumns(fields, AliasStrategyLowercase)
			Expect(filteredDetails.Columns).To(Equal([]ColumnInfo{{Name: "column1"}}))
			Expect(filteredDetails.UnmappedColumns).To(Equal([]ColumnInfo{{Name: "column2"}, {Name: "column3"}}))
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
