package integration_tests_test

import (
	. "bitbucket.org/jatone/genieql/scanner/internal/integration_tests"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"time"
)

var exampleTime, _ = time.Parse(time.RFC3339, "2016-03-15T02:51:14Z")

var table = []Type1{
	Type1{Field1: "Hello1", Field3: false, Field5: 1, Field7: exampleTime, Field8: newTime(exampleTime)},
	Type1{Field1: "Hello2", Field2: newString("World"), Field3: false, Field4: newBool(true), Field5: 1, Field6: newInt(2), Field7: exampleTime, Field8: newTime(exampleTime)},
}

var _ = Describe("Type1", func() {
	Describe("RowScanner", func() {
		It("should be able to scan a result", func() {
			for _, testEntry := range table {
				var found Type1
				scanner := NewType1RowScanner(TX.QueryRow(Type1Insert, explode(testEntry)...))
				Expect(scanner.Scan(&found)).ToNot(HaveOccurred())
				Expect(found.Field1).To(Equal(testEntry.Field1), testEntry.Field1)
				Expect(found.Field2).To(Equal(testEntry.Field2), testEntry.Field1)
				Expect(found.Field3).To(Equal(testEntry.Field3), testEntry.Field1)
				Expect(found.Field4).To(Equal(testEntry.Field4), testEntry.Field1)
				Expect(found.Field5).To(Equal(testEntry.Field5), testEntry.Field1)
				Expect(found.Field6).To(Equal(testEntry.Field6), testEntry.Field1)
				Expect(found.Field7.Unix()).To(Equal(testEntry.Field7.Unix()), testEntry.Field1)
				Expect(found.Field8.Unix()).To(Equal(testEntry.Field8.Unix()), testEntry.Field1)
			}
		})
	})

	Describe("Scanner", func() {
		It("should be able to scan a result", func() {
			for _, testEntry := range table {
				var found Type1
				scanner := NewType1Scanner(TX.Query(Type1Insert, explode(testEntry)...))
				Expect(scanner.Next()).To(BeTrue())
				err := scanner.Scan(&found)
				Expect(scanner.Close()).ToNot(HaveOccurred())
				Expect(err).ToNot(HaveOccurred())
				Expect(found.Field1).To(Equal(testEntry.Field1), testEntry.Field1)
				Expect(found.Field2).To(Equal(testEntry.Field2), testEntry.Field1)
				Expect(found.Field3).To(Equal(testEntry.Field3), testEntry.Field1)
				Expect(found.Field4).To(Equal(testEntry.Field4), testEntry.Field1)
				Expect(found.Field5).To(Equal(testEntry.Field5), testEntry.Field1)
				Expect(found.Field6).To(Equal(testEntry.Field6), testEntry.Field1)
				Expect(found.Field7.Unix()).To(Equal(testEntry.Field7.Unix()), testEntry.Field1)
				Expect(found.Field8.Unix()).To(Equal(testEntry.Field8.Unix()), testEntry.Field1)
			}
		})
	})
})

func explode(t Type1) []interface{} {
	return []interface{}{t.Field1, t.Field2, t.Field3, t.Field4, t.Field5, t.Field6, t.Field7, t.Field8}
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
