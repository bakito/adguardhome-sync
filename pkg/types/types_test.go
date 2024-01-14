package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AdGuardInstance", func() {
	var inst AdGuardInstance

	BeforeEach(func() {
		inst = AdGuardInstance{}
	})
	Context("Init", func() {
		BeforeEach(func() {
			inst.URL = "https://localhost:3000"
		})
		It("should correctly set Host and WebHost if only URL is set", func() {
			err := inst.Init()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(inst.Host).Should(Equal("localhost:3000"))
			Ω(inst.WebHost).Should(Equal("localhost:3000"))
			Ω(inst.URL).Should(Equal("https://localhost:3000"))
			Ω(inst.WebURL).Should(Equal("https://localhost:3000"))
		})
		It("should correctly set Host and WebHost if URL and WebURL are set", func() {
			inst.WebURL = "https://127.0.0.1:4000"
			err := inst.Init()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(inst.Host).Should(Equal("localhost:3000"))
			Ω(inst.WebHost).Should(Equal("127.0.0.1:4000"))
			Ω(inst.WebURL).Should(Equal(inst.WebURL))
			Ω(inst.URL).Should(Equal("https://localhost:3000"))
			Ω(inst.WebURL).Should(Equal("https://127.0.0.1:4000"))
		})
	})
	Context("LogDisabled", func() {
		It("It should log all features", func() {
			f := NewFeatures(false)
			Ω(f.collectDisabled()).Should(HaveLen(11))
		})
		It("It should log no features", func() {
			f := NewFeatures(true)
			Ω(f.collectDisabled()).Should(BeEmpty())
		})
	})
})
