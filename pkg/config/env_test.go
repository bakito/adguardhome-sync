package config

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("env", func() {
		Context("enrichReplicasFromEnv", func() {
			It("should have the origin URL from the config env var", func() {
				_ = os.Setenv("REPLICA0_URL", "https://origin-env:443")
				defer func() {
					_ = os.Unsetenv("REPLICA0_URL")
				}()
				_, err := enrichReplicasFromEnv(nil)

				Ω(err).Should(HaveOccurred())
				Ω(err.Error()).Should(ContainSubstring("numbered replica env variables must have a number id >= 1"))
			})
		})
	})
})
