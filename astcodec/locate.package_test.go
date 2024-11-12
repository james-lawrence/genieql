package astcodec_test

import (
	"go/build"

	"github.com/james-lawrence/genieql"
	. "github.com/james-lawrence/genieql/astcodec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Astutil", func() {
	Describe("LocatePackage", func() {
		It("find a the specified package", func() {
			var err error
			var p *build.Package

			p, err = LocatePackage("go/build", ".", build.Default, genieql.StrictPackageName("build"))
			Expect(err).ToNot(HaveOccurred())
			Expect(p.Name).To(Equal("build"))

			p, err = LocatePackage("does/not/exist", ".", build.Default, genieql.StrictPackageName("exist"))
			Expect(err).To(HaveOccurred())
			Expect(p).To(BeNil())
		})
	})
})
