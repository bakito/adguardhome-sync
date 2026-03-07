package client_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/client/model"
	"github.com/bakito/adguardhome-sync/internal/types"
)

var (
	username = uuid.NewString()
	password = uuid.NewString()
)

func TestClient_Host(t *testing.T) {
	inst := types.AdGuardInstance{URL: "https://foo.bar:3000"}
	err := inst.Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	cl, _ := client.New(inst, 0)
	host := cl.Host()
	if host != "foo.bar:3000" {
		t.Errorf("Host() = %s, want %s", host, "foo.bar:3000")
	}
}

func TestClient_Filter(t *testing.T) {
	t.Run("should read filter status", func(t *testing.T) {
		ts, cl := ClientGet(t, "filtering-status.json", "/filtering/status")
		defer ts.Close()
		fs, err := cl.Filtering()
		if err != nil {
			t.Fatalf("Filtering() error = %v", err)
		}
		if fs.Enabled == nil || !*fs.Enabled {
			t.Error("expected filtering to be enabled")
		}
		if fs.Filters == nil || len(*fs.Filters) != 2 {
			t.Errorf("len(filters) = %d, want 2", len(*fs.Filters))
		}
	})
	t.Run("should enable protection", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/config", `{"enabled":true,"interval":123}`)
		defer ts.Close()
		err := cl.ToggleFiltering(true, 123)
		if err != nil {
			t.Errorf("ToggleFiltering() error = %v", err)
		}
	})
	t.Run("should disable protection", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/config", `{"enabled":false,"interval":123}`)
		defer ts.Close()
		err := cl.ToggleFiltering(false, 123)
		if err != nil {
			t.Errorf("ToggleFiltering() error = %v", err)
		}
	})
	t.Run("should call RefreshFilters", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/refresh", `{"whitelist":true}`)
		defer ts.Close()
		err := cl.RefreshFilters(true)
		if err != nil {
			t.Errorf("RefreshFilters() error = %v", err)
		}
	})
	t.Run("should add Filters", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/add_url",
			`{"name":"","url":"foo","whitelist":true}`,
			`{"name":"","url":"bar","whitelist":true}`,
		)
		defer ts.Close()
		err := cl.AddFilter(true, model.Filter{Url: "foo"})
		if err != nil {
			t.Errorf("AddFilter(foo) error = %v", err)
		}
		err = cl.AddFilter(true, model.Filter{Url: "bar"})
		if err != nil {
			t.Errorf("AddFilter(bar) error = %v", err)
		}
	})
	t.Run("should update Filters", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/set_url",
			`{"data":{"enabled":false,"name":"","url":"foo"},"url":"foo","whitelist":true}`,
			`{"data":{"enabled":false,"name":"","url":"bar"},"url":"bar","whitelist":true}`,
		)
		defer ts.Close()
		err := cl.UpdateFilter(true, model.Filter{Url: "foo"})
		if err != nil {
			t.Errorf("UpdateFilter(foo) error = %v", err)
		}
		err = cl.UpdateFilter(true, model.Filter{Url: "bar"})
		if err != nil {
			t.Errorf("UpdateFilter(bar) error = %v", err)
		}
	})
	t.Run("should delete Filters", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/remove_url",
			`{"url":"foo","whitelist":true}`,
			`{"url":"bar","whitelist":true}`,
		)
		defer ts.Close()
		err := cl.DeleteFilter(true, model.Filter{Url: "foo"})
		if err != nil {
			t.Errorf("DeleteFilter(foo) error = %v", err)
		}
		err = cl.DeleteFilter(true, model.Filter{Url: "bar"})
		if err != nil {
			t.Errorf("DeleteFilter(bar) error = %v", err)
		}
	})
	t.Run("should set empty filter rules", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/set_rules",
			`{"rules":[]}`,
		)
		defer ts.Close()
		err := cl.SetCustomRules(new([]string{}))
		if err != nil {
			t.Errorf("SetCustomRules() error = %v", err)
		}
	})
	t.Run("should set nil filter rules", func(t *testing.T) {
		ts, cl := ClientPost(t, "/filtering/set_rules",
			`{}`,
		)
		defer ts.Close()
		err := cl.SetCustomRules(nil)
		if err != nil {
			t.Errorf("SetCustomRules() error = %v", err)
		}
	})
}

