package config_test

import (
	"fmt"
	"os"

	"github.com/bakito/adguardhome-sync/pkg/config"
	"github.com/bakito/adguardhome-sync/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var envVars = []string{
	"FEATURES_GENERAL_SETTINGS",
	"FEATURES_QUERY_LOG_CONFIG",
	"FEATURES_STATS_CONFIG",
	"FEATURES_CLIENT_SETTINGS",
	"FEATURES_SERVICES",
	"FEATURES_FILTERS",
	"FEATURES_DHCP_SERVER_CONFIG",
	"FEATURES_DHCP_STATIC_LEASES",
	"FEATURES_DNS_SERVER_CONFIG",
	"FEATURES_DNS_ACCESS_LISTS",
	"FEATURES_DNS_REWRITES",
	"REPLICA1_INTERFACE_NAME",
	"REPLICA1_DHCP_SERVER_ENABLED",
}

var deprecatedEnvVars = []string{
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
	"REPLICA1_INTERFACENAME",
	"REPLICA1_DHCPSERVERENABLED",
}

var _ = Describe("Run", func() {
	Context("deprecated", func() {
		BeforeEach(func() {
			for _, envVar := range deprecatedEnvVars {
				Ω(os.Setenv(envVar, "false")).ShouldNot(HaveOccurred())
			}
		})
		AfterEach(func() {
			for _, envVar := range deprecatedEnvVars {
				Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
			}
		})
		Context("Get", func() {
			It("features should be false", func() {
				cfg, err := config.Get("", nil)
				Ω(err).ShouldNot(HaveOccurred())
				verifyFeatures(cfg, false)
			})
		})
	})
	Context("current", func() {
		BeforeEach(func() {
			for _, envVar := range envVars {
				Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
			}
		})
		AfterEach(func() {
			for _, envVar := range envVars {
				Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
			}
		})
		Context("Get", func() {
			It("features should be true by default", func() {
				cfg, err := config.Get("", nil)
				Ω(err).ShouldNot(HaveOccurred())
				verifyFeatures(cfg, true)
			})
			It("features should be true by default", func() {
				cfg, err := config.Get("", nil)
				Ω(err).ShouldNot(HaveOccurred())
				verifyFeatures(cfg, true)
			})
			It("features should be false", func() {
				for _, envVar := range envVars {
					Ω(os.Setenv(envVar, "false")).ShouldNot(HaveOccurred())
				}
				cfg, err := config.Get("", nil)
				Ω(err).ShouldNot(HaveOccurred())
				verifyFeatures(cfg, false)
			})
			Context("interface name", func() {
				It("should set interface name of replica 1", func() {
					Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
					Ω(os.Setenv(fmt.Sprintf("REPLICA%s_INTERFACE_NAME", "1"), "eth0")).ShouldNot(HaveOccurred())
					cfg, err := config.Get("", nil)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Replicas[0].InterfaceName).Should(Equal("eth0"))
				})
			})
			Context("dhcp server", func() {
				It("should enable the dhcp server of replica 1", func() {
					Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
					Ω(os.Setenv(fmt.Sprintf("REPLICA%s_DHCPSERVERENABLED", "1"), "true")).ShouldNot(HaveOccurred())
					cfg, err := config.Get("", nil)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
					Ω(*cfg.Replicas[0].DHCPServerEnabled).Should(BeTrue())
				})
				It("should disable the dhcp server of replica 1", func() {
					Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
					Ω(os.Setenv(fmt.Sprintf("REPLICA%s_DHCPSERVERENABLED", "1"), "false")).ShouldNot(HaveOccurred())
					cfg, err := config.Get("", nil)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
					Ω(*cfg.Replicas[0].DHCPServerEnabled).Should(BeFalse())
				})
			})
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
