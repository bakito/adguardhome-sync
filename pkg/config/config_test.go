package config_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/pkg/config"
	flagsmock "github.com/bakito/adguardhome-sync/pkg/mocks/flags"
)

var _ = Describe("Config", func() {
	Context("Get", func() {
		var (
			flags          *flagsmock.MockFlags
			mockCtrl       *gm.Controller
			changedEnvVars []string
			setEnv         = func(name, value string) {
				_ = os.Setenv(name, value)
				changedEnvVars = append(changedEnvVars, name)
			}
		)
		BeforeEach(func() {
			mockCtrl = gm.NewController(GinkgoT())
			flags = flagsmock.NewMockFlags(mockCtrl)
			changedEnvVars = nil
		})
		AfterEach(func() {
			for _, envVar := range changedEnvVars {
				_ = os.Unsetenv(envVar)
				println(envVar)
			}
			defer mockCtrl.Finish()
		})
		Context("Get", func() {
			Context("Mixed Config", func() {
				It("should have the origin URL from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					_, err := config.Get("../../testdata/config_test_replicas_and_replica.yaml", flags)
					Ω(err).Should(HaveOccurred())
					Ω(err.Error()).Should(ContainSubstring("mixed replica config in use"))
				})
			})
			Context("Origin Url", func() {
				It("should have the origin URL from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Origin.URL).Should(Equal("https://origin-file:443"))
				})
				It("should have the origin URL from the config flags", func() {
					flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Origin.URL).Should(Equal("https://origin-flag:443"))
				})
				It("should have the origin URL from the config env var", func() {
					setEnv("ORIGIN_URL", "https://origin-env:443")
					flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Origin.URL).Should(Equal("https://origin-env:443"))
				})
			})
			Context("Replica insecure skip verify", func() {
				It("should have the insecure skip verify from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].InsecureSkipVerify).Should(BeFalse())
				})
				It("should have the insecure skip verify from the config flags", func() {
					flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].InsecureSkipVerify).Should(BeTrue())
				})
				It("should have the insecure skip verify from the config env var", func() {
					setEnv("REPLICA_INSECURE_SKIP_VERIFY", "false")
					flags.EXPECT().Changed(config.FlagReplicaISV).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetBool(config.FlagReplicaISV).Return(true, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].InsecureSkipVerify).Should(BeFalse())
				})
			})

			Context("Replica 1 insecure skip verify", func() {
				It("should have the insecure skip verify from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].InsecureSkipVerify).Should(BeFalse())
				})
				It("should have the insecure skip verify from the config env var", func() {
					setEnv("REPLICA1_INSECURE_SKIP_VERIFY", "true")
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].InsecureSkipVerify).Should(BeTrue())
				})
			})
			Context("API Port", func() {
				It("should have the api port from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9090))
				})
				It("should have the api port from the config flags", func() {
					flags.EXPECT().Changed(config.FlagApiPort).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetInt(config.FlagApiPort).Return(9990, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9990))
				})
				It("should have the api port from the config env var", func() {
					setEnv("API_PORT", "9999")
					flags.EXPECT().Changed(config.FlagApiPort).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetInt(config.FlagApiPort).Return(9990, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9999))
				})
			})

			Context("Replica DHCPServerEnabled", func() {
				It("should have the dhcp server enabled from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replica.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
					Ω(*cfg.Get().Replicas[0].DHCPServerEnabled).Should(BeFalse())
				})
			})

			Context("Replica 1 DHCPServerEnabled", func() {
				It("should have the dhcp server enabled from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Replicas[0].DHCPServerEnabled).ShouldNot(BeNil())
					Ω(*cfg.Get().Replicas[0].DHCPServerEnabled).Should(BeFalse())
				})
			})
			Context("API Port", func() {
				It("should have the api port from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9090))
				})
				It("should have the api port from the config flags", func() {
					flags.EXPECT().Changed(config.FlagApiPort).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetInt(config.FlagApiPort).Return(9990, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9990))
				})
				It("should have the api port from the config env var", func() {
					setEnv("API_PORT", "9999")
					flags.EXPECT().Changed(config.FlagApiPort).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetInt(config.FlagApiPort).Return(9990, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().API.Port).Should(Equal(9999))
				})
			})

			Context("Feature DNS Server Config", func() {
				It("should have the feature dns server config from the config file", func() {
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Features.DNS.ServerConfig).Should(BeFalse())
				})
				It("should have the feature dns server config from the config flags", func() {
					flags.EXPECT().Changed(config.FlagFeatureDnsServerConfig).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetBool(config.FlagFeatureDnsServerConfig).Return(true, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Features.DNS.ServerConfig).Should(BeTrue())
				})
				It("should have the feature dns server config from the config env var", func() {
					setEnv("FEATURES_DNS_SERVER_CONFIG", "false")
					flags.EXPECT().Changed(config.FlagFeatureDnsServerConfig).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetBool(config.FlagFeatureDnsServerConfig).Return(true, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Features.DNS.ServerConfig).Should(BeFalse())
				})
				It("should have the feature dns server config from the config DEPRECATED env var", func() {
					setEnv("FEATURES_DNS_SERVERCONFIG", "false")
					flags.EXPECT().Changed(config.FlagFeatureDnsServerConfig).Return(true).AnyTimes()
					flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
					flags.EXPECT().GetBool(config.FlagFeatureDnsServerConfig).Return(true, nil).AnyTimes()

					cfg, err := config.Get("../../testdata/config_test_replicas.yaml", flags)
					Ω(err).ShouldNot(HaveOccurred())
					Ω(cfg.Get().Features.DNS.ServerConfig).Should(BeFalse())
				})
			})
		})
	})
})
