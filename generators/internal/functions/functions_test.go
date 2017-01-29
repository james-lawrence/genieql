package functions_test

import (
	. "bitbucket.org/jatone/genieql/generators/internal/functions"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Functions", func() {
	DescribeTable("Batch Insert",
		func(examples ...Example1) {
			const (
				countQuery = `SELECT COUNT(*) FROM example1`
			)
			var (
				count   int
				scan    Example1
				results []Example1
			)

			scanner := NewExample1BatchInsertFunction(TX, examples...)
			defer scanner.Close()
			for scanner.Next() {
				Expect(scanner.Scan(&scan)).ToNot(HaveOccurred())
				results = append(results, scan)
			}
			Expect(scanner.Err()).ToNot(HaveOccurred())
			Expect(results).To(HaveLen(len(examples)))
			Expect(TX.QueryRow(countQuery).Scan(&count)).ToNot(HaveOccurred())
			Expect(count).To(Equal(len(examples)))
		},
		Entry("insert empty list"),
		Entry(
			"insert 1 item",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
		),
		Entry(
			"insert 2 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
		),
		Entry(
			"insert 3 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000002"},
		),
		Entry(
			"insert 4 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000002"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000003"},
		),
		Entry(
			"insert 5 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000002"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000003"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000004"},
		),
		Entry(
			"insert 6 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000002"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000003"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000004"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000005"},
		),
		Entry(
			"insert 7 items",
			Example1{UUIDField: "00000000-0000-0000-0000-000000000000"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000001"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000002"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000003"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000004"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000005"},
			Example1{UUIDField: "00000000-0000-0000-0000-000000000006"},
		),
	)
})
