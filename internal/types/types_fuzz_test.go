package types

import (
	"testing"
)

func FuzzMask(f *testing.F) {
	testcases := []string{"", "a", "ab", "abc", "abcd"}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(_ *testing.T, value string) {
		_ = mask(value)
	})
}

func FuzzAdGuardInstance_Init(f *testing.F) {
	// Seed corpus with valid and edge-case URLs
	seeds := []struct {
		url    string
		webURL string
	}{
		{"https://localhost:3000", ""},
		{"http://192.168.1.1:80", ""},
		{"https://example.com", "https://web.example.com"},
		{"", ""},
		{"not-a-url", ""},
		{"http://[::1]:8080", ""},
		{"ftp://example.com", ""},
		{"https://user:pass@example.com", ""},
		{"https://example.com/path/to/api", ""},
		{"https://example.com:443", "http://example.com:80"},
	}
	for _, s := range seeds {
		f.Add(s.url, s.webURL)
	}
	f.Fuzz(func(_ *testing.T, url, webURL string) {
		inst := AdGuardInstance{
			URL:    url,
			WebURL: webURL,
		}
		// Init may return an error for invalid URLs; that's acceptable.
		_ = inst.Init()
	})
}

func FuzzConfig_Init(f *testing.F) {
	seeds := []struct {
		originURL    string
		originWebURL string
		replicaURL   string
	}{
		{"https://origin:3000", "", "https://replica:3000"},
		{"http://127.0.0.1:80", "", "http://127.0.0.1:81"},
		{"", "", ""},
		{"not-a-url", "", "also-not-a-url"},
	}
	for _, s := range seeds {
		f.Add(s.originURL, s.originWebURL, s.replicaURL)
	}
	f.Fuzz(func(_ *testing.T, originURL, originWebURL, replicaURL string) {
		cfg := Config{
			Origin: &AdGuardInstance{
				URL:    originURL,
				WebURL: originWebURL,
			},
			Replicas: []Replica{
				{URL: replicaURL},
			},
		}
		_ = cfg.Init()
	})
}
