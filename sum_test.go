package promsrv_test

import (
	. "github.com/lab259/go-rscsrv-prometheus"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prometheus - Service", func() {
	It("works", func() {
		var srv Service

		Expect(srv).ToNot(BeNil())
	})
})
