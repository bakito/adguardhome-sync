package config

import (
	"os"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var ()
	BeforeEach(func() {
	})
	Context("configFilePath", func() {
		It("should return the same value", func() {
			path := uuid.NewString()
			result, err := configFilePath(path)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal(path))
		})
		It("should the file in HOME dir", func() {
			home := os.Getenv("HOME")
			defer func() {
				_ = os.Setenv("HOME", home)
			}()
			_ = os.Setenv("HOME", "/agh-test-home")
			result, err := configFilePath("")

			Ω(err).ShouldNot(HaveOccurred())
			Ω(result).Should(Equal("/agh-test-home/.config/.adguardhome-sync"))
		})
	})
})
