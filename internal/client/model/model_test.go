package model_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func TestMergeDhcpStaticLeases(t *testing.T) {
	tests := []struct {
		name        string
		l           *[]model.DhcpStaticLease
		other       *[]model.DhcpStaticLease
		wantAdds    int
		wantRemoves int
	}{
		{
			name:        "both nil",
			l:           nil,
			other:       nil,
			wantAdds:    0,
			wantRemoves: 0,
		},
		{
			name: "add lease",
			l:    &[]model.DhcpStaticLease{},
			other: &[]model.DhcpStaticLease{
				{Mac: "mac1"},
			},
			wantAdds:    1,
			wantRemoves: 0,
		},
		{
			name: "remove lease",
			l: &[]model.DhcpStaticLease{
				{Mac: "mac1"},
			},
			other:       &[]model.DhcpStaticLease{},
			wantAdds:    0,
			wantRemoves: 1,
		},
		{
			name: "no change",
			l: &[]model.DhcpStaticLease{
				{Mac: "mac1"},
			},
			other: &[]model.DhcpStaticLease{
				{Mac: "mac1"},
			},
			wantAdds:    0,
			wantRemoves: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adds, removes := model.MergeDhcpStaticLeases(tt.l, tt.other)
			if len(adds) != tt.wantAdds {
				t.Errorf("adds length = %d, want %d", len(adds), tt.wantAdds)
			}
			if len(removes) != tt.wantRemoves {
				t.Errorf("removes length = %d, want %d", len(removes), tt.wantRemoves)
			}
		})
	}
}

