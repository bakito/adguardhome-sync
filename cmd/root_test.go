package cmd

import (
	"fmt"
	"os"

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
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
	"REPLICA1_INTERFACENAME",
	"REPLICA1_DHCPSERVERENABLED",
}

var _ = Describe("Run", func() {
	var logger *zap.SugaredLogger
	BeforeEach(func() {
		logger = log.GetLogger("root")
		for _, envVar := range envVars {
			Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
		}
		initConfig()
	})
	AfterEach(func() {
		for _, envVar := range envVars {
			Ω(os.Unsetenv(envVar)).ShouldNot(HaveOccurred())
		}
	})
	Context("getConfig", func() {
		It("features should be true by default", func() {
			cfg, err := getConfig(logger)
			Ω(err).ShouldNot(HaveOccurred())
			verifyFeatures(cfg, true)
		})
		It("features should be false", func() {
			for _, envVar := range envVars {
				Ω(os.Setenv(envVar, "false")).ShouldNot(HaveOccurred())
			}
			cfg, err := getConfig(logger)
			Ω(err).ShouldNot(HaveOccurred())
			verifyFeatures(cfg, false)
		})
		Context("interface name", func() {
			It("should set interface name of replica 1", func() {
				Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
				Ω(os.Setenv(fmt.Sprintf(envReplicasInterfaceName, "1"), "eth0")).ShouldNot(HaveOccurred())
				cfg, err := getConfig(logger)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Replicas[0].InterfaceName).Should(Equal("eth0"))
			})
		})
		Context("dhcp server", func() {
			It("should enable the dhcp server of replica 1", func() {
				Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
				Ω(os.Setenv(fmt.Sprintf(envDHCPServerEnabled, "1"), "true")).ShouldNot(HaveOccurred())
				cfg, err := getConfig(logger)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
				Ω(*cfg.Replicas[0].DHCPServerEnabled).Should(BeTrue())
			})
			It("should disable the dhcp server of replica 1", func() {
				Ω(os.Setenv("REPLICA1_URL", "https://foo.bar")).ShouldNot(HaveOccurred())
				Ω(os.Setenv(fmt.Sprintf(envDHCPServerEnabled, "1"), "false")).ShouldNot(HaveOccurred())
				cfg, err := getConfig(logger)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
				Ω(*cfg.Replicas[0].DHCPServerEnabled).Should(BeFalse())
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
