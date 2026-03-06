package sync

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	gm "go.uber.org/mock/gomock"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/client/model"
	clientmock "github.com/bakito/adguardhome-sync/internal/mocks/client"
	"github.com/bakito/adguardhome-sync/internal/types"
	"github.com/bakito/adguardhome-sync/internal/versions"
)

func setup(t *testing.T) (*gm.Controller, *clientmock.MockClient, *worker, *actionContext, error) {
	t.Helper()
	mockCtrl := gm.NewController(t)
	cl := clientmock.NewMockClient(mockCtrl)
	w := &worker{
		createClient: func(_ types.AdGuardInstance, _ time.Duration) (client.Client, error) {
			return cl, nil
		},
		cfg: &types.Config{
			Origin: &types.AdGuardInstance{},
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
				TLSConfig:       true,
			},
			Replicas: []types.AdGuardInstance{
				{URL: "http://replica"},
			},
		},
	}
	te := errors.New(uuid.NewString())

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
	return mockCtrl, cl, w, ac, te
}

func Test_actionDNSRewrites(t *testing.T) {
	t.Run("should have no changes (empty slices)", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		reR := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().AddRewriteEntries()
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().UpdateRewriteEntries()
		err := actionDNSRewrites(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should add one rewrite entry", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		reR := model.RewriteEntries{}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().AddRewriteEntries(reO[0])
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().UpdateRewriteEntries()
		err := actionDNSRewrites(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should remove one rewrite entry", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{}
		reR := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().AddRewriteEntries()
		cl.EXPECT().DeleteRewriteEntries(reR[0])
		cl.EXPECT().UpdateRewriteEntries()
		err := actionDNSRewrites(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should update one rewrite entry", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer), Enabled: new(false)}}
		reR := model.RewriteEntries{{Domain: new(domain), Answer: new(answer), Enabled: new(true)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().AddRewriteEntries()
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().UpdateRewriteEntries(gm.Any())
		err := actionDNSRewrites(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should return error when error on RewriteList()", func(t *testing.T) {
		ctrl, cl, _, ac, te := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(nil, te)
		err := actionDNSRewrites(ac)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when error on AddRewriteEntries()", func(t *testing.T) {
		ctrl, cl, _, ac, te := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		reR := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().AddRewriteEntries().Return(te)
		err := actionDNSRewrites(ac)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	t.Run("should return error when error on DeleteRewriteEntries()", func(t *testing.T) {
		ctrl, cl, _, ac, te := setup(t)
		defer ctrl.Finish()
		domain := uuid.NewString()
		answer := uuid.NewString()
		reO := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		reR := model.RewriteEntries{{Domain: new(domain), Answer: new(answer)}}
		ac.origin.rewrites = &reO
		cl.EXPECT().RewriteList().Return(&reR, nil)
		cl.EXPECT().DeleteRewriteEntries().Return(te)
		err := actionDNSRewrites(ac)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func Test_actionClientSettings(t *testing.T) {
	t.Run("should have no changes (empty slices)", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		name := uuid.NewString()
		ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: new(name)}}}
		clR := &model.Clients{Clients: &model.ClientsArray{{Name: new(name)}}}
		cl.EXPECT().Clients().Return(clR, nil)
		err := actionClientSettings(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should add one client", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		name := uuid.NewString()
		ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: new(name)}}}
		clR := &model.Clients{Clients: &model.ClientsArray{}}
		cl.EXPECT().Clients().Return(clR, nil)
		cl.EXPECT().AddClient(&(*ac.origin.clients.Clients)[0])
		err := actionClientSettings(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should update one client", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		name := uuid.NewString()
		ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{{Name: new(name)}}}
		clR := &model.Clients{Clients: &model.ClientsArray{{Name: new(name), FilteringEnabled: new(true)}}}
		cl.EXPECT().Clients().Return(clR, nil)
		cl.EXPECT().UpdateClient(&(*ac.origin.clients.Clients)[0])
		err := actionClientSettings(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should delete one client", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		name := uuid.NewString()
		ac.origin.clients = &model.Clients{Clients: &model.ClientsArray{}}
		clR := &model.Clients{Clients: &model.ClientsArray{{Name: new(name)}}}
		cl.EXPECT().Clients().Return(clR, nil)
		cl.EXPECT().DeleteClient(&(*clR.Clients)[0])
		err := actionClientSettings(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should return error when error on Clients()", func(t *testing.T) {
		ctrl, cl, _, ac, te := setup(t)
		defer ctrl.Finish()
		cl.EXPECT().Clients().Return(nil, te)
		err := actionClientSettings(ac)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func Test_actionParental(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		cl.EXPECT().Parental()
		err := actionParental(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have parental enabled changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.parental = true
		cl.EXPECT().Parental()
		cl.EXPECT().ToggleParental(true)
		err := actionParental(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionProtection(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, _, _, ac, _ := setup(t)
		defer ctrl.Finish()
		err := actionProtection(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have protection enabled changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.status.ProtectionEnabled = true
		cl.EXPECT().ToggleProtection(true)
		err := actionProtection(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionSafeSearchConfig(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		cl.EXPECT().SafeSearchConfig().Return(ac.origin.safeSearch, nil)

		err := actionSafeSearchConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have safeSearch enabled changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.safeSearch = &model.SafeSearchConfig{Enabled: new(true)}
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SetSafeSearchConfig(ac.origin.safeSearch)
		err := actionSafeSearchConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have Duckduckgo safeSearch enabled changed", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.safeSearch = &model.SafeSearchConfig{Duckduckgo: new(true)}
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{Google: new(true)}, nil)
		cl.EXPECT().SetSafeSearchConfig(ac.origin.safeSearch)
		err := actionSafeSearchConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionProfileInfo(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		cl.EXPECT().ProfileInfo().Return(ac.origin.profileInfo, nil)
		err := actionProfileInfo(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have profileInfo language changed", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.profileInfo.Language = "de"
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en"}, nil)
		cl.EXPECT().SetProfileInfo(&model.ProfileInfo{
			Language: "de",
			Name:     "replica",
			Theme:    "auto",
		})
		err := actionProfileInfo(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should not change theme if feature is disabled", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.profileInfo.Language = "de"
		ac.cfg.Features.Theme = false
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{Name: "replica", Language: "en"}, nil)
		cl.EXPECT().SetProfileInfo(&model.ProfileInfo{
			Language: "de",
			Name:     "replica",
			Theme:    "",
		})
		err := actionProfileInfo(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should not sync profileInfo if language is not set", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.profileInfo.Language = ""
		cl.EXPECT().
			ProfileInfo().
			Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
		cl.EXPECT().SetProfileInfo(ac.origin.profileInfo).Times(0)
		err := actionProfileInfo(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should not sync profileInfo if theme is not set", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.profileInfo.Theme = ""
		cl.EXPECT().
			ProfileInfo().
			Return(&model.ProfileInfo{Name: "replica", Language: "en", Theme: "auto"}, nil)
		cl.EXPECT().SetProfileInfo(ac.origin.profileInfo).Times(0)
		err := actionProfileInfo(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionSafeBrowsing(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		cl.EXPECT().SafeBrowsing()
		err := actionSafeBrowsing(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should have safeBrowsing enabled changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.safeBrowsing = true
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().ToggleSafeBrowsing(true)
		err := actionSafeBrowsing(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionQueryLogConfig(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		qlc := &model.QueryLogConfigWithIgnored{}
		cl.EXPECT().QueryLogConfig().Return(qlc, nil)
		err := actionQueryLogConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have QueryLogConfig changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		qlc := &model.QueryLogConfigWithIgnored{}
		var interval model.QueryLogConfigInterval = 123
		ac.origin.queryLogConfig.Interval = &interval
		cl.EXPECT().QueryLogConfig().Return(qlc, nil)
		cl.EXPECT().
			SetQueryLogConfig(&model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{AnonymizeClientIp: nil, Interval: &interval, Enabled: nil}})
		err := actionQueryLogConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionStatsConfig(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		sc := &model.PutStatsConfigUpdateRequest{}
		cl.EXPECT().StatsConfig().Return(sc, nil)
		err := actionStatsConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have StatsConfig changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		sc := &model.PutStatsConfigUpdateRequest{}
		var interval float32 = 123
		ac.origin.statsConfig.Interval = interval
		cl.EXPECT().StatsConfig().Return(sc, nil)
		cl.EXPECT().SetStatsConfig(&model.PutStatsConfigUpdateRequest{Interval: interval})
		err := actionStatsConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_statusWithSetup(t *testing.T) {
	t.Run("should get the replica status", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		status := &model.ServerStatus{}
		inst := types.AdGuardInstance{
			AutoSetup: true,
		}
		cl.EXPECT().Status().Return(status, nil)
		st, err := w.statusWithSetup(l, inst, cl)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if st != status {
			t.Errorf("expected %v, got %v", status, st)
		}
	})
	t.Run("should runs setup before getting replica status", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		status := &model.ServerStatus{}
		inst := types.AdGuardInstance{
			AutoSetup: true,
		}
		cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
		cl.EXPECT().Setup()
		cl.EXPECT().Status().Return(status, nil)
		st, err := w.statusWithSetup(l, inst, cl)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if st != status {
			t.Errorf("expected %v, got %v", status, st)
		}
	})
	t.Run("should fail on setup", func(t *testing.T) {
		ctrl, cl, w, _, te := setup(t)
		defer ctrl.Finish()
		inst := types.AdGuardInstance{
			AutoSetup: true,
		}
		cl.EXPECT().Status().Return(nil, client.ErrSetupNeeded)
		cl.EXPECT().Setup().Return(te)
		st, err := w.statusWithSetup(l, inst, cl)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if st != nil {
			t.Errorf("expected nil, got %v", st)
		}
	})
}

func Test_actionBlockedServicesSchedule(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{}
		cl.EXPECT().BlockedServicesSchedule().Return(ac.origin.blockedServicesSchedule, nil)
		err := actionBlockedServicesSchedule(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have blockedServices schedule changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.blockedServicesSchedule = &model.BlockedServicesSchedule{Ids: new([]string{"bar"})}
		rbss := &model.BlockedServicesSchedule{}

		cl.EXPECT().BlockedServicesSchedule().Return(rbss, nil)
		cl.EXPECT().SetBlockedServicesSchedule(ac.origin.blockedServicesSchedule)
		err := actionBlockedServicesSchedule(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionFilters(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have changes user roles", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{UserRules: new([]string{"foo"})}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().SetCustomRules(ac.origin.filters.UserRules)
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have changed filtering config", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{Enabled: new(true), Interval: new(123)}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().ToggleFiltering(*ac.origin.filters.Enabled, *ac.origin.filters.Interval)
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should add a filter", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"})
		cl.EXPECT().RefreshFilters(gm.Any())
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should delete a filter", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{}
		rf := &model.FilterStatus{Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().DeleteFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"})
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should update a filter", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.filters = &model.FilterStatus{
			Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar", Enabled: true}}),
		}
		rf := &model.FilterStatus{Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().UpdateFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar", Enabled: true})
		cl.EXPECT().RefreshFilters(gm.Any())
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("should abort after failed added filter", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.cfg.ContinueOnError = false
		ac.origin.filters = &model.FilterStatus{Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar"}})}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().
			AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"}).
			Return(errors.New("test failure"))
		err := actionFilters(ac)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("should continue after failed added filter", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.cfg.ContinueOnError = true
		ac.origin.filters = &model.FilterStatus{
			Filters: new([]model.Filter{{Name: "foo", Url: "https://foo.bar"}, {Name: "bar", Url: "https://bar.foo"}}),
		}
		rf := &model.FilterStatus{}
		cl.EXPECT().Filtering().Return(rf, nil)
		cl.EXPECT().
			AddFilter(false, model.Filter{Name: "foo", Url: "https://foo.bar"}).
			Return(errors.New("test failure"))
		cl.EXPECT().AddFilter(false, model.Filter{Name: "bar", Url: "https://bar.foo"})
		cl.EXPECT().RefreshFilters(gm.Any())
		err := actionFilters(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionDNSAccessLists(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.accessList = &model.AccessList{}
		ral := &model.AccessList{}
		cl.EXPECT().AccessList().Return(ral, nil)
		err := actionDNSAccessLists(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have access list changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.accessList = &model.AccessList{}
		ral := &model.AccessList{BlockedHosts: new([]string{"foo"})}
		cl.EXPECT().AccessList().Return(ral, nil)
		cl.EXPECT().SetAccessList(ac.origin.accessList)
		err := actionDNSAccessLists(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionDNSServerConfig(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dnsConfig = &model.DNSConfig{}
		rdc := &model.DNSConfig{}
		cl.EXPECT().DNSConfig().Return(rdc, nil)
		err := actionDNSServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have dns config changes", func(t *testing.T) {
		ctrl, cl, _, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dnsConfig = &model.DNSConfig{}
		rdc := &model.DNSConfig{BootstrapDns: new([]string{"foo"})}
		cl.EXPECT().DNSConfig().Return(rdc, nil)
		cl.EXPECT().SetDNSConfig(ac.origin.dnsConfig)
		err := actionDNSServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_actionDHCPServerConfig(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, w, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dhcpServerConfig = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:  new("1.2.3.4"),
				RangeStart: new("1.2.3.5"),
				RangeEnd:   new("1.2.3.6"),
				SubnetMask: new("255.255.255.0"),
			},
		}
		rsc := &model.DhcpStatus{V4: ac.origin.dhcpServerConfig.V4}
		w.cfg.Features.DHCP.StaticLeases = false
		cl.EXPECT().DhcpConfig().Return(rsc, nil)
		err := actionDHCPServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should have changes", func(t *testing.T) {
		ctrl, cl, w, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dhcpServerConfig = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:  new("1.2.3.4"),
				RangeStart: new("1.2.3.5"),
				RangeEnd:   new("1.2.3.6"),
				SubnetMask: new("255.255.255.0"),
			},
		}
		rsc := &model.DhcpStatus{Enabled: new(true)}
		w.cfg.Features.DHCP.StaticLeases = false
		cl.EXPECT().DhcpConfig().Return(rsc, nil)
		cl.EXPECT().SetDhcpConfig(ac.origin.dhcpServerConfig)
		err := actionDHCPServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should use replica interface name", func(t *testing.T) {
		ctrl, cl, w, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dhcpServerConfig = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:  new("1.2.3.4"),
				RangeStart: new("1.2.3.5"),
				RangeEnd:   new("1.2.3.6"),
				SubnetMask: new("255.255.255.0"),
			},
		}
		rsc := &model.DhcpStatus{}
		w.cfg.Features.DHCP.StaticLeases = false
		ac.replica.InterfaceName = "foo"
		cl.EXPECT().DhcpConfig().Return(rsc, nil)
		oscClone := ac.origin.dhcpServerConfig.Clone()
		oscClone.InterfaceName = new("foo")
		cl.EXPECT().SetDhcpConfig(oscClone)
		err := actionDHCPServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should enable the target dhcp server", func(t *testing.T) {
		ctrl, cl, w, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dhcpServerConfig = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:  new("1.2.3.4"),
				RangeStart: new("1.2.3.5"),
				RangeEnd:   new("1.2.3.6"),
				SubnetMask: new("255.255.255.0"),
			},
		}
		rsc := &model.DhcpStatus{}
		w.cfg.Features.DHCP.StaticLeases = false
		ac.replica.DHCPServerEnabled = new(true)
		cl.EXPECT().DhcpConfig().Return(rsc, nil)
		oscClone := ac.origin.dhcpServerConfig.Clone()
		oscClone.Enabled = new(true)
		cl.EXPECT().SetDhcpConfig(oscClone)
		err := actionDHCPServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
	t.Run("should not sync empty IPv4", func(t *testing.T) {
		ctrl, _, w, ac, _ := setup(t)
		defer ctrl.Finish()
		ac.origin.dhcpServerConfig = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp: new(""),
			},
		}
		w.cfg.Features.DHCP.StaticLeases = false
		ac.replica.DHCPServerEnabled = new(false)
		err := actionDHCPServerConfig(ac)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func Test_worker_sync(t *testing.T) {
	t.Run("should have no changes", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		w.cfg = &types.Config{
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
		cl.EXPECT().Host().Times(2)
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
		cl.EXPECT().Parental()
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
		cl.EXPECT().BlockedServicesSchedule()
		cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
		cl.EXPECT().Clients().Return(&model.Clients{}, nil)
		cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
		cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
		cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
		cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
		cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
		cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

		// replica
		cl.EXPECT().Host()
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
		cl.EXPECT().Parental()
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
		cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
		cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
		cl.EXPECT().AddRewriteEntries()
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().UpdateRewriteEntries()
		cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
		cl.EXPECT().BlockedServicesSchedule()
		cl.EXPECT().Clients().Return(&model.Clients{}, nil)
		cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
		cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
		cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
		cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)
		w.sync()
	})
	t.Run("should not sync DHCP", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		w.cfg = &types.Config{
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
		cl.EXPECT().Host().Times(2)
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
		cl.EXPECT().Parental()
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
		cl.EXPECT().BlockedServicesSchedule()
		cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
		cl.EXPECT().Clients().Return(&model.Clients{}, nil)
		cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
		cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
		cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
		cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
		cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

		// replica
		cl.EXPECT().Host()
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
		cl.EXPECT().Parental()
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
		cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
		cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
		cl.EXPECT().AddRewriteEntries()
		cl.EXPECT().DeleteRewriteEntries()
		cl.EXPECT().UpdateRewriteEntries()
		cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
		cl.EXPECT().BlockedServicesSchedule()
		cl.EXPECT().Clients().Return(&model.Clients{}, nil)
		cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
		cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
		cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)
		w.sync()
	})
	t.Run("origin version is too small", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		// origin
		cl.EXPECT().Host()
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
		w.sync()
	})
	t.Run("replica version is too small", func(t *testing.T) {
		ctrl, cl, w, _, _ := setup(t)
		defer ctrl.Finish()
		// origin
		cl.EXPECT().Host()
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: versions.MinAgh}, nil)
		cl.EXPECT().ProfileInfo().Return(&model.ProfileInfo{}, nil)
		cl.EXPECT().Parental()
		cl.EXPECT().SafeSearchConfig().Return(&model.SafeSearchConfig{}, nil)
		cl.EXPECT().SafeBrowsing()
		cl.EXPECT().RewriteList().Return(&model.RewriteEntries{}, nil)
		cl.EXPECT().BlockedServicesSchedule()
		cl.EXPECT().Filtering().Return(&model.FilterStatus{}, nil)
		cl.EXPECT().Clients().Return(&model.Clients{}, nil)
		cl.EXPECT().QueryLogConfig().Return(&model.QueryLogConfigWithIgnored{}, nil)
		cl.EXPECT().StatsConfig().Return(&model.PutStatsConfigUpdateRequest{}, nil)
		cl.EXPECT().AccessList().Return(&model.AccessList{}, nil)
		cl.EXPECT().DNSConfig().Return(&model.DNSConfig{}, nil)
		cl.EXPECT().DhcpConfig().Return(&model.DhcpStatus{}, nil)
		cl.EXPECT().TLSConfig().Return(&model.TlsConfig{}, nil)

		// replica
		cl.EXPECT().Host().Times(2)
		cl.EXPECT().Status().Return(&model.ServerStatus{Version: "v0.106.9"}, nil)
		w.sync()
	})
}
