package types

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Types", func() {
	Context("AdGuardInstance", func() {
		var inst AdGuardInstance

		BeforeEach(func() {
			inst = AdGuardInstance{}
		})
		Context("Instance Init", func() {
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
	})
	Context("Config", func() {
		Context("init", func() {
			cfg := Config{
				Replicas: []AdGuardInstance{
					{URL: "https://localhost:3000"},
				},
			}
			err := cfg.Init()
			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.Replicas[0].Host).Should(Equal("localhost:3000"))
			Ω(cfg.Replicas[0].WebHost).Should(Equal("localhost:3000"))
			Ω(cfg.Replicas[0].URL).Should(Equal("https://localhost:3000"))
			Ω(cfg.Replicas[0].WebURL).Should(Equal("https://localhost:3000"))
		})
		Context("UniqueReplicas", func() {
			It("should return unique replicas in the array", func() {
				cfg := Config{
					Replicas: []AdGuardInstance{
						{URL: "a"},
						{URL: "a", APIPath: DefaultAPIPath},
						{URL: "a", APIPath: "foo"},
						{URL: "b", APIPath: DefaultAPIPath},
					},
					Replica: &AdGuardInstance{URL: "b", APIPath: DefaultAPIPath},
				}
				replicas := cfg.UniqueReplicas()
				Ω(replicas).Should(HaveLen(3))
			})
		})
		Context("mask", func() {
			It("should mask all names and passwords", func() {
				cfg := Config{
					Replicas: []AdGuardInstance{
						{URL: "a", Username: "user", Password: "pass"},
					},
					Replica: &AdGuardInstance{URL: "a", Username: "user", Password: "pass"},
					API:     API{Username: "user", Password: "pass"},
				}
				masked := cfg.mask()
				Ω(masked.Replicas[0].Username).Should(Equal("u**r"))
				Ω(masked.Replicas[0].Password).Should(Equal("p**s"))
				Ω(masked.Replica.Username).Should(Equal("u**r"))
				Ω(masked.Replica.Password).Should(Equal("p**s"))
				Ω(masked.API.Username).Should(Equal("u**r"))
				Ω(masked.API.Password).Should(Equal("p**s"))
			})
		})
	})
	Context("Feature", func() {
		Context("LogDisabled", func() {
			It("should log all features", func() {
				f := NewFeatures(false)
				Ω(f.collectDisabled()).Should(HaveLen(11))
			})
			It("should log no features", func() {
				f := NewFeatures(true)
				Ω(f.collectDisabled()).Should(BeEmpty())
			})
		})
	})
})
