package goloquent_test

import (
	. "github.com/onsi/ginkgo"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Private function test", func() {
	Context("All name should be reserved", func() {
		It("should not error", func() {
			// Expect(goloquent.isNameReserved("name/id")).Should(BeFalse())
			// Expect(goloquent.isNameReserved("name/ID")).Should(BeFalse())
			// Expect(goloquent.isNameReserved("Name/ID")).Should(BeFalse())
			// Expect(goloquent.isNameReserved("Parent")).Should(BeFalse())
			// Expect(goloquent.isNameReserved("parent")).Should(BeFalse())
			// Expect(goloquent.isNameReserved(" Parent ")).Should(BeFalse())
		})
	})
})
