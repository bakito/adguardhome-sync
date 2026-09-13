package types

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestAdGuardInstance_Init(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		webURL      string
		wantHost    string
		wantWebHost string
		wantURL     string
		wantWebURL  string
		wantErr     bool
	}{
		{
			name:        "should correctly set Host and WebHost if only URL is set",
			url:         "https://localhost:3000",
			wantHost:    "localhost:3000",
			wantWebHost: "localhost:3000",
			wantURL:     "https://localhost:3000",
			wantWebURL:  "https://localhost:3000",
		},
		{
			name:        "should correctly set Host and WebHost if URL and WebURL are set",
			url:         "https://localhost:3000",
			webURL:      "https://127.0.0.1:4000",
			wantHost:    "localhost:3000",
			wantWebHost: "127.0.0.1:4000",
			wantURL:     "https://localhost:3000",
			wantWebURL:  "https://127.0.0.1:4000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := AdGuardInstance{
				URL:    tt.url,
				WebURL: tt.webURL,
			}
			if err := inst.Init(); (err != nil) != tt.wantErr {
				t.Errorf("AdGuardInstance.Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if inst.Host != tt.wantHost {
				t.Errorf("AdGuardInstance.Host = %v, want %v", inst.Host, tt.wantHost)
			}
			if inst.WebHost != tt.wantWebHost {
				t.Errorf("AdGuardInstance.WebHost = %v, want %v", inst.WebHost, tt.wantWebHost)
			}
			if inst.URL != tt.wantURL {
				t.Errorf("AdGuardInstance.URL = %v, want %v", inst.URL, tt.wantURL)
			}
			if inst.WebURL != tt.wantWebURL {
				t.Errorf("AdGuardInstance.WebURL = %v, want %v", inst.WebURL, tt.wantWebURL)
			}
		})
	}
}

func TestConfig_Init(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []Replica{
			{URL: "https://localhost:3000"},
		},
	}
	err := cfg.Init()
	if err != nil {
		t.Fatalf("Config.Init() error = %v", err)
	}
	if cfg.Replicas[0].Host != "localhost:3000" {
		t.Errorf("cfg.Replicas[0].Host = %v, want localhost:3000", cfg.Replicas[0].Host)
	}
	if cfg.Replicas[0].WebHost != "localhost:3000" {
		t.Errorf("cfg.Replicas[0].WebHost = %v, want localhost:3000", cfg.Replicas[0].WebHost)
	}
	if cfg.Replicas[0].URL != "https://localhost:3000" {
		t.Errorf("cfg.Replicas[0].URL = %v, want https://localhost:3000", cfg.Replicas[0].URL)
	}
	if cfg.Replicas[0].WebURL != "https://localhost:3000" {
		t.Errorf("cfg.Replicas[0].WebURL = %v, want https://localhost:3000", cfg.Replicas[0].WebURL)
	}
}

func TestConfig_UniqueReplicas(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []Replica{
			{URL: "a"},
			{URL: "a", APIPath: DefaultAPIPath},
			{URL: "a", APIPath: "foo"},
			{URL: "b", APIPath: DefaultAPIPath},
		},
		Replica: &Replica{URL: "b"},
	}
	replicas := cfg.UniqueReplicas()
	if len(replicas) != 3 {
		t.Errorf("len(cfg.UniqueReplicas()) = %v, want 3", len(replicas))
	}
}

func TestConfig_mask(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []Replica{
			{URL: "a", Username: "user", Password: "pass"},
		},
		Replica: &Replica{URL: "a", Username: "user", Password: "pass"},
		API:     API{Username: "user", Password: "pass"},
	}
	masked := cfg.mask()
	if masked.Replicas[0].Username != "u**r" {
		t.Errorf("masked.Replicas[0].Username = %v, want u**r", masked.Replicas[0].Username)
	}
	if masked.Replicas[0].Password != "p**s" {
		t.Errorf("masked.Replicas[0].Password = %v, want p**s", masked.Replicas[0].Password)
	}
	if masked.Replica.Username != "u**r" {
		t.Errorf("masked.Replica.Username = %v, want u**r", masked.Replica.Username)
	}
	if masked.Replica.Password != "p**s" {
		t.Errorf("masked.Replica.Password = %v, want p**s", masked.Replica.Password)
	}
	if masked.API.Username != "u**r" {
		t.Errorf("masked.API.Username = %v, want u**r", masked.API.Username)
	}
	if masked.API.Password != "p**s" {
		t.Errorf("masked.API.Password = %v, want p**s", masked.API.Password)
	}
}

