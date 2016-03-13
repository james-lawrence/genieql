package integration_tests_test

import (
	. "bitbucket.org/jatone/genieql/scanner/internal/integration_tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"
)

var table = []Type1{
	Type1{Field1: "Hello", Field3: false, Field5: 1},
	Type1{Field1: "Hello", Field2: newString("World"), Field3: false, Field4: newBool(true), Field5: 1, Field6: newInt(2)},
}

var _ = Describe("Type1", func() {
	Describe("RowScanner", func() {
		It("should be able to scan a result", func() {
			for _, testEntry := range table {
				var found Type1
				scanner := NewType1RowScanner(TX.QueryRow(Type1Insert, explode(testEntry)...))
				Expect(scanner.Scan(&found)).ToNot(HaveOccurred())
				Expect(found).To(Equal(testEntry))
			}
		})
	})

	Describe("Scanner", func() {
		It("should be able to scan a result", func() {
			for _, testEntry := range table {
				var found Type1
				scanner := NewType1Scanner(TX.Query(Type1Insert, explode(testEntry)...))
				err := scanner.Scan(&found)
				Expect(scanner.Close()).ToNot(HaveOccurred())
				Expect(err).ToNot(HaveOccurred())
				Expect(found).To(Equal(testEntry))
			}
		})
	})
})

func explode(t Type1) []interface{} {
	return []interface{}{t.Field1, t.Field2, t.Field3, t.Field4, t.Field5, t.Field6}
}

func newTime(t time.Time) *time.Time {
	return &t
}

func newString(s string) *string {
	return &s
}

func newBool(b bool) *bool {
	return &b
}

func newInt(i int) *int {
	return &i
}
