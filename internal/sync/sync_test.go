package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/client/model"
	clientmock "github.com/bakito/adguardhome-sync/internal/mocks/client"
	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/adguardhome-sync/internal/utils"
	"github.com/bakito/adguardhome-sync/internal/versions"
)

type testEnv struct {
	mockCtrl *gm.Controller
	cl       *clientmock.MockClient
	w        *worker
	ac       *actionContext
	te       error
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	mockCtrl := gm.NewController(t)
	cl := clientmock.NewMockClient(mockCtrl)
	te := errors.New(uuid.NewString())
	w := &worker{
		createClient: func(_ types.AdGuardInstance, _ time.Duration) (client.Client, error) {
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
				Theme:           true,
			},
			Replicas: []types.AdGuardInstance{
				{},
			},
		},
	}

	ac := &actionContext{
		cfg: w.cfg,
		rl:  l,
		origin: &origin{
			profileInfo: &model.ProfileInfo{
				Name:     "origin",
				Language: "en",
				Theme:    "auto",
			},
			status:         &model.ServerStatus{},
			safeSearch:     &model.SafeSearchConfig{},
			queryLogConfig: &model.QueryLogConfigWithIgnored{},
			statsConfig:    &model.PutStatsConfigUpdateRequest{},
		},
		replicaStatus: &model.ServerStatus{},
		client:        cl,
		replica:       w.cfg.Replicas[0],
	}
	return &testEnv{
		mockCtrl: mockCtrl,
		cl:       cl,
		w:        w,
		ac:       ac,
		te:       te,
	}
}

