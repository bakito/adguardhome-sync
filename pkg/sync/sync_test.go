package sync

import (
	"errors"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	clientmock "github.com/bakito/adguardhome-sync/pkg/mocks/client"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/bakito/adguardhome-sync/pkg/versions"
	gm "github.com/golang/mock/gomock"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var boolTrue = true

var _ = Describe("Sync", func() {
	var (
		mockCtrl *gm.Controller
		cl       *clientmock.MockClient
		w        *worker
		te       error
	)

	BeforeEach(func() {
		mockCtrl = gm.NewController(GinkgoT())
		cl = clientmock.NewMockClient(mockCtrl)
		w = &worker{
			createClient: func(instance types.AdGuardInstance) (client.Client, error) {
				return cl, nil
			},
			cfg: &types.Config{
				Features: types.Features{
					DHCP: types.DHCP{
						ServerConfig: true,
						StaticLeases: true,
					},
					DNS: types.DNS{
						ServerConfig: true,
						Rewrites:     true,
						AccessLists:  true,
					},
					Filters:         true,
					ClientSettings:  true,
					Services:        true,
					GeneralSettings: true,
					StatsConfig:     true,
					QueryLogConfig:  true,
				},
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
				reO    model.RewriteEntries
				reR    model.RewriteEntries
			)

			BeforeEach(func() {
				domain = uuid.NewString()
				answer = uuid.NewString()
				reO = []model.RewriteEntry{{Domain: utils.Ptr(domain), Answer: utils.Ptr(answer)}}
				reR = []model.RewriteEntry{{Domain: utils.Ptr(domain), Answer: utils.Ptr(answer)}}
			})
			It("should have no changes (empty slices)", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one rewrite entry", func() {
				reR = []model.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries(reO[0])
				cl.EXPECT().DeleteRewriteEntries()
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []model.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []model.RewriteEntry{}
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on RewriteList()", func() {
				cl.EXPECT().RewriteList().Return(nil, te)
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on AddRewriteEntries()", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().AddRewriteEntries().Return(te)
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteRewriteEntries()", func() {
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().DeleteRewriteEntries().Return(te)
				err := w.syncRewrites(l, &reO, cl)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("syncClients", func() {
			var (
				clO  *model.Clients
				clR  *model.Clients
				name string
			)
			BeforeEach(func() {
				name = uuid.NewString()
				clO = &model.Clients{Clients: &model.ClientsArray{{Name: utils.Ptr(name)}}}
				clR = &model.Clients{Clients: &model.ClientsArray{{Name: utils.Ptr(name)}}}
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
				clR.Clients = &model.ClientsArray{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients(&(*clO.Clients)[0])
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should update one client", func() {
				(*clR.Clients)[0].FilteringEnabled = utils.Ptr(true)
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients(&(*clO.Clients)[0])
				cl.EXPECT().DeleteClients()
				err := w.syncClients(clO, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should delete one client", func() {
				clO.Clients = &model.ClientsArray{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients(&(*clR.Clients)[0])
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
				cl.EXPECT().DeleteClients()
				cl.EXPECT().AddClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on UpdateClients()", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().DeleteClients()
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteClients()", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().DeleteClients().Return(te)
				err := w.syncClients(clO, cl)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("syncGeneralSettings", func() {
			var (
				o  *origin
				rs *model.ServerStatus
			)
			BeforeEach(func() {
				o = &origin{
					status: &model.ServerStatus{},
				}
				rs = &model.ServerStatus{}
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
				cl.EXPECT().SetQueryLogConfig(false, 123.0, false)
				cl.EXPECT().StatsConfig().Return(sc, nil)
				err := w.syncConfigs(o, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have StatsConfig changes", func() {
				o.statsConfig.Interval = 123
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				cl.EXPECT().StatsConfig().Return(sc, nil)
				cl.EXPECT().SetStatsConfig(123.0)
				err := w.syncConfigs(o, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("statusWithSetup", func() {
			var (
				status *model.ServerStatus
				inst   types.AdGuardInstance
			)
			BeforeEach(func() {
				status = &model.ServerStatus{}
				inst = types.AdGuardInstance{
					AutoSetup: true,
				}
			})
			It("should get the replica status", func() {
				cl.EXPECT().Status().Return(status, nil)
				st, err := w.statusWithSetup(l, inst, cl)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(st).Should(Equal(status))
			})
			It("should runs setup before getting replica status", func() {
				cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
				cl.EXPECT().Setup()
				cl.EXPECT().Status().Return(status, nil)
				st, err := w.statusWithSetup(l, inst, cl)
				Ω(err).ShouldNot(HaveOccurred())
				Ω(st).Should(Equal(status))
			})
			It("should fail on setup", func() {
				cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
				cl.EXPECT().Setup().Return(te)
				st, err := w.statusWithSetup(l, inst, cl)
				Ω(err).Should(HaveOccurred())
				Ω(st).Should(BeNil())
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
				of *model.FilterStatus
				rf *model.FilterStatus
			)
			BeforeEach(func() {
				of = &model.FilterStatus{}
				rf = &model.FilterStatus{}
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
				of.UserRules = utils.Ptr([]string{"foo"})
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
				of.Enabled = utils.Ptr(true)
				of.Interval = utils.Ptr(123)
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().ToggleFiltering(*of.Enabled, *of.Interval)
				err := w.syncFilters(of, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("syncDNS", func() {
			var (
				oal *model.AccessList
				ral *model.AccessList
				odc *model.DNSConfig
				rdc *model.DNSConfig
			)
			BeforeEach(func() {
				oal = &model.AccessList{}
				ral = &model.AccessList{}
				odc = &model.DNSConfig{}
				rdc = &model.DNSConfig{}
			})
			It("should have no changes", func() {
				cl.EXPECT().AccessList().Return(ral, nil)
				cl.EXPECT().DNSConfig().Return(rdc, nil)
				err := w.syncDNS(oal, odc, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have access list changes", func() {
				ral.BlockedHosts = utils.Ptr([]string{"foo"})
				cl.EXPECT().AccessList().Return(ral, nil)
				cl.EXPECT().DNSConfig().Return(rdc, nil)
				cl.EXPECT().SetAccessList(oal)
				err := w.syncDNS(oal, odc, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have dns config changes", func() {
				rdc.BootstrapDns = utils.Ptr([]string{"foo"})
				cl.EXPECT().AccessList().Return(ral, nil)
				cl.EXPECT().DNSConfig().Return(rdc, nil)
				cl.EXPECT().SetDNSConfig(odc)
				err := w.syncDNS(oal, odc, cl)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("syncDHCPServer", func() {
			var (
				osc *model.DhcpStatus
				rsc *model.DhcpStatus
			)
			BeforeEach(func() {
				osc = &model.DhcpStatus{V4: &model.DhcpConfigV4{
					GatewayIp:  utils.Ptr("1.2.3.4"),
					RangeStart: utils.Ptr("1.2.3.5"),
					RangeEnd:   utils.Ptr("1.2.3.6"),
					SubnetMask: utils.Ptr("255.255.255.0"),
				}}
				rsc = &model.DhcpStatus{}
				w.cfg.Features.DHCP.StaticLeases = false
			})
			It("should have no changes", func() {
				rsc.V4 = osc.V4
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				err := w.syncDHCPServer(osc, cl, types.AdGuardInstance{})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changes", func() {
				rsc.Enabled = utils.Ptr(true)
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				cl.EXPECT().SetDhcpConfig(osc)
				err := w.syncDHCPServer(osc, cl, types.AdGuardInstance{})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should use replica interface name", func() {
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := osc.Clone()
				oscClone.InterfaceName = utils.Ptr("foo")
				cl.EXPECT().SetDhcpConfig(oscClone)
				err := w.syncDHCPServer(osc, cl, types.AdGuardInstance{InterfaceName: "foo"})
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should enable the target dhcp server", func() {
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := osc.Clone()
				oscClone.Enabled = utils.Ptr(true)
				cl.EXPECT().SetDhcpConfig(oscClone)
				err := w.syncDHCPServer(osc, cl, types.AdGuardInstance{DHCPServerEnabled: &boolTrue})
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("sync", func() {
			BeforeEach(func() {
				w.cfg = &types.Config{
					Origin:  types.AdGuardInstance{},
					Replica: &types.AdGuardInstance{URL: "foo"},
					Features: types.Features{
						DHCP: types.DHCP{
							ServerConfig: true,
							StaticLeases: true,
						},
						DNS: types.DNS{
							ServerConfig: true,
							Rewrites:     true,
							AccessLists:  true,
						},
						Filters:         true,
						ClientSettings:  true,
						Services:        true,
						GeneralSettings: true,
						StatsConfig:     true,
						QueryLogConfig:  true,
					},
				}
			})
			It("should have no changes", func() {
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().Services()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().Services()
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
				cl.EXPECT().AddDHCPStaticLeases().Return(nil)
				cl.EXPECT().DeleteDHCPStaticLeases().Return(nil)
				w.sync()
			})
			It("should not sync DHCP", func() {
				w.cfg.Features.DHCP.ServerConfig = false
				w.cfg.Features.DHCP.StaticLeases = false
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().Services()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().AddFilters(false)
				cl.EXPECT().UpdateFilters(false)
				cl.EXPECT().DeleteFilters(false)
				cl.EXPECT().AddFilters(true)
				cl.EXPECT().UpdateFilters(true)
				cl.EXPECT().DeleteFilters(true)
				cl.EXPECT().Services()
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().AddClients()
				cl.EXPECT().UpdateClients()
				cl.EXPECT().DeleteClients()
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				w.sync()
			})
			It("origin version is too small", func() {
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
				w.sync()
			})
			It("replica version is too small", func() {
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().Services()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
				w.sync()
			})
			It("replica version is with incompatible API", func() {
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearch()
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().Services()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&types.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&types.IntervalConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.IncompatibleAPI}, nil)
				w.sync()
			})
		})
	})
})
