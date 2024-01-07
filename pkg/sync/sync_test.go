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

var _ = Describe("Sync", func() {
	var (
		mockCtrl *gm.Controller
		cl       *clientmock.MockClient
		w        *worker
		te       error
		ac       *actionContext
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
				Replicas: []types.AdGuardInstance{
					{},
				},
			},
		}
		te = errors.New(uuid.NewString())

		ac = &actionContext{
			continueOnError: false,
			rl:              l,
			origin: &origin{
				profileInfo: &model.ProfileInfo{
					Name:     "origin",
					Language: "en",
					Theme:    "auto",
				},
				status:         &model.ServerStatus{},
				safeSearch:     &model.SafeSearchConfig{},
				queryLogConfig: &model.QueryLogConfig{},
				statsConfig:    &model.StatsConfig{},
			},
			replicaStatus: &model.ServerStatus{},
			client:        cl,
			replica:       w.cfg.Replicas[0],
		}
	})
	AfterEach(func() {
		defer mockCtrl.Finish()
	})

	Context("worker", func() {
		Context("actionDNSRewrites", func() {
			var (
				domain string
				answer string
				reO    model.RewriteEntries
				reR    model.RewriteEntries
			)

			BeforeEach(func() {
				domain = uuid.NewString()
				answer = uuid.NewString()
				reO = model.RewriteEntries{{Domain: utils.Ptr(domain), Answer: utils.Ptr(answer)}}
				reR = model.RewriteEntries{{Domain: utils.Ptr(domain), Answer: utils.Ptr(answer)}}
			})
			It("should have no changes (empty slices)", func() {
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				err := actionDNSRewrites(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one rewrite entry", func() {
				reR = []model.RewriteEntry{}
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries(reO[0])
				cl.EXPECT().DeleteRewriteEntries()
				err := actionDNSRewrites(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []model.RewriteEntry{}
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := actionDNSRewrites(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should remove one rewrite entry", func() {
				reO = []model.RewriteEntry{}
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries(reR[0])
				err := actionDNSRewrites(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on RewriteList()", func() {
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(nil, te)
				err := actionDNSRewrites(ac)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on AddRewriteEntries()", func() {
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().AddRewriteEntries().Return(te)
				err := actionDNSRewrites(ac)
				Ω(err).Should(HaveOccurred())
			})
			It("should return error when error on DeleteRewriteEntries()", func() {
				ac.origin.rewrites = &reO
				cl.EXPECT().RewriteList().Return(&reR, nil)
				cl.EXPECT().DeleteRewriteEntries().Return(te)
				err := actionDNSRewrites(ac)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("actionClientSettings", func() {
			var (
				clR  *model.Clients
				name string
			)
			BeforeEach(func() {
				name = uuid.NewString()
				ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: utils.Ptr(name)}}}
				clR = &model.Clients{Clients: &model.ClientsArray{{Name: utils.Ptr(name)}}}
			})
			It("should have no changes (empty slices)", func() {
				cl.EXPECT().Clients().Return(clR, nil)
				err := actionClientSettings(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should add one client", func() {
				clR.Clients = &model.ClientsArray{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().AddClient(&(*ac.origin.clients.Clients)[0])
				err := actionClientSettings(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should update one client", func() {
				(*clR.Clients)[0].FilteringEnabled = utils.Ptr(true)
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().UpdateClient(&(*ac.origin.clients.Clients)[0])
				err := actionClientSettings(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should delete one client", func() {
				ac.origin.clients.Clients = &model.ClientsArray{}
				cl.EXPECT().Clients().Return(clR, nil)
				cl.EXPECT().DeleteClient(&(*clR.Clients)[0])
				err := actionClientSettings(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should return error when error on Clients()", func() {
				cl.EXPECT().Clients().Return(nil, te)
				err := actionClientSettings(ac)
				Ω(err).Should(HaveOccurred())
			})
		})
		Context("actionParental", func() {
			It("should have no changes", func() {
				cl.EXPECT().Parental()
				err := actionParental(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have parental enabled changes", func() {
				ac.origin.parental = true
				cl.EXPECT().Parental()
				cl.EXPECT().ToggleParental(true)
				err := actionParental(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionProtection", func() {
			It("should have no changes", func() {
				err := actionProtection(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have protection enabled changes", func() {
				ac.origin.status.ProtectionEnabled = true
				cl.EXPECT().ToggleProtection(true)
				err := actionProtection(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionSafeSearchConfig", func() {
			It("should have no changes", func() {
				cl.EXPECT().SafeSearchConfig().Return(ac.origin.safeSearch, nil)

				err := actionSafeSearchConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have safeSearch enabled changes", func() {
				ac.origin.safeSearch = &model.SafeSearchConfig{Enabled: utils.Ptr(true)}
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SetSafeSearchConfig(ac.origin.safeSearch)
				err := actionSafeSearchConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have Duckduckgo safeSearch enabled changed", func() {
				ac.origin.safeSearch = &model.SafeSearchConfig{Duckduckgo: utils.Ptr(true)}
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{Google: utils.Ptr(true)}, nil)
				cl.EXPECT().SetSafeSearchConfig(ac.origin.safeSearch)
				err := actionSafeSearchConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionProfileInfo", func() {
			It("should have no changes", func() {
				cl.EXPECT().ProfileInfo().Return(ac.origin.profileInfo, nil)
				err := actionProfileInfo(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have profileInfo language changed", func() {
				ac.origin.profileInfo.Language = "de"
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en"}, nil)
				cl.EXPECT().SetProfileInfo(&model.ProfileInfo{
					Language: "de",
					Name:     "replica",
					Theme:    "auto",
				})
				err := actionProfileInfo(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should not sync profileInfo if language is not set", func() {
				ac.origin.profileInfo.Language = ""
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
				cl.EXPECT().SetProfileInfo(ac.origin.profileInfo).Times(0)
				err := actionProfileInfo(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should not sync profileInfo if theme is not set", func() {
				ac.origin.profileInfo.Theme = ""
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
				cl.EXPECT().SetProfileInfo(ac.origin.profileInfo).Times(0)
				err := actionProfileInfo(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionSafeBrowsing", func() {
			It("should have no changes", func() {
				cl.EXPECT().SafeBrowsing()
				err := actionSafeBrowsing(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})

			It("should have safeBrowsing enabled changes", func() {
				ac.origin.safeBrowsing = true
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().ToggleSafeBrowsing(true)
				err := actionSafeBrowsing(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionQueryLogConfig", func() {
			var qlc *model.QueryLogConfig
			BeforeEach(func() {
				qlc = &model.QueryLogConfig{}
			})
			It("should have no changes", func() {
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				err := actionQueryLogConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have QueryLogConfig changes", func() {
				var interval model.QueryLogConfigInterval = 123
				ac.origin.queryLogConfig.Interval = &interval
				cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				cl.EXPECT().SetQueryLogConfig(&model.QueryLogConfig{AnonymizeClientIp: nil, Interval: &interval, Enabled: nil})
				err := actionQueryLogConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("syncConfigs", func() {
			var sc *model.StatsConfig
			BeforeEach(func() {
				sc = &model.StatsConfig{}
			})
			It("should have no changes", func() {
				cl.EXPECT().StatsConfig().Return(sc, nil)
				err := actionStatsConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have StatsConfig changes", func() {
				var interval model.StatsConfigInterval = 123
				ac.origin.statsConfig.Interval = &interval
				cl.EXPECT().StatsConfig().Return(sc, nil)
				cl.EXPECT().SetStatsConfig(&model.StatsConfig{Interval: &interval})
				err := actionStatsConfig(ac)
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
		Context("actionBlockedServices", func() {
			var rbs *model.BlockedServicesArray
			BeforeEach(func() {
				ac.origin.blockedServices = &model.BlockedServicesArray{"foo"}
				rbs = &model.BlockedServicesArray{"foo"}
			})
			It("should have no changes", func() {
				cl.EXPECT().BlockedServices().Return(rbs, nil)
				err := actionBlockedServices(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have blockedServices changes", func() {
				ac.origin.blockedServices = &model.BlockedServicesArray{"bar"}

				cl.EXPECT().BlockedServices().Return(rbs, nil)
				cl.EXPECT().SetBlockedServices(ac.origin.blockedServices)
				err := actionBlockedServices(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("actionBlockedServicesSchedule", func() {
			var rbss *model.BlockedServicesSchedule
			BeforeEach(func() {
				ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{}
				rbss = &model.BlockedServicesSchedule{}
			})
			It("should have no changes", func() {
				cl.EXPECT().BlockedServicesSchedule().Return(ac.origin.blockedServicesSchedule, nil)
				err := actionBlockedServicesSchedule(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have blockedServices schedule changes", func() {
				ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{Ids: utils.Ptr([]string{"bar"})}

				cl.EXPECT().BlockedServicesSchedule().Return(rbss, nil)
				cl.EXPECT().SetBlockedServicesSchedule(ac.origin.blockedServicesSchedule)
				err := actionBlockedServicesSchedule(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
		Context("syncFilters", func() {
			var rf *model.FilterStatus
			BeforeEach(func() {
				ac.origin.filters = &model.FilterStatus{}
				rf = &model.FilterStatus{}
			})
			It("should have no changes", func() {
				cl.EXPECT().Filtering().Return(rf, nil)
				err := actionFilters(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changes user roles", func() {
				ac.origin.filters.UserRules = utils.Ptr([]string{"foo"})
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().SetCustomRules(ac.origin.filters.UserRules)
				err := actionFilters(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changed filtering config", func() {
				ac.origin.filters.Enabled = utils.Ptr(true)
				ac.origin.filters.Interval = utils.Ptr(123)
				cl.EXPECT().Filtering().Return(rf, nil)
				cl.EXPECT().ToggleFiltering(*ac.origin.filters.Enabled, *ac.origin.filters.Interval)
				err := actionFilters(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("actionDNSAccessLists", func() {
			var ral *model.AccessList
			BeforeEach(func() {
				ac.origin.accessList = &model.AccessList{}
				ral = &model.AccessList{}
			})
			It("should have no changes", func() {
				cl.EXPECT().AccessList().Return(ral, nil)
				err := actionDNSAccessLists(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have access list changes", func() {
				ral.BlockedHosts = utils.Ptr([]string{"foo"})
				cl.EXPECT().AccessList().Return(ral, nil)
				cl.EXPECT().SetAccessList(ac.origin.accessList)
				err := actionDNSAccessLists(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("actionDNSServerConfig", func() {
			var rdc *model.DNSConfig
			BeforeEach(func() {
				ac.origin.dnsConfig = &model.DNSConfig{}
				rdc = &model.DNSConfig{}
			})
			It("should have no changes", func() {
				cl.EXPECT().DNSConfig().Return(rdc, nil)
				err := actionDNSServerConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have dns config changes", func() {
				rdc.BootstrapDns = utils.Ptr([]string{"foo"})
				cl.EXPECT().DNSConfig().Return(rdc, nil)
				cl.EXPECT().SetDNSConfig(ac.origin.dnsConfig)
				err := actionDNSServerConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
		})

		Context("actionDHCPServerConfig", func() {
			var rsc *model.DhcpStatus
			BeforeEach(func() {
				ac.origin.dhcpServerConfig = &model.DhcpStatus{V4: &model.DhcpConfigV4{
					GatewayIp:  utils.Ptr("1.2.3.4"),
					RangeStart: utils.Ptr("1.2.3.5"),
					RangeEnd:   utils.Ptr("1.2.3.6"),
					SubnetMask: utils.Ptr("255.255.255.0"),
				}}
				rsc = &model.DhcpStatus{}
				w.cfg.Features.DHCP.StaticLeases = false
			})
			It("should have no changes", func() {
				rsc.V4 = ac.origin.dhcpServerConfig.V4
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				err := actionDHCPServerConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should have changes", func() {
				rsc.Enabled = utils.Ptr(true)
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				cl.EXPECT().SetDhcpConfig(ac.origin.dhcpServerConfig)
				err := actionDHCPServerConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should use replica interface name", func() {
				ac.replica.InterfaceName = "foo"
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := ac.origin.dhcpServerConfig.Clone()
				oscClone.InterfaceName = utils.Ptr("foo")
				cl.EXPECT().SetDhcpConfig(oscClone)
				err := actionDHCPServerConfig(ac)
				Ω(err).ShouldNot(HaveOccurred())
			})
			It("should enable the target dhcp server", func() {
				ac.replica.DHCPServerEnabled = utils.Ptr(true)
				cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := ac.origin.dhcpServerConfig.Clone()
				oscClone.Enabled = utils.Ptr(true)
				cl.EXPECT().SetDhcpConfig(oscClone)
				err := actionDHCPServerConfig(ac)
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
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().BlockedServices()
				cl.EXPECT().BlockedServicesSchedule()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&model.StatsConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&model.StatsConfig{}, nil)
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().BlockedServices()
				cl.EXPECT().BlockedServicesSchedule()
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
				w.sync()
			})
			It("should not sync DHCP", func() {
				w.cfg.Features.DHCP.ServerConfig = false
				w.cfg.Features.DHCP.StaticLeases = false
				// origin
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().BlockedServices()
				cl.EXPECT().BlockedServicesSchedule()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&model.StatsConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&model.StatsConfig{}, nil)
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().AddRewriteEntries()
				cl.EXPECT().DeleteRewriteEntries()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().BlockedServices()
				cl.EXPECT().BlockedServicesSchedule()
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
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
				cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				cl.EXPECT().Parental()
				cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				cl.EXPECT().SafeBrowsing()
				cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				cl.EXPECT().BlockedServices()
				cl.EXPECT().BlockedServicesSchedule()
				cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfig{}, nil)
				cl.EXPECT().StatsConfig().Return(&model.StatsConfig{}, nil)
				cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)

				// replica
				cl.EXPECT().Host()
				cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
				w.sync()
			})
		})
	})
})
