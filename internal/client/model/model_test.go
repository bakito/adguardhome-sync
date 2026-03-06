package model_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/google/uuid"

	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func ptr[T any](v T) *T {
	return &v
}

func TestFilteringStatus_Unmarshal(t *testing.T) {
	b, err := os.ReadFile("../../../testdata/filtering-status.json")
	if err != nil {
		t.Fatalf("failed to read testdata: %v", err)
	}
	fs := &model.FilterStatus{}
	err = json.Unmarshal(b, fs)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v", err)
	}
}

func TestFilters_Merge(t *testing.T) {
	url := "https://" + uuid.NewString()

	t.Run("should add a missing filter", func(t *testing.T) {
		originFilters := []model.Filter{{Url: url}}
		replicaFilters := []model.Filter{}
		a, u, d := model.MergeFilters(&replicaFilters, &originFilters)
		if len(a) != 1 {
			t.Errorf("expected 1 added filter but got %d", len(a))
		}
		if len(u) != 0 {
			t.Errorf("expected 0 updated filters but got %d", len(u))
		}
		if len(d) != 0 {
			t.Errorf("expected 0 deleted filters but got %d", len(d))
		}
		if a[0].Url != url {
			t.Errorf("expected url %s but got %s", url, a[0].Url)
		}
	})

	t.Run("should remove additional filter", func(t *testing.T) {
		originFilters := []model.Filter{}
		replicaFilters := []model.Filter{{Url: url}}
		_, _, d := model.MergeFilters(&replicaFilters, &originFilters)
		if len(d) != 1 {
			t.Errorf("expected 1 deleted filter but got %d", len(d))
		}
		if d[0].Url != url {
			t.Errorf("expected url %s but got %s", url, d[0].Url)
		}
	})

	t.Run("should update existing filter when enabled differs", func(t *testing.T) {
		enabled := true
		originFilters := []model.Filter{{Url: url, Enabled: enabled}}
		replicaFilters := []model.Filter{{Url: url, Enabled: !enabled}}
		_, u, _ := model.MergeFilters(&replicaFilters, &originFilters)
		if len(u) != 1 {
			t.Errorf("expected 1 updated filter but got %d", len(u))
		}
		if u[0].Enabled != enabled {
			t.Errorf("expected enabled %v but got %v", enabled, u[0].Enabled)
		}
	})

	t.Run("should update existing filter when name differs", func(t *testing.T) {
		name1 := uuid.NewString()
		name2 := uuid.NewString()
		originFilters := []model.Filter{{Url: url, Name: name1}}
		replicaFilters := []model.Filter{{Url: url, Name: name2}}
		_, u, _ := model.MergeFilters(&replicaFilters, &originFilters)
		if len(u) != 1 {
			t.Errorf("expected 1 updated filter but got %d", len(u))
		}
		if u[0].Name != name1 {
			t.Errorf("expected name %s but got %s", name1, u[0].Name)
		}
	})

	t.Run("should have no changes", func(t *testing.T) {
		originFilters := []model.Filter{{Url: url}}
		replicaFilters := []model.Filter{{Url: url}}
		a, u, d := model.MergeFilters(&replicaFilters, &originFilters)
		if len(a) != 0 || len(u) != 0 || len(d) != 0 {
			t.Errorf("expected no changes but got a=%d, u=%d, d=%d", len(a), len(u), len(d))
		}
	})
}

func TestAdGuardInstance_Key(t *testing.T) {
	url := "https://" + uuid.NewString()
	apiPath := "/" + uuid.NewString()
	i := &types.AdGuardInstance{URL: url, APIPath: apiPath}
	expected := url + "#" + apiPath
	if i.Key() != expected {
		t.Errorf("expected %s but got %s", expected, i.Key())
	}
}

func TestRewriteEntry_Key(t *testing.T) {
	domain := uuid.NewString()
	answer := uuid.NewString()
	re := &model.RewriteEntry{Domain: ptr(domain), Answer: ptr(answer)}
	expected := domain + "#" + answer
	if re.Key() != expected {
		t.Errorf("expected %s but got %s", expected, re.Key())
	}
}

func TestQueryLogConfig_Equal(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
		var interval model.QueryLogConfigInterval = 1
		a := &model.QueryLogConfigWithIgnored{
			QueryLogConfig: model.QueryLogConfig{
				Enabled:           ptr(true),
				Interval:          &interval,
				AnonymizeClientIp: ptr(true),
			},
		}
		b := &model.QueryLogConfigWithIgnored{
			QueryLogConfig: model.QueryLogConfig{
				Enabled:           ptr(true),
				Interval:          &interval,
				AnonymizeClientIp: ptr(true),
			},
		}
		if !a.Equals(b) {
			t.Error("expected a to equal b")
		}
	})
	t.Run("should not be equal when enabled differs", func(t *testing.T) {
		a := &model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Enabled: ptr(true)}}
		b := &model.QueryLogConfigWithIgnored{QueryLogConfig: model.QueryLogConfig{Enabled: ptr(false)}}
		if a.Equals(b) {
			t.Error("expected a to not equal b")
		}
	})
}