func TestClient_Status(t *testing.T) {
	t.Run("should read status", func(t *testing.T) {
		ts, cl := ClientGet(t, "status.json", "/status")
		defer ts.Close()
		fs, err := cl.Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if len(fs.DnsAddresses) != 1 {
			t.Errorf("len(DnsAddresses) = %d, want 1", len(fs.DnsAddresses))
		}
		if fs.DnsAddresses[0] != "192.168.1.2" {
			t.Errorf("DnsAddresses[0] = %s, want %s", fs.DnsAddresses[0], "192.168.1.2")
		}
		if fs.Version != "v0.105.2" {
			t.Errorf("Version = %s, want %s", fs.Version, "v0.105.2")
		}
	})
	t.Run("should return ErrSetupNeeded", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Location", "/install.html")
			w.WriteHeader(http.StatusFound)
		}))
		defer ts.Close()
		cl, err := client.New(types.AdGuardInstance{URL: ts.URL}, 0)
		if err != nil {
			t.Fatalf("client.New error = %v", err)
		}
		_, err = cl.Status()
		if !errors.Is(err, client.ErrSetupNeeded) {
			t.Errorf("error = %v, want %v", err, client.ErrSetupNeeded)
		}
	})
}

func TestClient_Setup(t *testing.T) {
	ts, cl := ClientPost(t,
		"/install/configure",
		fmt.Sprintf(
			`{"web":{"ip":"0.0.0.0","port":3000,"status":"","can_autofix":false},"dns":{"ip":"0.0.0.0","port":53,"status":"","can_autofix":false},"username":%q,"password":%q}`,
			username,
			password,
		),
	)
	defer ts.Close()
	err := cl.Setup()
	if err != nil {
		t.Errorf("Setup() error = %v", err)
	}
}

func TestClient_RewriteList(t *testing.T) {
	t.Run("should read RewriteList", func(t *testing.T) {
		ts, cl := ClientGet(t, "rewrite-list.json", "/rewrite/list")
		defer ts.Close()
		rwl, err := cl.RewriteList()
		if err != nil {
			t.Fatalf("RewriteList() error = %v", err)
		}
		if len(*rwl) != 2 {
			t.Errorf("len(rwl) = %d, want 2", len(*rwl))
		}
	})
	t.Run("should add RewriteList", func(t *testing.T) {
		ts, cl := ClientPost(t, "/rewrite/add", `{"answer":"foo","domain":"foo"}`, `{"answer":"bar","domain":"bar"}`)
		defer ts.Close()
		err := cl.AddRewriteEntries(
			model.RewriteEntry{Answer: new("foo"), Domain: new("foo")},
			model.RewriteEntry{Answer: new("bar"), Domain: new("bar")},
		)
		if err != nil {
			t.Errorf("AddRewriteEntries() error = %v", err)
		}
	})
	t.Run("should delete RewriteList", func(t *testing.T) {
		ts, cl := ClientPost(t, "/rewrite/delete", `{"answer":"foo","domain":"foo"}`, `{"answer":"bar","domain":"bar"}`)
		defer ts.Close()
		err := cl.DeleteRewriteEntries(
			model.RewriteEntry{Answer: new("foo"), Domain: new("foo")},
			model.RewriteEntry{Answer: new("bar"), Domain: new("bar")},
		)
		if err != nil {
			t.Errorf("DeleteRewriteEntries() error = %v", err)
		}
	})
}

