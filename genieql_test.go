package genieql_test

import (
	"bytes"
	"go/ast"

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
			Expect(FormatOutput(buffer, []byte(invalidCode))).To(MatchError("2:1: expected 'package', found 'func'"))
		})
	})

	Describe("TableDetails", func() {
		It("should filter out columns that do not match the provided fields", func() {
			details := TableDetails{
				Table:           "table",
				Naturalkey:      []string{"column1"},
				Columns:         []string{"column1", "column2", "column3"},
				UnmappedColumns: []string{},
			}

			filteredDetails := details.OnlyMappedColumns([]*ast.Field{}, AliasStrategyLowercase)
			Expect(filteredDetails.Columns).To(BeEmpty())
			Expect(filteredDetails.UnmappedColumns).To(Equal(details.Columns))

			fields := []*ast.Field{
				{Names: []*ast.Ident{&ast.Ident{Name: "column1"}}},
			}

			filteredDetails = details.OnlyMappedColumns(fields, AliasStrategyLowercase)
			Expect(filteredDetails.Columns).To(Equal([]string{"column1"}))
			Expect(filteredDetails.UnmappedColumns).To(Equal([]string{"column2", "column3"}))
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