func TestAccessList_Equals(t *testing.T) {
	tests := []struct {
		name string
		al   *model.AccessList
		o    *model.AccessList
		want bool
	}{
		{
			name: "equal",
			al: &model.AccessList{
				AllowedClients: &[]string{"a", "b"},
			},
			o: &model.AccessList{
				AllowedClients: &[]string{"b", "a"},
			},
			want: true,
		},
		{
			name: "not equal",
			al: &model.AccessList{
				AllowedClients: &[]string{"a", "b"},
			},
			o: &model.AccessList{
				AllowedClients: &[]string{"c", "a"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.al.Equals(tt.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSafeSearchConfig_Equals(t *testing.T) {
	tests := []struct {
		name string
		ssc  *model.SafeSearchConfig
		o    *model.SafeSearchConfig
		want bool
	}{
		{
			name: "equal",
			ssc:  &model.SafeSearchConfig{Enabled: ptr(true)},
			o:    &model.SafeSearchConfig{Enabled: ptr(true)},
			want: true,
		},
		{
			name: "not equal",
			ssc:  &model.SafeSearchConfig{Enabled: ptr(true)},
			o:    &model.SafeSearchConfig{Enabled: ptr(false)},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ssc.Equals(tt.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileInfo_Equals(t *testing.T) {
	tests := []struct {
		name      string
		pi        *model.ProfileInfo
		o         *model.ProfileInfo
		withTheme bool
		want      bool
	}{
		{
			name:      "equal with theme",
			pi:        &model.ProfileInfo{Language: "en", Theme: "dark"},
			o:         &model.ProfileInfo{Language: "en", Theme: "dark"},
			withTheme: true,
			want:      true,
		},
		{
			name:      "not equal with theme",
			pi:        &model.ProfileInfo{Language: "en", Theme: "dark"},
			o:         &model.ProfileInfo{Language: "en", Theme: "light"},
			withTheme: true,
			want:      false,
		},
		{
			name:      "equal without theme",
			pi:        &model.ProfileInfo{Language: "en", Theme: "dark"},
			o:         &model.ProfileInfo{Language: "en", Theme: "light"},
			withTheme: false,
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pi.Equals(tt.o, tt.withTheme); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProfileInfo_ShouldSyncFor(t *testing.T) {
	tests := []struct {
		name      string
		pi        *model.ProfileInfo
		o         *model.ProfileInfo
		withTheme bool
		want      *model.ProfileInfo
	}{
		{
			name:      "equal",
			pi:        &model.ProfileInfo{Name: "a", Language: "en", Theme: "light"},
			o:         &model.ProfileInfo{Name: "a", Language: "en", Theme: "light"},
			withTheme: false,
			want:      nil,
		},
		{
			name:      "should sync language",
			pi:        &model.ProfileInfo{Name: "a", Language: "en"},
			o:         &model.ProfileInfo{Name: "a", Language: "de"},
			withTheme: false,
			want:      &model.ProfileInfo{Name: "a", Language: "de"},
		},
		{
			name:      "should sync theme",
			pi:        &model.ProfileInfo{Name: "a", Language: "en", Theme: "light"},
			o:         &model.ProfileInfo{Name: "a", Language: "en", Theme: "dark"},
			withTheme: true,
			want:      &model.ProfileInfo{Name: "a", Language: "en", Theme: "dark"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pi.ShouldSyncFor(tt.o, tt.withTheme)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShouldSyncFor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockedServicesSchedule_Equals(t *testing.T) {
	tests := []struct {
		name string
		bss  *model.BlockedServicesSchedule
		o    *model.BlockedServicesSchedule
		want bool
	}{
		{
			name: "equal",
			bss:  &model.BlockedServicesSchedule{Ids: &[]string{"a"}},
			o:    &model.BlockedServicesSchedule{Ids: &[]string{"a"}},
			want: true,
		},
		{
			name: "not equal",
			bss:  &model.BlockedServicesSchedule{Ids: &[]string{"a"}},
			o:    &model.BlockedServicesSchedule{Ids: &[]string{"b"}},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bss.Equals(tt.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBlockedServicesSchedule_ServicesString(t *testing.T) {
	tests := []struct {
		name string
		bss  *model.BlockedServicesSchedule
		want string
	}{
		{
			name: "nil",
			bss:  &model.BlockedServicesSchedule{Ids: nil},
			want: "[]",
		},
		{
			name: "sorted",
			bss:  &model.BlockedServicesSchedule{Ids: &[]string{"b", "a"}},
			want: "[a,b]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bss.ServicesString(); got != tt.want {
				t.Errorf("ServicesString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStats_Add(t *testing.T) {
	s1 := model.NewStats()
	s2 := &model.Stats{
		NumDnsQueries: ptr(10),
		DnsQueries:    ptr(make([]int, 24)),
	}
	(*s2.DnsQueries)[0] = 5

	s1.Add(s2)
	if *s1.NumDnsQueries != 10 {
		t.Errorf("NumDnsQueries = %d, want 10", *s1.NumDnsQueries)
	}
	if (*s1.DnsQueries)[0] != 5 {
		t.Errorf("DnsQueries[0] = %d, want 5", (*s1.DnsQueries)[0])
	}
}

func TestTlsConfig_Equals(t *testing.T) {
	tests := []struct {
		name string
		c    *model.TlsConfig
		o    *model.TlsConfig
		want bool
	}{
		{
			name: "equal",
			c:    &model.TlsConfig{ServerName: ptr("a")},
			o:    &model.TlsConfig{ServerName: ptr("a")},
			want: true,
		},
		{
			name: "not equal",
			c:    &model.TlsConfig{ServerName: ptr("a")},
			o:    &model.TlsConfig{ServerName: ptr("b")},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Equals(tt.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStatsConfigResponse_Equals(t *testing.T) {
	tests := []struct {
		name string
		sc   *model.GetStatsConfigResponse
		o    *model.GetStatsConfigResponse
		want bool
	}{
		{
			name: "equal",
			sc:   &model.GetStatsConfigResponse{Interval: 1.0},
			o:    &model.GetStatsConfigResponse{Interval: 1.0},
			want: true,
		},
		{
			name: "not equal",
			sc:   &model.GetStatsConfigResponse{Interval: 1.0},
			o:    &model.GetStatsConfigResponse{Interval: 2.0},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sc.Equals(tt.o); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilteringStatus_ParseJSON(t *testing.T) {
	b, err := os.ReadFile("../../../testdata/filtering-status.json")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}
	fs := &model.FilterStatus{}
	if err := json.Unmarshal(b, fs); err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
}

func TestMergeFilters(t *testing.T) {
	url := "https://" + uuid.NewString()

	tests := []struct {
		name           string
		originFilters  []model.Filter
		replicaFilters []model.Filter
		wantAdded      int
		wantUpdated    int
		wantDeleted    int
		check          func(t *testing.T, a, u, d []model.Filter)
	}{
		{
			name:           "should add a missing filter",
			originFilters:  []model.Filter{{Url: url}},
			replicaFilters: []model.Filter{},
			wantAdded:      1,
			check: func(t *testing.T, a, _, _ []model.Filter) {
				t.Helper()
				if a[0].Url != url {
					t.Errorf("expected added filter URL %s, got %s", url, a[0].Url)
				}
			},
		},
		{
			name:           "should remove additional filter",
			originFilters:  []model.Filter{},
			replicaFilters: []model.Filter{{Url: url}},
			wantDeleted:    1,
			check: func(t *testing.T, _, _, d []model.Filter) {
				t.Helper()
				if d[0].Url != url {
					t.Errorf("expected deleted filter URL %s, got %s", url, d[0].Url)
				}
			},
		},
		{
			name: "should update existing filter when enabled differs",
			originFilters: []model.Filter{
				{Url: url, Enabled: true},
			},
			replicaFilters: []model.Filter{
				{Url: url, Enabled: false},
			},
			wantUpdated: 1,
			check: func(t *testing.T, _, u, _ []model.Filter) {
				t.Helper()
				if !u[0].Enabled {
					t.Error("expected updated filter to be enabled")
				}
			},
		},
		{
			name: "should update existing filter when name differs",
			originFilters: []model.Filter{
				{Url: url, Name: "name1"},
			},
			replicaFilters: []model.Filter{
				{Url: url, Name: "name2"},
			},
			wantUpdated: 1,
			check: func(t *testing.T, _, u, _ []model.Filter) {
				t.Helper()
				if u[0].Name != "name1" {
					t.Errorf("expected updated filter name 'name1', got '%s'", u[0].Name)
				}
			},
		},
		{
			name: "should have no changes",
			originFilters: []model.Filter{
				{Url: url},
			},
			replicaFilters: []model.Filter{
				{Url: url},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, u, d := model.MergeFilters(&tt.replicaFilters, &tt.originFilters)
			if len(a) != tt.wantAdded {
				t.Errorf("added = %d, want %d", len(a), tt.wantAdded)
			}
			if len(u) != tt.wantUpdated {
				t.Errorf("updated = %d, want %d", len(u), tt.wantUpdated)
			}
			if len(d) != tt.wantDeleted {
				t.Errorf("deleted = %d, want %d", len(d), tt.wantDeleted)
			}
			if tt.check != nil {
				tt.check(t, a, u, d)
			}
		})
	}
}

func TestAdGuardInstance_Key(t *testing.T) {
	url := "https://" + uuid.NewString()
	apiPath := "/" + uuid.NewString()
	i := &types.AdGuardInstance{URL: url, APIPath: apiPath}
	want := url + "#" + apiPath
	if got := i.Key(); got != want {
		t.Errorf("Key() = %s, want %s", got, want)
	}
}

func TestRewriteEntry_Key(t *testing.T) {
	domain := uuid.NewString()
	answer := uuid.NewString()
	re := &model.RewriteEntry{Domain: &domain, Answer: &answer}
	want := domain + "#" + answer
	if got := re.Key(); got != want {
		t.Errorf("Key() = %s, want %s", got, want)
	}
}

func TestQueryLogConfigWithIgnored_Equals(t *testing.T) {
	var interval1 model.QueryLogConfigInterval = 1.0
	var interval2 model.QueryLogConfigInterval = 2.0

	tests := []struct {
		name string
		a    model.QueryLogConfigWithIgnored
		b    model.QueryLogConfigWithIgnored
		want bool
	}{
		{
			name: "should be equal",
			a: model.QueryLogConfigWithIgnored{
				QueryLogConfig: model.QueryLogConfig{
					Enabled:           ptr(true),
					Interval:          &interval1,
					AnonymizeClientIp: ptr(true),
				},
			},
			b: model.QueryLogConfigWithIgnored{
				QueryLogConfig: model.QueryLogConfig{
					Enabled:           ptr(true),
					Interval:          &interval1,
					AnonymizeClientIp: ptr(true),
				},
			},
			want: true,
		},
		{
			name: "should not be equal when enabled differs",
			a:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Enabled: ptr(true)}},
			b:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Enabled: ptr(false)}},
			want: false,
		},
		{
			name: "should not be equal when interval differs",
			a:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Interval: &interval1}},
			b:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Interval: &interval2}},
			want: false,
		},
		{
			name: "should not be equal when anonymizeClientIP differs",
			a:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{AnonymizeClientIp: ptr(true)}},
			b:    model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{AnonymizeClientIp: ptr(false)}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Equals(&tt.b); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRewriteEntries_Merge(t *testing.T) {
	domain := uuid.NewString()

	tests := []struct {
		name        string
		originRE    model.RewriteEntries
		replicaRE   model.RewriteEntries
		wantAdded   int
		wantRemoved int
		wantDeleted int
		wantUpdated int
	}{
		{
			name:      "should add a missing rewrite entry",
			originRE:  model.RewriteEntries{{Domain: &domain}},
			replicaRE: model.RewriteEntries{},
			wantAdded: 1,
		},
		{
			name:        "should remove additional rewrite entry",
			originRE:    model.RewriteEntries{},
			replicaRE:   model.RewriteEntries{{Domain: &domain}},
			wantRemoved: 1,
		},
		{
			name:      "should have no changes",
			originRE:  model.RewriteEntries{{Domain: &domain}},
			replicaRE: model.RewriteEntries{{Domain: &domain}},
		},
		{
			name:        "should remove target duplicate",
			originRE:    model.RewriteEntries{{Domain: &domain}},
			replicaRE:   model.RewriteEntries{{Domain: &domain}, {Domain: &domain}},
			wantRemoved: 1,
		},
		{
			name:        "should remove target duplicate",
			originRE:    model.RewriteEntries{{Domain: &domain}, {Domain: &domain}},
			replicaRE:   model.RewriteEntries{{Domain: &domain}},
			wantDeleted: 1,
		},
		{
			name:        "should update a changed",
			originRE:    model.RewriteEntries{{Domain: &domain, Enabled: ptr(true)}},
			replicaRE:   model.RewriteEntries{{Domain: &domain, Enabled: ptr(false)}},
			wantUpdated: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, r, d, u := tt.replicaRE.Merge(&tt.originRE)
			if len(a) != tt.wantAdded {
				t.Errorf("added = %d, want %d", len(a), tt.wantAdded)
			}
			if len(r) != tt.wantRemoved {
				t.Errorf("removed = %d, want %d", len(r), tt.wantRemoved)
			}
			if len(d) != tt.wantDeleted {
				t.Errorf("deleted = %d, want %d", len(d), tt.wantDeleted)
			}
			if len(u) != tt.wantUpdated {
				t.Errorf("updated = %d, want %d", len(u), tt.wantUpdated)
			}
		})
	}
}

func TestConfig_UniqueReplicas(t *testing.T) {
	url := "https://" + uuid.NewString()
	apiPath := "/" + uuid.NewString()

	tests := []struct {
		name string
		cfg  types.Config
		want int
	}{
		{
			name: "should be empty if nothing defined",
			cfg:  types.Config{},
			want: 0,
		},
		{
			name: "should be empty if replica url is not set",
			cfg:  types.Config{Replica: &types.AdGuardInstance{URL: ""}},
			want: 0,
		},
		{
			name: "should be empty if replicas url is not set",
			cfg:  types.Config{Replicas: []types.AdGuardInstance{{URL: ""}}},
			want: 0,
		},
		{
			name: "should return only one replica if same url and apiPath",
			cfg: types.Config{
				Replica: &types.AdGuardInstance{URL: url, APIPath: apiPath},
				Replicas: []types.AdGuardInstance{
					{URL: url, APIPath: apiPath},
					{URL: url, APIPath: apiPath},
				},
			},
			want: 1,
		},
		{
			name: "should return 3 one replicas if urls are different",
			cfg: types.Config{
				Replica: &types.AdGuardInstance{URL: url, APIPath: apiPath},
				Replicas: []types.AdGuardInstance{
					{URL: url + "1", APIPath: apiPath},
					{URL: url, APIPath: apiPath + "1"},
				},
			},
			want: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.cfg.UniqueReplicas()
			if len(r) != tt.want {
				t.Errorf("UniqueReplicas() len = %d, want %d", len(r), tt.want)
			}
		})
	}

	t.Run("should set default api apiPath if not set", func(t *testing.T) {
		cfg := types.Config{
			Replica:  &types.AdGuardInstance{URL: url},
			Replicas: []types.AdGuardInstance{{URL: url + "1"}},
		}
		r := cfg.UniqueReplicas()
		if len(r) != 2 {
			t.Errorf("len = %d, want 2", len(r))
		}
		if r[0].APIPath != types.DefaultAPIPath {
			t.Errorf("APIPath[0] = %s, want %s", r[0].APIPath, types.DefaultAPIPath)
		}
		if r[1].APIPath != types.DefaultAPIPath {
			t.Errorf("APIPath[1] = %s, want %s", r[1].APIPath, types.DefaultAPIPath)
		}
	})
}

func TestClients_Merge(t *testing.T) {
	name := uuid.NewString()
	disallowed := true

	tests := []struct {
		name           string
		originClients  func() *model.Clients
		replicaClients model.Clients
		wantAdded      int
		wantUpdated    int
		wantDeleted    int
		check          func(t *testing.T, a, u, d []*model.Client)
	}{
		{
			name: "should add a missing client",
			originClients: func() *model.Clients {
				c := &model.Clients{}
				c.Add(model.Client{Name: &name})
				return c
			},
			replicaClients: model.Clients{},
			wantAdded:      1,
			check: func(t *testing.T, a, _, _ []*model.Client) {
				t.Helper()
				if *a[0].Name != name {
					t.Errorf("expected added client name %s, got %s", name, *a[0].Name)
				}
			},
		},
		{
			name: "should remove additional client",
			originClients: func() *model.Clients {
				return &model.Clients{}
			},
			replicaClients: func() model.Clients {
				c := model.Clients{}
				c.Add(model.Client{Name: &name})
				return c
			}(),
			wantDeleted: 1,
			check: func(t *testing.T, _, _, d []*model.Client) {
				t.Helper()
				if *d[0].Name != name {
					t.Errorf("expected deleted client name %s, got %s", name, *d[0].Name)
				}
			},
		},
		{
			name: "should update existing client when name differs",
			originClients: func() *model.Clients {
				c := &model.Clients{}
				c.Add(model.Client{Name: &name, FilteringEnabled: &disallowed})
				return c
			},
			replicaClients: func() model.Clients {
				c := model.Clients{}
				enabled := !disallowed
				c.Add(model.Client{Name: &name, FilteringEnabled: &enabled})
				return c
			}(),
			wantUpdated: 1,
			check: func(t *testing.T, _, u, _ []*model.Client) {
				t.Helper()
				if *u[0].FilteringEnabled != disallowed {
					t.Errorf("expected updated client filtering enabled %v, got %v", disallowed, *u[0].FilteringEnabled)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, u, d := tt.replicaClients.Merge(tt.originClients())
			if len(a) != tt.wantAdded {
				t.Errorf("added = %d, want %d", len(a), tt.wantAdded)
			}
			if len(u) != tt.wantUpdated {
				t.Errorf("updated = %d, want %d", len(u), tt.wantUpdated)
			}
			if len(d) != tt.wantDeleted {
				t.Errorf("deleted = %d, want %d", len(d), tt.wantDeleted)
			}
			if tt.check != nil {
				tt.check(t, a, u, d)
			}
		})
	}
}

func TestClient_Equals(t *testing.T) {
	cl1 := &model.Client{
		Name:                    ptr("foo"),
		BlockedServicesSchedule: &model.Schedule{TimeZone: ptr("UTC")},
	}
	cl2 := &model.Client{
		Name:                    ptr("foo"),
		BlockedServicesSchedule: &model.Schedule{TimeZone: ptr("Local")},
	}

	if !cl1.Equals(cl2) {
		t.Error("should equal if only timezone differs on empty blocked service schedule")
	}
}

func TestBlockedServices_Equals(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
		s1 := &model.BlockedServicesArray{"a", "b"}
		s2 := &model.BlockedServicesArray{"b", "a"}
		if !model.EqualsStringSlice(s1, s2, true) {
			t.Error("expected equal")
		}
	})
	t.Run("should not be equal different values", func(t *testing.T) {
		s1 := &model.BlockedServicesArray{"a", "b"}
		s2 := &model.BlockedServicesArray{"B", "a"}
		if model.EqualsStringSlice(s1, s2, true) {
			t.Error("expected not equal")
		}
	})
	t.Run("should not be equal different length", func(t *testing.T) {
		s1 := &model.BlockedServicesArray{"a", "b"}
		s2 := &model.BlockedServicesArray{"b", "a", "c"}
		if model.EqualsStringSlice(s1, s2, true) {
			t.Error("expected not equal")
		}
	})
}

func TestDNSConfig_Equals(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
		dc1 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"a"})}
		dc2 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"a"})}
		if !dc1.Equals(dc2) {
			t.Error("expected equal")
		}
	})
	t.Run("should not be equal", func(t *testing.T) {
		dc1 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"a"})}
		dc2 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"b"})}
		if dc1.Equals(dc2) {
			t.Error("expected not equal")
		}
	})
}

func TestDHCPServerConfig(t *testing.T) {
	t.Run("Equals", func(t *testing.T) {
		dc1 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:     ptr("1.2.3.4"),
				LeaseDuration: ptr(123),
				RangeStart:    ptr("1.2.3.5"),
				RangeEnd:      ptr("1.2.3.6"),
				SubnetMask:    ptr("255.255.255.0"),
			},
		}
		dc2 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:     ptr("1.2.3.4"),
				LeaseDuration: ptr(123),
				RangeStart:    ptr("1.2.3.5"),
				RangeEnd:      ptr("1.2.3.6"),
				SubnetMask:    ptr("255.255.255.0"),
			},
		}
		if !dc1.Equals(dc2) {
			t.Error("expected equal")
		}

		dc1.V4.GatewayIp = ptr("1.2.3.3")
		if dc1.Equals(dc2) {
			t.Error("expected not equal")
		}
	})

	t.Run("Clone should be equal", func(t *testing.T) {
		dc1 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:     ptr("1.2.3.4"),
				LeaseDuration: ptr(123),
				RangeStart:    ptr("1.2.3.5"),
				RangeEnd:      ptr("1.2.3.6"),
				SubnetMask:    ptr("255.255.255.0"),
			},
		}
		if !dc1.Clone().Equals(dc1) {
			t.Error("cloned config should be equal")
		}
	})

	t.Run("HasConfig", func(t *testing.T) {
		dc1 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{},
			V6: &model.DhcpConfigV6{},
		}
		if dc1.HasConfig() {
			t.Error("expected no config")
		}

		dc1.V6.RangeStart = ptr("1.2.3.5")
		if !dc1.HasConfig() {
			t.Error("expected config")
		}

		dc1.V4.GatewayIp = ptr("")
		if !dc1.HasConfig() {
			t.Error("expected config")
		}

		dc1 = &model.DhcpStatus{
			V4: &model.DhcpConfigV4{
				GatewayIp:     ptr("1.2.3.4"),
				LeaseDuration: ptr(123),
				RangeStart:    ptr("1.2.3.5"),
				RangeEnd:      ptr("1.2.3.6"),
				SubnetMask:    ptr("255.255.255.0"),
			},
			V6: &model.DhcpConfigV6{},
		}
		if !dc1.HasConfig() {
			t.Error("expected config")
		}
	})

	t.Run("CleanAndEquals", func(t *testing.T) {
		ds1 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{},
		}
		ds2 := &model.DhcpStatus{
			V6: &model.DhcpConfigV6{},
		}
		// both should be cleaned to nil and then be equal
		if !ds1.CleanAndEquals(ds2) {
			t.Error("expected CleanAndEquals to be true")
		}
	})

	t.Run("EqualsStringSlice_Branches", func(t *testing.T) {
		if model.EqualsStringSlice(nil, &[]string{"a"}, true) {
			t.Error("expected false")
		}
		if model.EqualsStringSlice(&[]string{"a"}, nil, true) {
			t.Error("expected false")
		}
	})

	t.Run("Clients_Add", func(t *testing.T) {
		cs := &model.Clients{}
		cs.Add(model.Client{Name: ptr("c1")})
		if len(*cs.Clients) != 1 {
			t.Errorf("expected 1 client, got %d", len(*cs.Clients))
		}
		cs.Add(model.Client{Name: ptr("c2")})
		if len(*cs.Clients) != 2 {
			t.Errorf("expected 2 clients, got %d", len(*cs.Clients))
		}
	})

	t.Run("MergeFilters_Branches", func(t *testing.T) {
		a, u, r := model.MergeFilters(nil, nil)
		if a != nil || u != nil || r != nil {
			t.Error("expected nil results")
		}
	})

	t.Run("QueryLogConfigInterval_Equals", func(t *testing.T) {
		var i1 model.QueryLogConfigInterval = 1.0
		var i2 model.QueryLogConfigInterval = 1.0
		var i3 model.QueryLogConfigInterval = 2.0
		if !i1.Equals(&i2) {
			t.Error("expected true")
		}
		if i1.Equals(&i3) {
			t.Error("expected false")
		}
	})

	t.Run("ProfileInfo_ShouldSyncFor_Branches", func(t *testing.T) {
		pi := &model.ProfileInfo{Name: "n1", Language: "en"}
		if pi.ShouldSyncFor(&model.ProfileInfo{Name: "n1", Language: "en"}, false) != nil {
			t.Error("expected nil")
		}
		if pi.ShouldSyncFor(&model.ProfileInfo{Name: "", Language: "en"}, false) != nil {
			t.Error("expected nil for empty name")
		}
		if pi.ShouldSyncFor(&model.ProfileInfo{Name: "n1", Language: ""}, false) != nil {
			t.Error("expected nil for empty language")
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}
