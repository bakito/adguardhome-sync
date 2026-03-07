package model_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/types"
)

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
	var interval1 model.QueryLogConfigInterval = 1
	var interval2 model.QueryLogConfigInterval = 2

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
}

func ptr[T any](v T) *T {
	return &v
}