func TestClient_SafeBrowsing(t *testing.T) {
	t.Run("should read safebrowsing status", func(t *testing.T) {
		ts, cl := ClientGet(t, "safebrowsing-status.json", "/safebrowsing/status")
		defer ts.Close()
		sb, err := cl.SafeBrowsing()
		if err != nil {
			t.Fatalf("SafeBrowsing() error = %v", err)
		}
		if !sb {
			t.Error("expected safebrowsing to be true")
		}
	})
	t.Run("should enable safebrowsing", func(t *testing.T) {
		ts, cl := ClientPost(t, "/safebrowsing/enable", "")
		defer ts.Close()
		err := cl.ToggleSafeBrowsing(true)
		if err != nil {
			t.Errorf("ToggleSafeBrowsing(true) error = %v", err)
		}
	})
	t.Run("should disable safebrowsing", func(t *testing.T) {
		ts, cl := ClientPost(t, "/safebrowsing/disable", "")
		defer ts.Close()
		err := cl.ToggleSafeBrowsing(false)
		if err != nil {
			t.Errorf("ToggleSafeBrowsing(false) error = %v", err)
		}
	})
}

func TestClient_SafeSearchConfig(t *testing.T) {
	t.Run("should read safesearch status", func(t *testing.T) {
		ts, cl := ClientGet(t, "safesearch-status.json", "/safesearch/status")
		defer ts.Close()
		ss, err := cl.SafeSearchConfig()
		if err != nil {
			t.Fatalf("SafeSearchConfig() error = %v", err)
		}
		if ss.Enabled == nil || !*ss.Enabled {
			t.Error("expected safesearch to be enabled")
		}
	})
	t.Run("should enable safesearch", func(t *testing.T) {
		ts, cl := ClientPut(t, "/safesearch/settings", `{"enabled":true}`)
		defer ts.Close()
		err := cl.SetSafeSearchConfig(&model.SafeSearchConfig{Enabled: new(true)})
		if err != nil {
			t.Errorf("SetSafeSearchConfig(true) error = %v", err)
		}
	})
	t.Run("should disable safesearch", func(t *testing.T) {
		ts, cl := ClientPut(t, "/safesearch/settings", `{"enabled":false}`)
		defer ts.Close()
		err := cl.SetSafeSearchConfig(&model.SafeSearchConfig{Enabled: new(false)})
		if err != nil {
			t.Errorf("SetSafeSearchConfig(false) error = %v", err)
		}
	})
}

func TestClient_Parental(t *testing.T) {
	t.Run("should read parental status", func(t *testing.T) {
		ts, cl := ClientGet(t, "parental-status.json", "/parental/status")
		defer ts.Close()
		p, err := cl.Parental()
		if err != nil {
			t.Fatalf("Parental() error = %v", err)
		}
		if !p {
			t.Error("expected parental to be true")
		}
	})
	t.Run("should enable parental", func(t *testing.T) {
		ts, cl := ClientPost(t, "/parental/enable", "")
		defer ts.Close()
		err := cl.ToggleParental(true)
		if err != nil {
			t.Errorf("ToggleParental(true) error = %v", err)
		}
	})
	t.Run("should disable parental", func(t *testing.T) {
		ts, cl := ClientPost(t, "/parental/disable", "")
		defer ts.Close()
		err := cl.ToggleParental(false)
		if err != nil {
			t.Errorf("ToggleParental(false) error = %v", err)
		}
	})
}

func TestClient_Protection(t *testing.T) {
	t.Run("should enable protection", func(t *testing.T) {
		ts, cl := ClientPost(t, "/dns_config", `{"protection_enabled":true}`)
		defer ts.Close()
		err := cl.ToggleProtection(true)
		if err != nil {
			t.Errorf("ToggleProtection(true) error = %v", err)
		}
	})
	t.Run("should disable protection", func(t *testing.T) {
		ts, cl := ClientPost(t, "/dns_config", `{"protection_enabled":false}`)
		defer ts.Close()
		err := cl.ToggleProtection(false)
		if err != nil {
			t.Errorf("ToggleProtection(false) error = %v", err)
		}
	})
}