func TestSync(t *testing.T) {
	t.Run("worker", func(t *testing.T) {
		t.Run("actionDNSRewrites", func(t *testing.T) {
			domain := uuid.NewString()
			answer := uuid.NewString()
			reO := model.RewriteEntries{{Domain: &domain, Answer: &answer}}
			reR := model.RewriteEntries{{Domain: &domain, Answer: &answer}}

			t.Run("should have no changes (empty slices)", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.rewrites = &reO
				env.cl.EXPECT().RewriteList().Return(&reR, nil)
				env.cl.EXPECT().AddRewriteEntries()
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().UpdateRewriteEntries()
				err := actionDNSRewrites(env.ac)
				if err != nil {
					t.Errorf("actionDNSRewrites() error = %v, want nil", err)
				}
			})
			t.Run("should add one rewrite entry", func(t *testing.T) {
				env := newTestEnv(t)
				reRLocal := model.RewriteEntries{}
				env.ac.origin.rewrites = &reO
				env.cl.EXPECT().RewriteList().Return(&reRLocal, nil)
				env.cl.EXPECT().AddRewriteEntries(reO[0])
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().UpdateRewriteEntries()
				err := actionDNSRewrites(env.ac)
				if err != nil {
					t.Errorf("actionDNSRewrites() error = %v, want nil", err)
				}
			})
			t.Run("should remove one rewrite entry", func(t *testing.T) {
				env := newTestEnv(t)
				reOLocal := model.RewriteEntries{}
				env.ac.origin.rewrites = &reOLocal
				env.cl.EXPECT().RewriteList().Return(&reR, nil)
				env.cl.EXPECT().AddRewriteEntries()
				env.cl.EXPECT().DeleteRewriteEntries(reR[0])
				env.cl.EXPECT().UpdateRewriteEntries()
				err := actionDNSRewrites(env.ac)
				if err != nil {
					t.Errorf("actionDNSRewrites() error = %v, want nil", err)
				}
			})
			t.Run("should update one rewrite entry", func(t *testing.T) {
				env := newTestEnv(t)
				reOLocal := model.RewriteEntries{{Domain: &domain, Answer: &answer, Enabled: utils.Ptr(false)}}
				reRLocal := model.RewriteEntries{{Domain: &domain, Answer: &answer, Enabled: utils.Ptr(true)}}
				env.ac.origin.rewrites = &reOLocal
				env.cl.EXPECT().RewriteList().Return(&reRLocal, nil)
				env.cl.EXPECT().AddRewriteEntries()
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().UpdateRewriteEntries(gm.Any())
				err := actionDNSRewrites(env.ac)
				if err != nil {
					t.Errorf("actionDNSRewrites() error = %v, want nil", err)
				}
			})
			t.Run("should return error when error on RewriteList()", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.rewrites = &reO
				env.cl.EXPECT().RewriteList().Return(nil, env.te)
				err := actionDNSRewrites(env.ac)
				if err == nil {
					t.Error("actionDNSRewrites() error = nil, want error")
				}
			})
			t.Run("should return error when error on AddRewriteEntries()", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.rewrites = &reO
				env.cl.EXPECT().RewriteList().Return(&reR, nil)
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().AddRewriteEntries().Return(env.te)
				err := actionDNSRewrites(env.ac)
				if err == nil {
					t.Error("actionDNSRewrites() error = nil, want error")
				}
			})
			t.Run("should return error when error on DeleteRewriteEntries()", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.rewrites = &reO
				env.cl.EXPECT().RewriteList().Return(&reR, nil)
				env.cl.EXPECT().DeleteRewriteEntries().Return(env.te)
				err := actionDNSRewrites(env.ac)
				if err == nil {
					t.Error("actionDNSRewrites() error = nil, want error")
				}
			})
		})

		t.Run("actionClientSettings", func(t *testing.T) {
			name := uuid.NewString()

			t.Run("should have no changes (empty slices)", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: &name}}}
				clR := &model.Clients{Clients: &model.ClientsArray{{Name: &name}}}
				env.cl.EXPECT().Clients().Return(clR, nil)
				err := actionClientSettings(env.ac)
				if err != nil {
					t.Errorf("actionClientSettings() error = %v, want nil", err)
				}
			})
			t.Run("should add one client", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: &name}}}
				clRLocal := &model.Clients{Clients: &model.ClientsArray{}}
				env.cl.EXPECT().Clients().Return(clRLocal, nil)
				env.cl.EXPECT().AddClient(&(*env.ac.origin.clients.Clients)[0])
				err := actionClientSettings(env.ac)
				if err != nil {
					t.Errorf("actionClientSettings() error = %v, want nil", err)
				}
			})
			t.Run("should update one client", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: &name}}}
				clRLocal := &model.Clients{Clients: &model.ClientsArray{{Name: &name, FilteringEnabled: utils.Ptr(true)}}}
				env.cl.EXPECT().Clients().Return(clRLocal, nil)
				env.cl.EXPECT().UpdateClient(&(*env.ac.origin.clients.Clients)[0])
				err := actionClientSettings(env.ac)
				if err != nil {
					t.Errorf("actionClientSettings() error = %v, want nil", err)
				}
			})
			t.Run("should delete one client", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{}}
				clR := &model.Clients{Clients: &model.ClientsArray{{Name: &name}}}
				env.cl.EXPECT().Clients().Return(clR, nil)
				env.cl.EXPECT().DeleteClient(&(*clR.Clients)[0])
				err := actionClientSettings(env.ac)
				if err != nil {
					t.Errorf("actionClientSettings() error = %v, want nil", err)
				}
			})
			t.Run("should return error when error on Clients()", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().Clients().Return(nil, env.te)
				err := actionClientSettings(env.ac)
				if err == nil {
					t.Error("actionClientSettings() error = nil, want error")
				}
			})
		})

		t.Run("actionParental", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().Parental()
				err := actionParental(env.ac)
				if err != nil {
					t.Errorf("actionParental() error = %v, want nil", err)
				}
			})
			t.Run("should have parental enabled changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.parental = true
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().ToggleParental(true)
				err := actionParental(env.ac)
				if err != nil {
					t.Errorf("actionParental() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionProtection", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				err := actionProtection(env.ac)
				if err != nil {
					t.Errorf("actionProtection() error = %v, want nil", err)
				}
			})
			t.Run("should have protection enabled changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.status.ProtectionEnabled = true
				env.cl.EXPECT().ToggleProtection(true)
				err := actionProtection(env.ac)
				if err != nil {
					t.Errorf("actionProtection() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionSafeSearchConfig", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().SafeSearchConfig().Return(env.ac.origin.safeSearch, nil)

				err := actionSafeSearchConfig(env.ac)
				if err != nil {
					t.Errorf("actionSafeSearchConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have safeSearch enabled changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.safeSearch = &model.SafeSearchConfig{Enabled: utils.Ptr(true)}
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SetSafeSearchConfig(env.ac.origin.safeSearch)
				err := actionSafeSearchConfig(env.ac)
				if err != nil {
					t.Errorf("actionSafeSearchConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have Duckduckgo safeSearch enabled changed", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.safeSearch = &model.SafeSearchConfig{Duckduckgo: utils.Ptr(true)}
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{Google: utils.Ptr(true)}, nil)
				env.cl.EXPECT().SetSafeSearchConfig(env.ac.origin.safeSearch)
				err := actionSafeSearchConfig(env.ac)
				if err != nil {
					t.Errorf("actionSafeSearchConfig() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionProfileInfo", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().ProfileInfo().Return(env.ac.origin.profileInfo, nil)
				err := actionProfileInfo(env.ac)
				if err != nil {
					t.Errorf("actionProfileInfo() error = %v, want nil", err)
				}
			})
			t.Run("should have profileInfo language changed", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.profileInfo.Language = "de"
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en"}, nil)
				env.cl.EXPECT().SetProfileInfo(&model.ProfileInfo{
					Language: "de",
					Name:     "replica",
					Theme:    "auto",
				})
				err := actionProfileInfo(env.ac)
				if err != nil {
					t.Errorf("actionProfileInfo() error = %v, want nil", err)
				}
			})
			t.Run("should not change theme if feature is disabled", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.profileInfo.Language = "de"
				env.ac.cfg.Features.Theme = false
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en"}, nil)
				env.cl.EXPECT().SetProfileInfo(&model.ProfileInfo{
					Language: "de",
					Name:     "replica",
					Theme:    "",
				})
				err := actionProfileInfo(env.ac)
				if err != nil {
					t.Errorf("actionProfileInfo() error = %v, want nil", err)
				}
			})
			t.Run("should not sync profileInfo if language is not set", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.profileInfo.Language = ""
				env.cl.EXPECT().
					ProfileInfo().
					Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
				env.cl.EXPECT().SetProfileInfo(env.ac.origin.profileInfo).Times(0)
				err := actionProfileInfo(env.ac)
				if err != nil {
					t.Errorf("actionProfileInfo() error = %v, want nil", err)
				}
			})
			t.Run("should not sync profileInfo if theme is not set", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.profileInfo.Theme = ""
				env.cl.EXPECT().
					ProfileInfo().
					Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
				env.cl.EXPECT().SetProfileInfo(env.ac.origin.profileInfo).Times(0)
				err := actionProfileInfo(env.ac)
				if err != nil {
					t.Errorf("actionProfileInfo() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionSafeBrowsing", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().SafeBrowsing()
				err := actionSafeBrowsing(env.ac)
				if err != nil {
					t.Errorf("actionSafeBrowsing() error = %v, want nil", err)
				}
			})

			t.Run("should have safeBrowsing enabled changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.safeBrowsing = true
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().ToggleSafeBrowsing(true)
				err := actionSafeBrowsing(env.ac)
				if err != nil {
					t.Errorf("actionSafeBrowsing() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionQueryLogConfig", func(t *testing.T) {
			qlc := &model.QueryLogConfigWithIgnored{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				err := actionQueryLogConfig(env.ac)
				if err != nil {
					t.Errorf("actionQueryLogConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have QueryLogConfig changes", func(t *testing.T) {
				env := newTestEnv(t)
				var interval model.QueryLogConfigInterval = 123
				env.ac.origin.queryLogConfig.Interval = &interval
				env.cl.EXPECT().QueryLogConfig().Return(qlc, nil)
				env.cl.EXPECT().
					SetQueryLogConfig(&model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{AnonymizeClientIp: nil, Interval: &interval, Enabled: nil}})
				err := actionQueryLogConfig(env.ac)
				if err != nil {
					t.Errorf("actionQueryLogConfig() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionStatsConfig", func(t *testing.T) {
			sc := &model.PutStatsConfigUpdateRequest{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().StatsConfig().Return(sc, nil)
				err := actionStatsConfig(env.ac)
				if err != nil {
					t.Errorf("actionStatsConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have StatsConfig changes", func(t *testing.T) {
				env := newTestEnv(t)
				var interval float32 = 123
				env.ac.origin.statsConfig.Interval = interval
				env.cl.EXPECT().StatsConfig().Return(sc, nil)
				env.cl.EXPECT().SetStatsConfig(&model.PutStatsConfigUpdateRequest{Interval: interval})
				err := actionStatsConfig(env.ac)
				if err != nil {
					t.Errorf("actionStatsConfig() error = %v, want nil", err)
				}
			})
		})

		t.Run("statusWithSetup", func(t *testing.T) {
			status := &model.ServerStatus{}
			inst := types.AdGuardInstance{
				AutoSetup: true,
			}
			t.Run("should get the replica status", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().Status().Return(status, nil)
				st, err := env.w.statusWithSetup(l, inst, env.cl)
				if err != nil {
					t.Errorf("statusWithSetup() error = %v, want nil", err)
				}
				if diff := cmp.Diff(status, st); diff != "" {
					t.Errorf("statusWithSetup() mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("should runs setup before getting replica status", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
				env.cl.EXPECT().Setup()
				env.cl.EXPECT().Status().Return(status, nil)
				st, err := env.w.statusWithSetup(l, inst, env.cl)
				if err != nil {
					t.Errorf("statusWithSetup() error = %v, want nil", err)
				}
				if diff := cmp.Diff(status, st); diff != "" {
					t.Errorf("statusWithSetup() mismatch (-want +got):\n%s", diff)
				}
			})
			t.Run("should fail on setup", func(t *testing.T) {
				env := newTestEnv(t)
				env.cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
				env.cl.EXPECT().Setup().Return(env.te)
				st, err := env.w.statusWithSetup(l, inst, env.cl)
				if err == nil {
					t.Error("statusWithSetup() error = nil, want error")
				}
				if st != nil {
					t.Errorf("statusWithSetup() st = %v, want nil", st)
				}
			})
		})

		t.Run("actionBlockedServicesSchedule", func(t *testing.T) {
			rbss := &model.BlockedServicesSchedule{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{}
				env.cl.EXPECT().BlockedServicesSchedule().Return(env.ac.origin.blockedServicesSchedule, nil)
				err := actionBlockedServicesSchedule(env.ac)
				if err != nil {
					t.Errorf("actionBlockedServicesSchedule() error = %v, want nil", err)
				}
			})
			t.Run("should have blockedServices schedule changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{Ids: utils.Ptr([]string{"bar"})}

				env.cl.EXPECT().BlockedServicesSchedule().Return(rbss, nil)
				env.cl.EXPECT().SetBlockedServicesSchedule(env.ac.origin.blockedServicesSchedule)
				err := actionBlockedServicesSchedule(env.ac)
				if err != nil {
					t.Errorf("actionBlockedServicesSchedule() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionFilters", func(t *testing.T) {
			rf := &model.FilterStatus{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				env.cl.EXPECT().Filtering().Return(rf, nil)
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
			t.Run("should have changes user roles", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.UserRules = utils.Ptr([]string{"foo"})
				env.cl.EXPECT().Filtering().Return(rf, nil)
				env.cl.EXPECT().SetCustomRules(env.ac.origin.filters.UserRules)
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
			t.Run("should have changed filtering config", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.Enabled = utils.Ptr(true)
				env.ac.origin.filters.Interval = utils.Ptr(123)
				env.cl.EXPECT().Filtering().Return(rf, nil)
				env.cl.EXPECT().ToggleFiltering(*env.ac.origin.filters.Enabled, *env.ac.origin.filters.Interval)
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
			t.Run("should add a filter", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.Filters = utils.Ptr([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})
				env.cl.EXPECT().Filtering().Return(rf, nil)
				env.cl.EXPECT().AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"})
				env.cl.EXPECT().RefreshFilters(gm.Any())
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
			t.Run("should delete a filter", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				rfLocal := &model.FilterStatus{Filters: utils.Ptr([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
				env.cl.EXPECT().Filtering().Return(rfLocal, nil)
				env.cl.EXPECT().DeleteFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"})
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
			t.Run("should update a filter", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.Filters = utils.Ptr(
					[]model.Filter{{Name: "foo", Url: "https://foo.bar", Enabled: true}},
				)
				rfLocal := &model.FilterStatus{Filters: utils.Ptr([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
				env.cl.EXPECT().Filtering().Return(rfLocal, nil)
				env.cl.EXPECT().UpdateFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar", Enabled: true})
				env.cl.EXPECT().RefreshFilters(gm.Any())
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})

			t.Run("should abort after failed added filter", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.cfg.ContinueOnError = false
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.Filters = utils.Ptr([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})
				env.cl.EXPECT().Filtering().Return(rf, nil)
				env.cl.EXPECT().
					AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"}).
					Return(errors.New("test failure"))
				err := actionFilters(env.ac)
				if err == nil {
					t.Error("actionFilters() error = nil, want error")
				}
			})

			t.Run("should continue after failed added filter", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.cfg.ContinueOnError = true
				env.ac.origin.filters = &model.FilterStatus{}
				env.ac.origin.filters.Filters = utils.Ptr(
					[]model.Filter{{Name: "foo", Url: "https://foo.bar"}, {Name: "bar", Url: "https://bar.foo"}},
				)
				env.cl.EXPECT().Filtering().Return(rf, nil)
				env.cl.EXPECT().
					AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"}).
					Return(errors.New("test failure"))
				env.cl.EXPECT().AddFilter(false, model.Filter{Name: "bar", Url: "https://bar.foo"})
				env.cl.EXPECT().RefreshFilters(gm.Any())
				err := actionFilters(env.ac)
				if err != nil {
					t.Errorf("actionFilters() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionDNSAccessLists", func(t *testing.T) {
			ral := &model.AccessList{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.accessList = &model.AccessList{}
				env.cl.EXPECT().AccessList().Return(ral, nil)
				err := actionDNSAccessLists(env.ac)
				if err != nil {
					t.Errorf("actionDNSAccessLists() error = %v, want nil", err)
				}
			})
			t.Run("should have access list changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.accessList = &model.AccessList{}
				ralLocal := &model.AccessList{BlockedHosts: utils.Ptr([]string{"foo"})}
				env.cl.EXPECT().AccessList().Return(ralLocal, nil)
				env.cl.EXPECT().SetAccessList(env.ac.origin.accessList)
				err := actionDNSAccessLists(env.ac)
				if err != nil {
					t.Errorf("actionDNSAccessLists() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionDNSServerConfig", func(t *testing.T) {
			rdc := &model.DNSConfig{}
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dnsConfig = &model.DNSConfig{}
				env.cl.EXPECT().DNSConfig().Return(rdc, nil)
				err := actionDNSServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDNSServerConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have dns config changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dnsConfig = &model.DNSConfig{}
				rdcLocal := &model.DNSConfig{BootstrapDns: utils.Ptr([]string{"foo"})}
				env.cl.EXPECT().DNSConfig().Return(rdcLocal, nil)
				env.cl.EXPECT().SetDNSConfig(env.ac.origin.dnsConfig)
				err := actionDNSServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDNSServerConfig() error = %v, want nil", err)
				}
			})
		})

		t.Run("actionDHCPServerConfig", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dhcpServerConfig = &model.DhcpStatus{
					V4: &model.DhcpConfigV4{
						GatewayIp:  utils.Ptr("1.2.3.4"),
						RangeStart: utils.Ptr("1.2.3.5"),
						RangeEnd:   utils.Ptr("1.2.3.6"),
						SubnetMask: utils.Ptr("255.255.255.0"),
					},
				}
				env.w.cfg.Features.DHCP.StaticLeases = false
				rsc := &model.DhcpStatus{V4: env.ac.origin.dhcpServerConfig.V4}
				env.cl.EXPECT().DhcpConfig().Return(rsc, nil)
				err := actionDHCPServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDHCPServerConfig() error = %v, want nil", err)
				}
			})
			t.Run("should have changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dhcpServerConfig = &model.DhcpStatus{
					V4: &model.DhcpConfigV4{
						GatewayIp:  utils.Ptr("1.2.3.4"),
						RangeStart: utils.Ptr("1.2.3.5"),
						RangeEnd:   utils.Ptr("1.2.3.6"),
						SubnetMask: utils.Ptr("255.255.255.0"),
					},
				}
				env.w.cfg.Features.DHCP.StaticLeases = false
				rscLocal := &model.DhcpStatus{Enabled: utils.Ptr(true)}
				env.cl.EXPECT().DhcpConfig().Return(rscLocal, nil)
				env.cl.EXPECT().SetDhcpConfig(env.ac.origin.dhcpServerConfig)
				err := actionDHCPServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDHCPServerConfig() error = %v, want nil", err)
				}
			})
			t.Run("should use replica interface name", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dhcpServerConfig = &model.DhcpStatus{
					V4: &model.DhcpConfigV4{
						GatewayIp:  utils.Ptr("1.2.3.4"),
						RangeStart: utils.Ptr("1.2.3.5"),
						RangeEnd:   utils.Ptr("1.2.3.6"),
						SubnetMask: utils.Ptr("255.255.255.0"),
					},
				}
				env.w.cfg.Features.DHCP.StaticLeases = false
				env.ac.replica.InterfaceName = "foo"
				rsc := &model.DhcpStatus{}
				env.cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := env.ac.origin.dhcpServerConfig.Clone()
				oscClone.InterfaceName = utils.Ptr("foo")
				env.cl.EXPECT().SetDhcpConfig(oscClone)
				err := actionDHCPServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDHCPServerConfig() error = %v, want nil", err)
				}
			})
			t.Run("should enable the target dhcp server", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dhcpServerConfig = &model.DhcpStatus{
					V4: &model.DhcpConfigV4{
						GatewayIp:  utils.Ptr("1.2.3.4"),
						RangeStart: utils.Ptr("1.2.3.5"),
						RangeEnd:   utils.Ptr("1.2.3.6"),
						SubnetMask: utils.Ptr("255.255.255.0"),
					},
				}
				env.w.cfg.Features.DHCP.StaticLeases = false
				env.ac.replica.DHCPServerEnabled = utils.Ptr(true)
				rsc := &model.DhcpStatus{}
				env.cl.EXPECT().DhcpConfig().Return(rsc, nil)
				oscClone := env.ac.origin.dhcpServerConfig.Clone()
				oscClone.Enabled = utils.Ptr(true)
				env.cl.EXPECT().SetDhcpConfig(oscClone)
				err := actionDHCPServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDHCPServerConfig() error = %v, want nil", err)
				}
			})
			t.Run("should not sync empty IPv4", func(t *testing.T) {
				env := newTestEnv(t)
				env.ac.origin.dhcpServerConfig = &model.DhcpStatus{
					V4: &model.DhcpConfigV4{
						GatewayIp:  utils.Ptr("1.2.3.4"),
						RangeStart: utils.Ptr("1.2.3.5"),
						RangeEnd:   utils.Ptr("1.2.3.6"),
						SubnetMask: utils.Ptr("255.255.255.0"),
					},
				}
				env.w.cfg.Features.DHCP.StaticLeases = false
				env.ac.replica.DHCPServerEnabled = utils.Ptr(false)
				env.ac.origin.dhcpServerConfig.V4 = &model.DhcpConfigV4{
					GatewayIp: utils.Ptr(""),
				}
				err := actionDHCPServerConfig(env.ac)
				if err != nil {
					t.Errorf("actionDHCPServerConfig() error = %v, want nil", err)
				}
			})
		})

		t.Run("sync", func(t *testing.T) {
			t.Run("should have no changes", func(t *testing.T) {
				env := newTestEnv(t)
				env.w.cfg = &types.Config{
					Origin:  &types.AdGuardInstance{},
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
						TLSConfig:       true,
					},
				}
				// origin
				env.cl.EXPECT().Host().Times(2)
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				env.cl.EXPECT().BlockedServicesSchedule()
				env.cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				env.cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				env.cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
				env.cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
				env.cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				env.cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				env.cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
				env.cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

				// replica
				env.cl.EXPECT().Host()
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
				env.cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
				env.cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				env.cl.EXPECT().AddRewriteEntries()
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().UpdateRewriteEntries()
				env.cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				env.cl.EXPECT().BlockedServicesSchedule()
				env.cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				env.cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				env.cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				env.cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
				env.cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)
				env.w.sync()
			})
			t.Run("should not sync DHCP", func(t *testing.T) {
				env := newTestEnv(t)
				env.w.cfg = &types.Config{
					Origin:  &types.AdGuardInstance{},
					Replica: &types.AdGuardInstance{URL: "foo"},
					Features: types.Features{
						DHCP: types.DHCP{
							ServerConfig: false,
							StaticLeases: false,
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
						TLSConfig:       true,
					},
				}
				// origin
				env.cl.EXPECT().Host().Times(2)
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				env.cl.EXPECT().BlockedServicesSchedule()
				env.cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				env.cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				env.cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
				env.cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
				env.cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				env.cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				env.cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

				// replica
				env.cl.EXPECT().Host()
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
				env.cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
				env.cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				env.cl.EXPECT().AddRewriteEntries()
				env.cl.EXPECT().DeleteRewriteEntries()
				env.cl.EXPECT().UpdateRewriteEntries()
				env.cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				env.cl.EXPECT().BlockedServicesSchedule()
				env.cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				env.cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				env.cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				env.cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)
				env.w.sync()
			})
			t.Run("origin version is too small", func(t *testing.T) {
				env := newTestEnv(t)
				env.w.cfg = &types.Config{
					Origin:  &types.AdGuardInstance{},
					Replica: &types.AdGuardInstance{URL: "foo"},
				}
				// origin
				env.cl.EXPECT().Host()
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
				env.w.sync()
			})
			t.Run("replica version is too small", func(t *testing.T) {
				env := newTestEnv(t)
				env.w.cfg = &types.Config{
					Origin:  &types.AdGuardInstance{},
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
						TLSConfig:       true,
					},
				}
				// origin
				env.cl.EXPECT().Host()
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
				env.cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
				env.cl.EXPECT().Parental()
				env.cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
				env.cl.EXPECT().SafeBrowsing()
				env.cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
				env.cl.EXPECT().BlockedServicesSchedule()
				env.cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
				env.cl.EXPECT().Clients().Return(&model.Clients{}, nil)
				env.cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
				env.cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
				env.cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
				env.cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
				env.cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
				env.cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

				// replica
				env.cl.EXPECT().Host().Times(2)
				env.cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
				env.w.sync()
			})
		})
	})
}
