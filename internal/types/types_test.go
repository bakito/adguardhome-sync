package types

import (
	"strings"
	"testing"
)

func TestAdGuardInstance_Init(t *testing.T) {
	t.Run("should correctly set Host and WebHost if only URL is set", func(t *testing.T) {
		inst := AdGuardInstance{URL: "https://localhost:3000"}
		err := inst.Init()
		if err != nil {
			t.Errorf("Init() error = %v", err)
		}
		if inst.Host != "localhost:3000" {
			t.Errorf("inst.Host = %v, want %v", inst.Host, "localhost:3000")
		}
		if inst.WebHost != "localhost:3000" {
			t.Errorf("inst.WebHost = %v, want %v", inst.WebHost, "localhost:3000")
		}
		if inst.URL != "https://localhost:3000" {
			t.Errorf("inst.URL = %v, want %v", inst.URL, "https://localhost:3000")
		}
		if inst.WebURL != "https://localhost:3000" {
			t.Errorf("inst.WebURL = %v, want %v", inst.WebURL, "https://localhost:3000")
		}
	})

	t.Run("should correctly set Host and WebHost if URL and WebURL are set", func(t *testing.T) {
		inst := AdGuardInstance{
			URL:    "https://localhost:3000",
			WebURL: "https://127.0.0.1:4000",
		}
		err := inst.Init()
		if err != nil {
			t.Errorf("Init() error = %v", err)
		}
		if inst.Host != "localhost:3000" {
			t.Errorf("inst.Host = %v, want %v", inst.Host, "localhost:3000")
		}
		if inst.WebHost != "127.0.0.1:4000" {
			t.Errorf("inst.WebHost = %v, want %v", inst.WebHost, "127.0.0.1:4000")
		}
		if inst.URL != "https://localhost:3000" {
			t.Errorf("inst.URL = %v, want %v", inst.URL, "https://localhost:3000")
		}
		if inst.WebURL != "https://127.0.0.1:4000" {
			t.Errorf("inst.WebURL = %v, want %v", inst.WebURL, "https://127.0.0.1:4000")
		}
	})
}

func TestConfig_Init(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []AdGuardInstance{
			{URL: "https://localhost:3000"},
		},
	}
	err := cfg.Init()
	if err != nil {
		t.Errorf("Init() error = %v", err)
	}
	if cfg.Replicas[0].Host != "localhost:3000" {
		t.Errorf("cfg.Replicas[0].Host = %v, want %v", cfg.Replicas[0].Host, "localhost:3000")
	}
	if cfg.Replicas[0].WebHost != "localhost:3000" {
		t.Errorf("cfg.Replicas[0].WebHost = %v, want %v", cfg.Replicas[0].WebHost, "localhost:3000")
	}
	if cfg.Replicas[0].URL != "https://localhost:3000" {
		t.Errorf("cfg.Replicas[0].URL = %v, want %v", cfg.Replicas[0].URL, "https://localhost:3000")
	}
	if cfg.Replicas[0].WebURL != "https://localhost:3000" {
		t.Errorf("cfg.Replicas[0].WebURL = %v, want %v", cfg.Replicas[0].WebURL, "https://localhost:3000")
	}
}

func TestConfig_UniqueReplicas(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []AdGuardInstance{
			{URL: "a"},
			{URL: "a", APIPath: DefaultAPIPath},
			{URL: "a", APIPath: "foo"},
			{URL: "b", APIPath: DefaultAPIPath},
		},
		Replica: &AdGuardInstance{URL: "b"},
	}
	replicas := cfg.UniqueReplicas()
	if len(replicas) != 3 {
		t.Errorf("len(replicas) = %v, want 3", len(replicas))
	}
}

func TestConfig_Mask(t *testing.T) {
	cfg := Config{
		Origin: &AdGuardInstance{},
		Replicas: []AdGuardInstance{
			{URL: "a", Username: "user", Password: "pass"},
		},
		Replica: &AdGuardInstance{URL: "a", Username: "user", Password: "pass"},
		API:     API{Username: "user", Password: "pass"},
	}
	masked := cfg.mask()
	if masked.Replicas[0].Username != "u**r" {
		t.Errorf("masked.Replicas[0].Username = %v, want %v", masked.Replicas[0].Username, "u**r")
	}
	if masked.Replicas[0].Password != "p**s" {
		t.Errorf("masked.Replicas[0].Password = %v, want %v", masked.Replicas[0].Password, "p**s")
	}
	if masked.Replica.Username != "u**r" {
		t.Errorf("masked.Replica.Username = %v, want %v", masked.Replica.Username, "u**r")
	}
	if masked.Replica.Password != "p**s" {
		t.Errorf("masked.Replica.Password = %v, want %v", masked.Replica.Password, "p**s")
	}
	if masked.API.Username != "u**r" {
		t.Errorf("masked.API.Username = %v, want %v", masked.API.Username, "u**r")
	}
	if masked.API.Password != "p**s" {
		t.Errorf("masked.API.Password = %v, want %v", masked.API.Password, "p**s")
	}
}

func TestMask(t *testing.T) {
	tests := []struct {
		value    string
		expected string
	}{
		{"", ""},
		{"a", "*"},
		{"ab", "**"},
		{"abc", "a*c"},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			if got := mask(tt.value); got != tt.expected {
				t.Errorf("mask() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFeature_LogDisabled(t *testing.T) {
	t.Run("should log all features", func(t *testing.T) {
		f := NewFeatures(false)
		if len(f.collectDisabled()) != 12 {
			t.Errorf("len(f.collectDisabled()) = %v, want 12", len(f.collectDisabled()))
		}
	})
	t.Run("should log no features", func(t *testing.T) {
		f := NewFeatures(true)
		if len(f.collectDisabled()) != 1 {
			t.Errorf("len(f.collectDisabled()) = %v, want 1", len(f.collectDisabled()))
		}
	})
}

func TestTLS_Enabled(t *testing.T) {
	t.Run("should use enabled", func(t *testing.T) {
		tls := TLS{CertDir: "/path/to/certs"}
		if !tls.Enabled() {
			t.Error("tls.Enabled() = false, want true")
		}
	})
	t.Run("should use disabled", func(t *testing.T) {
		tls := TLS{CertDir: " "}
		if tls.Enabled() {
			t.Error("tls.Enabled() = true, want false")
		}
	})
}

func TestTLS_Certs(t *testing.T) {
	t.Run("should use default crt and key", func(t *testing.T) {
		tls := TLS{CertDir: "/path/to/certs"}
		crt, key := tls.Certs()
		crt = normalizePath(crt)
		key = normalizePath(key)
		if crt != "/path/to/certs/tls.crt" {
			t.Errorf("crt = %v, want %v", crt, "/path/to/certs/tls.crt")
		}
		if key != "/path/to/certs/tls.key" {
			t.Errorf("key = %v, want %v", key, "/path/to/certs/tls.key")
		}
	})
	t.Run("should use custom crt and key", func(t *testing.T) {
		tls := TLS{
			CertDir:  "/path/to/certs",
			CertName: "foo.crt",
			KeyName:  "bar.key",
		}
		crt, key := tls.Certs()
		crt = normalizePath(crt)
		key = normalizePath(key)
		if crt != "/path/to/certs/foo.crt" {
			t.Errorf("crt = %v, want %v", crt, "/path/to/certs/foo.crt")
		}
		if key != "/path/to/certs/bar.key" {
			t.Errorf("key = %v, want %v", key, "/path/to/certs/bar.key")
		}
	})
}

func normalizePath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}
