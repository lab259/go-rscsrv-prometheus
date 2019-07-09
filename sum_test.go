package boilerplate_test

import (
	. "github.com/lab259/go-package-boilerplate"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sum", func() {
	It("works", func() {
		Expect(Sum(1, 1)).To(Equal(2))
	})
})
