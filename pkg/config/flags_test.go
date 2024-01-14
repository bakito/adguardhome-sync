package config

import (
	"strings"

	flagsmock "github.com/bakito/adguardhome-sync/pkg/mocks/flags"
	"github.com/bakito/adguardhome-sync/pkg/types"
	gm "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		cfg      *types.Config
		flags    *flagsmock.MockFlags
		mockCtrl *gm.Controller
	)
	BeforeEach(func() {
		cfg = &types.Config{
			Replica: &types.AdGuardInstance{},
			Features: types.Features{
				DNS: types.DNS{
					AccessLists:  true,
					ServerConfig: true,
					Rewrites:     true,
				},
				DHCP: types.DHCP{
					ServerConfig: true,
					StaticLeases: true,
				},
				GeneralSettings: true,
				QueryLogConfig:  true,
				StatsConfig:     true,
				ClientSettings:  true,
				Services:        true,
				Filters:         true,
			},
		}
		mockCtrl = gm.NewController(GinkgoT())
		flags = flagsmock.NewMockFlags(mockCtrl)
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})
	Context("readFlags", func() {
		It("should not change the config with nil flags", func() {
			clone := cfg.DeepCopy()
			err := readFlags(cfg, nil)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg).Should(Equal(clone))
		})
		It("should not change the config with no changed flags", func() {
			clone := cfg.DeepCopy()
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg).Should(Equal(clone))
		})
	})
	Context("readFeatureFlags", func() {
		It("should disable all flags", func() {
			flags.EXPECT().Changed(gm.Any()).DoAndReturn(func(name string) bool {
				return strings.HasPrefix(name, "feature")
			}).AnyTimes()
			flags.EXPECT().GetBool(gm.Any()).Return(false, nil).AnyTimes()
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.Features).Should(Equal(types.Features{
				DNS: types.DNS{
					AccessLists:  false,
					ServerConfig: false,
					Rewrites:     false,
				},
				DHCP: types.DHCP{
					ServerConfig: false,
					StaticLeases: false,
				},
				GeneralSettings: false,
				QueryLogConfig:  false,
				StatsConfig:     false,
				ClientSettings:  false,
				Services:        false,
				Filters:         false,
			}))
		})
	})
	Context("readApiFlags", func() {
		It("should change all values", func() {
			cfg.API = types.API{
				Port:     1111,
				Username: "2222",
				Password: "3333",
				DarkMode: false,
			}
			flags.EXPECT().Changed(gm.Any()).DoAndReturn(func(name string) bool {
				return strings.HasPrefix(name, "api")
			}).AnyTimes()
			flags.EXPECT().GetInt(FlagApiPort).Return(9999, nil)
			flags.EXPECT().GetString(FlagApiUsername).Return("aaaa", nil)
			flags.EXPECT().GetString(FlagApiPassword).Return("bbbb", nil)
			flags.EXPECT().GetBool(FlagApiDarkMode).Return(true, nil)
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.API).Should(Equal(types.API{
				Port:     9999,
				Username: "aaaa",
				Password: "bbbb",
				DarkMode: true,
			}))
		})
	})
	Context("readRootFlags", func() {
		It("should change all values", func() {
			cfg.Cron = "*/10 * * * *"
			cfg.PrintConfigOnly = false
			cfg.ContinueOnError = false
			cfg.RunOnStart = false

			flags.EXPECT().Changed(FlagCron).Return(true)
			flags.EXPECT().Changed(FlagRunOnStart).Return(true)
			flags.EXPECT().Changed(FlagPrintConfigOnly).Return(true)
			flags.EXPECT().Changed(FlagContinueOnError).Return(true)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			flags.EXPECT().GetString(FlagCron).Return("*/30 * * * *", nil)
			flags.EXPECT().GetBool(FlagRunOnStart).Return(true, nil)
			flags.EXPECT().GetBool(FlagPrintConfigOnly).Return(true, nil)
			flags.EXPECT().GetBool(FlagContinueOnError).Return(true, nil)
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.Cron).Should(Equal("*/30 * * * *"))
			Ω(cfg.RunOnStart).Should(BeTrue())
			Ω(cfg.PrintConfigOnly).Should(BeTrue())
			Ω(cfg.ContinueOnError).Should(BeTrue())
		})
	})
	Context("readOriginFlags", func() {
		It("should change all values", func() {
			cfg.Origin = types.AdGuardInstance{
				URL:                "1",
				WebURL:             "2",
				APIPath:            "3",
				Username:           "4",
				Password:           "5",
				Cookie:             "6",
				InsecureSkipVerify: false,
			}

			flags.EXPECT().Changed(FlagOriginURL).Return(true)
			flags.EXPECT().Changed(FlagOriginWebURL).Return(true)
			flags.EXPECT().Changed(FlagOriginApiPath).Return(true)
			flags.EXPECT().Changed(FlagOriginUsername).Return(true)
			flags.EXPECT().Changed(FlagOriginPassword).Return(true)
			flags.EXPECT().Changed(FlagOriginCookie).Return(true)
			flags.EXPECT().Changed(FlagOriginISV).Return(true)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			flags.EXPECT().GetString(FlagOriginURL).Return("a", nil)
			flags.EXPECT().GetString(FlagOriginWebURL).Return("b", nil)
			flags.EXPECT().GetString(FlagOriginApiPath).Return("c", nil)
			flags.EXPECT().GetString(FlagOriginUsername).Return("d", nil)
			flags.EXPECT().GetString(FlagOriginPassword).Return("e", nil)
			flags.EXPECT().GetString(FlagOriginCookie).Return("f", nil)
			flags.EXPECT().GetBool(FlagOriginISV).Return(true, nil)
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.Origin).Should(Equal(types.AdGuardInstance{
				URL:                "a",
				WebURL:             "b",
				APIPath:            "c",
				Username:           "d",
				Password:           "e",
				Cookie:             "f",
				InsecureSkipVerify: true,
			}))
		})
	})
	Context("readReplicaFlags", func() {
		It("should change all values", func() {
			cfg.Replica = &types.AdGuardInstance{
				URL:                "1",
				WebURL:             "2",
				APIPath:            "3",
				Username:           "4",
				Password:           "5",
				Cookie:             "6",
				InsecureSkipVerify: false,
				AutoSetup:          false,
				InterfaceName:      "7",
			}

			flags.EXPECT().Changed(FlagReplicaURL).Return(true)
			flags.EXPECT().Changed(FlagReplicaWebURL).Return(true)
			flags.EXPECT().Changed(FlagReplicaApiPath).Return(true)
			flags.EXPECT().Changed(FlagReplicaUsername).Return(true)
			flags.EXPECT().Changed(FlagReplicaPassword).Return(true)
			flags.EXPECT().Changed(FlagReplicaCookie).Return(true)
			flags.EXPECT().Changed(FlagReplicaISV).Return(true)
			flags.EXPECT().Changed(FlagReplicaAutoSetup).Return(true)
			flags.EXPECT().Changed(FlagReplicaInterfaceName).Return(true)
			flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

			flags.EXPECT().GetString(FlagReplicaURL).Return("a", nil)
			flags.EXPECT().GetString(FlagReplicaWebURL).Return("b", nil)
			flags.EXPECT().GetString(FlagReplicaApiPath).Return("c", nil)
			flags.EXPECT().GetString(FlagReplicaUsername).Return("d", nil)
			flags.EXPECT().GetString(FlagReplicaPassword).Return("e", nil)
			flags.EXPECT().GetString(FlagReplicaCookie).Return("f", nil)
			flags.EXPECT().GetBool(FlagReplicaISV).Return(true, nil)
			flags.EXPECT().GetBool(FlagReplicaAutoSetup).Return(true, nil)
			flags.EXPECT().GetString(FlagReplicaInterfaceName).Return("g", nil)
			err := readFlags(cfg, flags)

			Ω(err).ShouldNot(HaveOccurred())
			Ω(cfg.Replica).Should(Equal(&types.AdGuardInstance{
				URL:                "a",
				WebURL:             "b",
				APIPath:            "c",
				Username:           "d",
				Password:           "e",
				Cookie:             "f",
				InsecureSkipVerify: true,
				AutoSetup:          true,
				InterfaceName:      "g",
			}))
		})
	})
})