func Test_mask(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{name: "Empty password", value: "", expected: ""},
		{name: "1 char password", value: "a", expected: "*"},
		{name: "2 char password", value: "ab", expected: "**"},
		{name: "3 char password", value: "abc", expected: "a*c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mask(tt.value); got != tt.expected {
				t.Errorf("mask() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFeature_collectDisabled(t *testing.T) {
	tests := []struct {
		name string
		all  bool
		want int
	}{
		{name: "should log all features", all: false, want: 14},
		{name: "should log no features", all: true, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := NewFeatures(tt.all)
			if got := f.collectDisabled(); len(got) != tt.want {
				t.Errorf("Features.collectDisabled() = %v, want %v", len(got), tt.want)
			}
		})
	}
}

func TestTLS_Enabled(t *testing.T) {
	tests := []struct {
		name    string
		certDir string
		want    bool
	}{
		{name: "should use enabled", certDir: "/path/to/certs", want: true},
		{name: "should use disabled", certDir: " ", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tls := TLS{CertDir: tt.certDir}
			if got := tls.Enabled(); got != tt.want {
				t.Errorf("TLS.Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTLS_Certs(t *testing.T) {
	tests := []struct {
		name     string
		certDir  string
		certName string
		keyName  string
		wantCrt  string
		wantKey  string
	}{
		{
			name:    "should use default crt and key",
			certDir: "/path/to/certs",
			wantCrt: "/path/to/certs/tls.crt",
			wantKey: "/path/to/certs/tls.key",
		},
		{
			name:     "should use custom crt and key",
			certDir:  "/path/to/certs",
			certName: "foo.crt",
			keyName:  "bar.key",
			wantCrt:  "/path/to/certs/foo.crt",
			wantKey:  "/path/to/certs/bar.key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tls := TLS{
				CertDir:  tt.certDir,
				CertName: tt.certName,
				KeyName:  tt.keyName,
			}
			crt, key := tls.Certs()
			if normalizePath(crt) != tt.wantCrt {
				t.Errorf("TLS.Certs() crt = %v, want %v", normalizePath(crt), tt.wantCrt)
			}
			if normalizePath(key) != tt.wantKey {
				t.Errorf("TLS.Certs() key = %v, want %v", normalizePath(key), tt.wantKey)
			}
		})
	}
}

func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

func TestReplica_EffectiveFeatures(t *testing.T) {
	global := NewFeatures(true)
	global.GeneralSettings = false

	t.Run("should return global features when Features is nil", func(t *testing.T) {
		inst := Replica{}
		got := inst.EffectiveFeatures(global)
		if got != global {
			t.Errorf("EffectiveFeatures() = %+v, want %+v", got, global)
		}
	})

	t.Run("should return replica features when Features is set", func(t *testing.T) {
		replicaFeatures := NewFeatures(false)
		replicaFeatures.ClientSettings = true
		inst := Replica{Features: &replicaFeatures}
		got := inst.EffectiveFeatures(global)
		if got != replicaFeatures {
			t.Errorf("EffectiveFeatures() = %+v, want %+v", got, replicaFeatures)
		}
	})
}

func TestConfig_HasDHCPFeature(t *testing.T) {
	t.Run("should return true when global DHCP serverConfig is true", func(t *testing.T) {
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1"}},
		}
		cfg.Features.DHCP.ServerConfig = true
		if !cfg.HasDHCPFeature() {
			t.Error("HasDHCPFeature() = false, want true")
		}
	})

	t.Run("should return true when replica DHCP staticLeases is true", func(t *testing.T) {
		replicaFeat := NewFeatures(false)
		replicaFeat.DHCP.StaticLeases = true
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1", Features: &replicaFeat}},
		}
		if !cfg.HasDHCPFeature() {
			t.Error("HasDHCPFeature() = false, want true")
		}
	})

	t.Run("should return false when both global and replicas have DHCP disabled", func(t *testing.T) {
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1"}},
		}
		if cfg.HasDHCPFeature() {
			t.Error("HasDHCPFeature() = true, want false")
		}
	})
}

func TestConfig_HasTLSFeature(t *testing.T) {
	t.Run("should return true when global TLSConfig is true", func(t *testing.T) {
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1"}},
		}
		cfg.Features.TLSConfig = true
		if !cfg.HasTLSFeature() {
			t.Error("HasTLSFeature() = false, want true")
		}
	})

	t.Run("should return true when replica TLSConfig is true", func(t *testing.T) {
		replicaFeat := NewFeatures(false)
		replicaFeat.TLSConfig = true
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1", Features: &replicaFeat}},
		}
		if !cfg.HasTLSFeature() {
			t.Error("HasTLSFeature() = false, want true")
		}
	})

	t.Run("should return false when both global and replicas have TLS disabled", func(t *testing.T) {
		cfg := Config{
			Features: NewFeatures(false),
			Replicas: []Replica{{URL: "http://replica1"}},
		}
		if cfg.HasTLSFeature() {
			t.Error("HasTLSFeature() = true, want false")
		}
	})
}

func TestFeatures_UnmarshalYAML(t *testing.T) {
	yamlData := `
dns:
  serverConfig: false
clientSettings: false
`
	var f Features
	if err := yaml.Unmarshal([]byte(yamlData), &f); err != nil {
		t.Fatalf("yaml.Unmarshal error = %v", err)
	}

	if f.DNS.ServerConfig {
		t.Error("DNS.ServerConfig = true, want false")
	}
	if f.ClientSettings {
		t.Error("ClientSettings = true, want false")
	}
	// Unspecified fields should default to true (from NewFeatures(true))
	if !f.GeneralSettings {
		t.Error("GeneralSettings = false, want true")
	}
	if !f.DNS.Rewrites {
		t.Error("DNS.Rewrites = false, want true")
	}
	if !f.Filters.Blacklist {
		t.Error("Filters.Blacklist = false, want true")
	}
	if f.TLSConfig {
		t.Error("TLSConfig = true, want false")
	}
}

func TestFeatures_UnmarshalJSON(t *testing.T) {
	jsonData := `{"dns": {"serverConfig": false}, "clientSettings": false}`
	var f Features
	if err := json.Unmarshal([]byte(jsonData), &f); err != nil {
		t.Fatalf("json.Unmarshal error = %v", err)
	}

	if f.DNS.ServerConfig {
		t.Error("DNS.ServerConfig = true, want false")
	}
	if f.ClientSettings {
		t.Error("ClientSettings = true, want false")
	}
	// Unspecified fields should default to true
	if !f.GeneralSettings {
		t.Error("GeneralSettings = false, want true")
	}
	if !f.DNS.Rewrites {
		t.Error("DNS.Rewrites = false, want true")
	}
	if !f.Filters.Blacklist {
		t.Error("Filters.Blacklist = false, want true")
	}
	if f.TLSConfig {
		t.Error("TLSConfig = true, want false")
	}
}
