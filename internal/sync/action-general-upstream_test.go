package sync

import (
	"testing"

	"github.com/bakito/adguardhome-sync/internal/client/model"
)

func TestNormalizeUpstreamMode(t *testing.T) {
	for _, tc := range []struct {
		name, input, want string
	}{
		{"legacy load balancing", "", "load_balance"},
		{"explicit load balancing", "load_balance", "load_balance"},
		{"parallel upstreams", "parallel", "parallel"},
		{"fastest address", "fastest_addr", "fastest_addr"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := &model.DNSConfig{UpstreamMode: tc.input}
			normalizeUpstreamMode(config)
			if config.UpstreamMode != tc.want {
				t.Errorf("UpstreamMode = %q, want %q", config.UpstreamMode, tc.want)
			}
		})
	}
	t.Run("nil config", func(t *testing.T) { normalizeUpstreamMode(nil) })
}
