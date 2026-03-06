package versions_test

import (
	"testing"

	"github.com/bakito/adguardhome-sync/internal/versions"
)

func TestIsNewerThan(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want bool
	}{
		{"v0.106.10", "v0.106.9", true},
		{"v0.106.9", "v0.106.10", false},
		{"v0.106.10", "0.106.9", true},
		{"v0.106.9", "0.106.10", false},
		{"v0.108.0-b.72", versions.MinAgh, true},
		{"0.108.0-b.72", versions.MinAgh, true},
		{versions.MinAgh, "v0.108.0-b.72", false},
		{versions.MinAgh, "0.108.0-b.72", false},
	}
	for _, tt := range tests {
		t.Run(tt.v1+" > "+tt.v2, func(t *testing.T) {
			if got := versions.IsNewerThan(tt.v1, tt.v2); got != tt.want {
				t.Errorf("IsNewerThan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSame(t *testing.T) {
	tests := []struct {
		v1   string
		v2   string
		want bool
	}{
		{"v0.106.9", "v0.106.9", true},
		{"0.106.9", "v0.106.9", true},
	}
	for _, tt := range tests {
		t.Run(tt.v1+" == "+tt.v2, func(t *testing.T) {
			if got := versions.IsSame(tt.v1, tt.v2); got != tt.want {
				t.Errorf("IsSame() = %v, want %v", got, tt.want)
			}
		})
	}
}
