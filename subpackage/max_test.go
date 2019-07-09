package subpackage_test

import (
	. "github.com/lab259/go-package-boilerplate/subpackage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Max", func() {
	It("works", func() {
		Expect(Max(1, 2, 100)).To(Equal(100))
	})
})