func TestClient_BlockedServicesSchedule(t *testing.T) {
	t.Run("should read BlockedServicesSchedule", func(t *testing.T) {
		ts, cl := ClientGet(t, "blockedservicesschedule-get.json", "/blocked_services/get")
		defer ts.Close()
		s, err := cl.BlockedServicesSchedule()
		if err != nil {
			t.Fatalf("BlockedServicesSchedule() error = %v", err)
		}
		if s.Ids == nil || len(*s.Ids) != 3 {
			t.Errorf("len(Ids) = %d, want 3", len(*s.Ids))
		}
	})
	t.Run("should set BlockedServicesSchedule", func(t *testing.T) {
		ts, cl := ClientPost(t, "/blocked_services/update",
			`{"ids":["bar","foo"],"schedule":{"mon":{"end":99,"start":1}}}`)
		defer ts.Close()
		err := cl.SetBlockedServicesSchedule(&model.BlockedServicesSchedule{
			Ids: new([]string{"foo", "bar"}),
			Schedule: &model.Schedule{
				Mon: &model.DayRange{
					Start: new(float32(1.0)),
					End:   new(float32(99.0)),
				},
			},
		})
		if err != nil {
			t.Errorf("SetBlockedServicesSchedule() error = %v", err)
		}
	})
}

func TestClient_Clients(t *testing.T) {
	t.Run("should read Clients", func(t *testing.T) {
		ts, cl := ClientGet(t, "clients.json", "/clients")
		defer ts.Close()
		c, err := cl.Clients()
		if err != nil {
			t.Fatalf("Clients() error = %v", err)
		}
		if c.Clients == nil || len(*c.Clients) != 2 {
			t.Errorf("len(Clients) = %d, want 2", len(*c.Clients))
		}
	})
	t.Run("should add Clients", func(t *testing.T) {
		ts, cl := ClientPost(t, "/clients/add",
			`{"ids":["id"],"name":"foo"}`,
		)
		defer ts.Close()
		err := cl.AddClient(&model.Client{Name: new("foo"), Ids: new([]string{"id"})})
		if err != nil {
			t.Errorf("AddClient() error = %v", err)
		}
	})
	t.Run("should update Clients", func(t *testing.T) {
		ts, cl := ClientPost(t, "/clients/update",
			`{"data":{"ids":["id"],"name":"foo"},"name":"foo"}`,
		)
		defer ts.Close()
		err := cl.UpdateClient(&model.Client{Name: new("foo"), Ids: new([]string{"id"})})
		if err != nil {
			t.Errorf("UpdateClient() error = %v", err)
		}
	})
	t.Run("should delete Clients", func(t *testing.T) {
		ts, cl := ClientPost(t, "/clients/delete",
			`{"ids":["id"],"name":"foo"}`,
		)
		defer ts.Close()
		err := cl.DeleteClient(&model.Client{Name: new("foo"), Ids: new([]string{"id"})})
		if err != nil {
			t.Errorf("DeleteClient() error = %v", err)
		}
	})
}

func TestClient_QueryLogConfig(t *testing.T) {
	t.Run("should read QueryLogConfig", func(t *testing.T) {
		ts, cl := ClientGet(t, "querylog_config.json", "/querylog/config")
		defer ts.Close()
		qlc, err := cl.QueryLogConfig()
		if err != nil {
			t.Fatalf("QueryLogConfig() error = %v", err)
		}
		if qlc.Enabled == nil || !*qlc.Enabled {
			t.Error("expected querylog to be enabled")
		}
		if qlc.Interval == nil || *qlc.Interval != model.QueryLogConfigInterval(90) {
			t.Errorf("Interval = %v, want 90", qlc.Interval)
		}
	})
	t.Run("should set QueryLogConfig", func(t *testing.T) {
		ts, cl := ClientPut(t,
			"/querylog/config/update",
			`{"anonymize_client_ip":true,"enabled":true,"interval":123,"ignored":["foo.bar"]}`,
		)
		defer ts.Close()

		var interval model.QueryLogConfigInterval = 123
		err := cl.SetQueryLogConfig(&model.QueryLogConfigWithIgnored{
			QueryLogConfig: model.QueryLogConfig{
				AnonymizeClientIp: new(true),
				Interval:          &interval,
				Enabled:           new(true),
			},
			Ignored: []string{"foo.bar"},
		})
		if err != nil {
			t.Errorf("SetQueryLogConfig() error = %v", err)
		}
	})
}

