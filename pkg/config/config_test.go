package config_test

import (
	"os"

	"github.com/bakito/adguardhome-sync/pkg/config"
	flagsmock "github.com/bakito/adguardhome-sync/pkg/mocks/flags"
	gm "github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Context("Get", func() {
		var (
			flags    *flagsmock.MockFlags
			mockCtrl *gm.Controller
		)
		BeforeEach(func() {
			mockCtrl = gm.NewController(GinkgoT())
			flags = flagsmock.NewMockFlags(mockCtrl)
		})
		AfterEach(func() {
			defer mockCtrl.Finish()
		})
		Context("Get", func() {
			It("should have the origin URL from the config file", func() {
				flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

				cfg, err := config.Get("../../testdata/config_test.yaml", flags)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Origin.URL).Should(Equal("https://origin-file:443"))
			})
			It("should have the origin URL from the config flags", func() {
				flags.EXPECT().Changed(config.FlagOriginURL).Return(true).AnyTimes()
				flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()
				flags.EXPECT().GetString(config.FlagOriginURL).Return("https://origin-flag:443", nil).AnyTimes()

				cfg, err := config.Get("../../testdata/config_test.yaml", flags)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Origin.URL).Should(Equal("https://origin-flag:443"))
			})
			It("should have the origin URL from the config env var", func() {
				os.Setenv("ORIGIN_URL", "https://origin-env:443")
				defer func() {
					_ = os.Unsetenv("ORIGIN_URL")
				}()
				flags.EXPECT().Changed(gm.Any()).Return(false).AnyTimes()

				cfg, err := config.Get("../../testdata/config_test.yaml", flags)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(cfg.Origin.URL).Should(Equal("https://origin-env:443"))
			})
		})
	})
})
