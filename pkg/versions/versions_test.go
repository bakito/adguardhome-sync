package versions_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bakito/adguardhome-sync/pkg/versions"
)

var _ = Describe("Versions", func() {
	Context("IsNewerThan", func() {
		It("should correctly parse json", func() {
			Ω(versions.IsNewerThan("v0.106.10", "v0.106.9")).Should(BeTrue())
			Ω(versions.IsNewerThan("v0.106.9", "v0.106.10")).Should(BeFalse())
			Ω(versions.IsNewerThan("v0.106.10", "0.106.9")).Should(BeTrue())
			Ω(versions.IsNewerThan("v0.106.9", "0.106.10")).Should(BeFalse())
		})
	})
	Context("IsSame", func() {
		It("should be the same version", func() {
			Ω(versions.IsSame("v0.106.9", "v0.106.9")).Should(BeTrue())
			Ω(versions.IsSame("0.106.9", "v0.106.9")).Should(BeTrue())
		})
	})
})