func TestClient_StatsConfig(t *testing.T) {
	t.Run("should read StatsConfig", func(t *testing.T) {
		ts, cl := ClientGet(t, "stats_info.json", "/stats/config")
		defer ts.Close()
		sc, err := cl.StatsConfig()
		if err != nil {
			t.Fatalf("StatsConfig() error = %v", err)
		}
		if sc.Interval != float32(1) {
			t.Errorf("Interval = %v, want 1", sc.Interval)
		}
	})
	t.Run("should set StatsConfig", func(t *testing.T) {
		ts, cl := ClientPost(t, "/stats/config/update", `{"enabled":false,"ignored":null,"interval":123}`)
		defer ts.Close()

		var interval float32 = 123
		err := cl.SetStatsConfig(&model.PutStatsConfigUpdateRequest{Interval: interval})
		if err != nil {
			t.Errorf("SetStatsConfig() error = %v", err)
		}
	})
}

func TestClient_HelperFunctions(t *testing.T) {
	t.Run("doGet", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()
		cl, err := client.New(types.AdGuardInstance{URL: ts.URL}, 0)
		if err != nil {
			t.Fatalf("client.New error = %v", err)
		}
		_, err = cl.Status()
		if err == nil || err.Error() != "401 Unauthorized" {
			t.Errorf("error = %v, want 401 Unauthorized", err)
		}
	})

	t.Run("doPost", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer ts.Close()
		cl, err := client.New(types.AdGuardInstance{URL: ts.URL}, 0)
		if err != nil {
			t.Fatalf("client.New error = %v", err)
		}
		var interval float32 = 123
		err = cl.SetStatsConfig(&model.PutStatsConfigUpdateRequest{Interval: interval})
		if err == nil || err.Error() != "401 Unauthorized" {
			t.Errorf("error = %v, want 401 Unauthorized", err)
		}
	})
}

func ClientGet(t *testing.T, file, path string) (*httptest.Server, client.Client) {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != types.DefaultAPIPath+path {
			t.Errorf("Path = %s, want %s", r.URL.Path, types.DefaultAPIPath+path)
		}
		b, err := os.ReadFile(filepath.Join("..", "..", "testdata", file))
		if err != nil {
			t.Errorf("ReadFile error = %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(b)
	}))
	cl, err := client.New(types.AdGuardInstance{URL: ts.URL}, 0)
	if err != nil {
		t.Fatalf("client.New error = %v", err)
	}
	return ts, cl
}

func ClientPost(t *testing.T, path string, content ...string) (*httptest.Server, client.Client) {
	t.Helper()
	var index int
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.URL.Path != types.DefaultAPIPath+path {
			t.Errorf("Path = %s, want %s", r.URL.Path, types.DefaultAPIPath+path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != content[index] {
			t.Errorf("Body = %s, want %s", string(body), content[index])
		}
		index++
	}))

	cl, err := client.New(types.AdGuardInstance{URL: ts.URL, Username: username, Password: password}, 0)
	if err != nil {
		t.Fatalf("client.New error = %v", err)
	}
	return ts, cl
}

func ClientPut(t *testing.T, path string, content ...string) (*httptest.Server, client.Client) {
	t.Helper()
	var index int
	ts := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		if r.URL.Path != types.DefaultAPIPath+path {
			t.Errorf("Path = %s, want %s", r.URL.Path, types.DefaultAPIPath+path)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != content[index] {
			t.Errorf("Body = %s, want %s", string(body), content[index])
		}
		index++
	}))

	cl, err := client.New(types.AdGuardInstance{URL: ts.URL, Username: username, Password: password}, 0)
	if err != nil {
		t.Fatalf("client.New error = %v", err)
	}
	return ts, cl
}
