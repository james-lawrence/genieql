package astutil_test

import (
	. "bitbucket.org/jatone/genieql/astutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Astutil", func() {
	Describe("ExprList", func() {
		It("should work with no arguments", func() {
			Expect(ExprList()).To(BeEmpty())
		})
	})
})
