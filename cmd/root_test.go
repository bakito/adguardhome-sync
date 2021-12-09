package cmd

import (
	"os"

	"github.com/bakito/adguardhome-sync/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var envVars = []string{
	"FEATURES_GENERALSETTINGS",
	"FEATURES_QUERYLOGCONFIG",
	"FEATURES_STATSCONFIG",
	"FEATURES_CLIENTSETTINGS",
	"FEATURES_SERVICES",
	"FEATURES_FILTERS",
	"FEATURES_DHCP_SERVERCONFIG",
	"FEATURES_DHCP_STATICLEASES",
	"FEATURES_DNS_SERVERCONFIG",
	"FEATURES_DNS_ACCESSLISTS",
	"FEATURES_DNS_REWRITES",
}

var _ = Describe("Run", func() {
	BeforeEach(func() {
		for _, envVar := range envVars {
			Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
		}
		initConfig()
	})
	Context("getConfig", func() {
		It("features should be true by default", func() {
			cfg, err := getConfig()
			Ω(err).ShouldNot(HaveOccurred())
			verifyFeatures(cfg, true)
		})
		It("features should be false", func() {
			for _, envVar := range envVars {
				Ω(os.Setenv(envVar, "false")).ShouldNot(HaveOccurred())
			}
			cfg, err := getConfig()
			Ω(err).ShouldNot(HaveOccurred())
			verifyFeatures(cfg, false)
		})
	})
})

func verifyFeatures(cfg *types.Config, value bool) {
	Ω(cfg.Features.GeneralSettings).Should(Equal(value))
	Ω(cfg.Features.QueryLogConfig).Should(Equal(value))
	Ω(cfg.Features.StatsConfig).Should(Equal(value))
	Ω(cfg.Features.ClientSettings).Should(Equal(value))
	Ω(cfg.Features.Services).Should(Equal(value))
	Ω(cfg.Features.Filters).Should(Equal(value))
	Ω(cfg.Features.DHCP.ServerConfig).Should(Equal(value))
	Ω(cfg.Features.DHCP.StaticLeases).Should(Equal(value))
	Ω(cfg.Features.DNS.ServerConfig).Should(Equal(value))
	Ω(cfg.Features.DNS.AccessLists).Should(Equal(value))
	Ω(cfg.Features.DNS.Rewrites).Should(Equal(value))
}