func TestRewriteEntries_Merge(t *testing.T) {
	domain := uuid.NewString()

	t.Run("should add a missing rewrite entry", func(t *testing.T) {
		originRE := model.RewriteEntries{{Domain: ptr(domain)}}
		replicaRE := model.RewriteEntries{}
		a, _, _, _ := replicaRE.Merge(&originRE)
		if len(a) != 1 {
			t.Errorf("expected 1 added entry but got %d", len(a))
		}
		if *a[0].Domain != domain {
			t.Errorf("expected domain %s but got %s", domain, *a[0].Domain)
		}
	})

	t.Run("should remove additional rewrite entry", func(t *testing.T) {
		originRE := model.RewriteEntries{}
		replicaRE := model.RewriteEntries{{Domain: ptr(domain)}}
		_, r, _, _ := replicaRE.Merge(&originRE)
		if len(r) != 1 {
			t.Errorf("expected 1 removed entry but got %d", len(r))
		}
		if *r[0].Domain != domain {
			t.Errorf("expected domain %s but got %s", domain, *r[0].Domain)
		}
	})

	t.Run("should remove target duplicate", func(t *testing.T) {
		originRE := model.RewriteEntries{{Domain: ptr(domain)}}
		replicaRE := model.RewriteEntries{{Domain: ptr(domain)}, {Domain: ptr(domain)}}
		_, r, _, _ := replicaRE.Merge(&originRE)
		if len(r) != 1 {
			t.Errorf("expected 1 removed entry but got %d", len(r))
		}
	})
}

func TestConfig_UniqueReplicas(t *testing.T) {
	url := "https://" + uuid.NewString()
	apiPath := "/" + uuid.NewString()

	t.Run("should return only one replica if same url and apiPath", func(t *testing.T) {
		cfg := &types.Config{
			Replica:  &types.AdGuardInstance{URL: url, APIPath: apiPath},
			Replicas: []types.AdGuardInstance{{URL: url, APIPath: apiPath}, {URL: url, APIPath: apiPath}},
		}
		r := cfg.UniqueReplicas()
		if len(r) != 1 {
			t.Errorf("expected 1 unique replica but got %d", len(r))
		}
	})
}

func TestClients_Merge(t *testing.T) {
	name := uuid.NewString()

	t.Run("should add a missing client", func(t *testing.T) {
		originClients := &model.Clients{}
		originClients.Add(model.Client{Name: ptr(name)})
		replicaClients := model.Clients{}
		a, _, _ := replicaClients.Merge(originClients)
		if len(a) != 1 {
			t.Errorf("expected 1 added client but got %d", len(a))
		}
		if *a[0].Name != name {
			t.Errorf("expected name %s but got %s", name, *a[0].Name)
		}
	})
}

func TestClient_Equals(t *testing.T) {
	t.Run("should equal if only timezone differs on empty blocked service schedule", func(t *testing.T) {
		cl1 := &model.Client{
			Name:                    ptr("foo"),
			BlockedServicesSchedule: &model.Schedule{TimeZone: ptr("UTC")},
		}
		cl2 := &model.Client{
			Name:                    ptr("foo"),
			BlockedServicesSchedule: &model.Schedule{TimeZone: ptr("Local")},
		}
		if !cl1.Equals(cl2) {
			t.Error("expected cl1 to equal cl2")
		}
	})
}

func TestBlockedServices_Equals(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
		s1 := &model.BlockedServicesArray{"a", "b"}
		s2 := &model.BlockedServicesArray{"b", "a"}
		if !model.EqualsStringSlice(s1, s2, true) {
			t.Error("expected s1 to equal s2")
		}
	})
}

func TestDNSConfig_Equal(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
		dc1 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"a"})}
		dc2 := &model.DNSConfig{LocalPtrUpstreams: ptr([]string{"a"})}
		if !dc1.Equals(dc2) {
			t.Error("expected dc1 to equal dc2")
		}
	})
}

func TestDhcpStatus_Equals(t *testing.T) {
	t.Run("should be equal", func(t *testing.T) {
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
			t.Error("expected dc1 to equal dc2")
		}
	})
}

func TestDhcpStatus_HasConfig(t *testing.T) {
	t.Run("should not have a config", func(t *testing.T) {
		dc1 := &model.DhcpStatus{
			V4: &model.DhcpConfigV4{},
			V6: &model.DhcpConfigV6{},
		}
		if dc1.HasConfig() {
			t.Error("expected HasConfig to be false")
		}
	})
}
