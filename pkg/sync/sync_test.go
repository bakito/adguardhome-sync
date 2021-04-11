package sync

import (
	"errors"
	"github.com/bakito/adguardhome-sync/pkg/client"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mc "github.com/bakito/adguardhome-sync/pkg/mocks/client"
	"github.com/bakito/adguardhome-sync/pkg/types"
	gm "github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

var _ = Describe("Sync", func() {
	var (
		mockCtrl *gm.Controller
		cl       *mc.MockClient
		w        *worker
		te       error
	)

	BeforeEach(func() {
		mockCtrl = gm.NewController(GinkgoT())
		cl = mc.NewMockClient(mockCtrl)
		w = &worker{
			createClient: func(instance types.AdGuardInstance) (client.Client, error) {
				return cl, nil
			},
		}
		te = errors.New(uuid.NewString())
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})

	Context("worker", func() {
		Context("syncRewrites", func() {
			var (
				domain string
				answer string
				reO    types.RewriteEntries
				reR    types.RewriteEntries
			)

			BeforeEach(func() {
				domain = uuid.NewString()
				answer = uuid.NewString()
				reO = []types.RewriteEntry{{Domain: domain, Answer: answer}}
				reR = []types.RewriteEntry{{Domain: domain, Answer: answer}}
			})
			It("should have no changes (empty slices)", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(&reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one rewrite entry", func() {
				reR = []types.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries(reO[0])
				cl.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(&reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []types.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := w.syncRewrites(&reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []types.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := w.syncRewrites(&reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on RewriteList()", func() {
				cl.EXPECT().RewriteList().Return(nil, te)
				err := w.syncRewrites(&reO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on AddRewriteEntries()", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries().Return(te)
				err := w.syncRewrites(&reO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteRewriteEntries()", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries().Return(te)
				err := w.syncRewrites(&reO, cl)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("syncClients", func() {
			var (
				clO  *types.Clients
				clR  *types.Clients
				name string
			)
			BeforeEach(func() {
				name = uuid.NewString()
				clO = &types.Clients{Clients: []types.Client{{Name: name}}}
				clR = &types.Clients{Clients: []types.Client{{Name: name}}}
			})
			It("should have no changes (empty slices)", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one client", func() {
				clR.Clients = []types.Client{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients(clO.Clients[0])
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should update one client", func() {
				clR.Clients[0].Disallowed = true
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients(clO.Clients[0])
				cl.EXPECT().DeleteClients()
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should delete one client", func() {
				clO.Clients = []types.Client{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients(clR.Clients[0])
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on Clients()", func() {
				cl.EXPECT().Clients().Return(nil, te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on AddClients()", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on UpdateClients()", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteClients()", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("syncGeneralSettings", func() {
			var (
				o  *origin
				rs *types.Status
			)
			BeforeEach(func() {
				o = &origin{
					status: &types.Status{},
				}
				rs = &types.Status{}
			})
			It("should have no changes", func() {
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				err := w.syncGeneralSettings(o, rs, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have protection enabled changes", func() {
				o.status.ProtectionEnabled = true
				cl.EXPECT().ToggleProtection(true)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				err := w.syncGeneralSettings(o, rs, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have parental enabled changes", func() {
				o.parental = true
				cl.EXPECT().Parental()
				cl.EXPECT().ToggleParental(true)
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				err := w.syncGeneralSettings(o, rs, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have safeSearch enabled changes", func() {
				o.safeSearch = true
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().ToggleSafeSearch(true)
				cl.EXPECT().SafeBrowsing()
				err := w.syncGeneralSettings(o, rs, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have safeBrowsing enabled changes", func() {
				o.safeBrowsing = true
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().ToggleSafeBrowsing(true)
				err := w.syncGeneralSettings(o, rs, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("syncConfigs", func() {
			var (
				o   *origin
				qlc *types.QueryLogConfig
				sc  *types.IntervalConfig
			)
			BeforeEach(func() {
				o = &origin{
					queryLogConfig: &types.QueryLogConfig{},
					statsConfig:    &types.IntervalConfig{},
				}
				qlc = &types.QueryLogConfig{}
				sc = &types.IntervalConfig{}
			})
			It("should have no changes", func() {
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				cl.EXPECT().StatsConfig().Return(sc, nil)
				err := w.syncConfigs(o, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have QueryLogConfig changes", func() {
				o.queryLogConfig.Interval = 123
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				cl.EXPECT().SetQueryLogConfig(false, 123, false)
				cl.EXPECT().StatsConfig().Return(sc, nil)
				err := w.syncConfigs(o, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have StatsConfig changes", func() {
				o.statsConfig.Interval = 123
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				cl.EXPECT().StatsConfig().Return(sc, nil)
				cl.EXPECT().SetStatsConfig(123)
				err := w.syncConfigs(o, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("syncServices", func() {
			var (
				os types.Services
				rs types.Services
			)
			BeforeEach(func() {
				os = []string{"foo"}
				rs = []string{"foo"}
			})
			It("should have no changes", func() {
				cl.EXPECT().Services().Return(rs, nil)
				err := w.syncServices(os, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have services changes", func() {
				os = []string{"bar"}
				cl.EXPECT().Services().Return(rs, nil)
				cl.EXPECT().SetServices(os)
				err := w.syncServices(os, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("syncFilters", func() {
			var (
				of *types.FilteringStatus
				rf *types.FilteringStatus
			)
			BeforeEach(func() {
				of = &types.FilteringStatus{}
				rf = &types.FilteringStatus{}
			})
			It("should have no changes", func() {
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				err := w.syncFilters(of, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changes user roles", func() {
				of.UserRules = []string{"foo"}
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().SetCustomRules(of.UserRules)
				err := w.syncFilters(of, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changed filtering config", func() {
				of.Enabled = true
				of.Interval = 123
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().ToggleFiltering(of.Enabled, of.Interval)
				err := w.syncFilters(of, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("sync", func() {

			It("should have no changes", func() {
				w.cfg = &types.Config{
					Origin:  types.AdGuardInstance{},
					Replica: types.AdGuardInstance{URL: "foo"},
				}
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&types.Status{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&types.RewriteEntries{}, nil)
				cl.EXPECT().Services()
				cl.EXPECT().Filtering().Return(&types.FilteringStatus{}, nil)
				cl.EXPECT().Clients().Return(&types.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&types.Status{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().RewriteList().Return(&types.RewriteEntries{}, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().Filtering().Return(&types.FilteringStatus{}, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().Services()
				cl.EXPECT().Clients().Return(&types.Clients{}, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				w.sync()
			})
		})
	})
})
