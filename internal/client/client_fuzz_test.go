package client_test

import (
	"testing"
	"time"

	"github.com/bakito/adguardhome-sync/internal/client"
	"github.com/bakito/adguardhome-sync/internal/types"
)

func FuzzClientNew(f *testing.F) {
	// Seed corpus with various cookie formats and URL combinations
	seeds := []struct {
		url    string
		cookie string
		user   string
		pass   string
	}{
		{"https://localhost:3000", "ID=abc123", "", ""},
		{"http://192.168.1.1:80", "", "admin", "secret"},
		{"https://example.com", "", "", ""},
		{"", "", "", ""},
		{"https://localhost:3000", "", "", ""},
		{"https://localhost:3000", "noequalsign", "", ""},
		{"https://localhost:3000", "a=b=c", "", ""},
		{"https://localhost:3000", "key=value; extra=stuff", "", ""},
		{"not-a-url", "ID=test", "", ""},
		{"https://user:pass@example.com", "", "", ""},
	}
	for _, s := range seeds {
		f.Add(s.url, s.cookie, s.user, s.pass)
	}
	f.Fuzz(func(_ *testing.T, url, cookie, username, password string) {
		inst := types.AdGuardInstance{
			URL:      url,
			Cookie:   cookie,
			Username: username,
			Password: password,
		}
		// Init and New may return errors for invalid input; that's acceptable.
		_ = inst.Init()
		_, _ = client.New(inst, 5*time.Second)
	})
}
