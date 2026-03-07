package model

import (
	"testing"

	"github.com/bakito/adguardhome-sync/internal/log"
)

func TestDhcpConfigV4_isValid(t *testing.T) {
	tests := []struct {
		name string
		v4   DhcpConfigV4
		want bool
	}{
		{
			name: "When GatewayIp is nil",
			v4: DhcpConfigV4{
				GatewayIp:  nil,
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When GatewayIp is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr(""),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When SubnetMask is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: nil,
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When SubnetMask is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr(""),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeStart is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: nil,
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeStart is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr(""),
				RangeEnd:   ptr("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeEnd is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   nil,
			},
			want: false,
		},
		{
			name: "When RangeEnd is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr(""),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v4.isValid(); got != tt.want {
				t.Errorf("isValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDhcpConfigV6_isValid(t *testing.T) {
	tests := []struct {
		name string
		v6   DhcpConfigV6
		want bool
	}{
		{
			name: "When RangeStart is nil",
			v6:   DhcpConfigV6{RangeStart: nil},
			want: false,
		},
		{
			name: "When RangeStart is \"\"",
			v6:   DhcpConfigV6{RangeStart: ptr("")},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.v6.isValid(); got != tt.want {
				t.Errorf("isValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSConfig_Sanitize(t *testing.T) {
	l := log.GetLogger("test")

	t.Run("should disable UsePrivatePtrResolvers if resolvers is nil", func(t *testing.T) {
		cfg := &DNSConfig{
			UsePrivatePtrResolvers: ptr(true),
			LocalPtrUpstreams:      nil,
		}
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil || *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false, got %v", cfg.UsePrivatePtrResolvers)
		}
	})

	t.Run("should disable UsePrivatePtrResolvers if resolvers is empty", func(t *testing.T) {
		cfg := &DNSConfig{
			UsePrivatePtrResolvers: ptr(true),
			LocalPtrUpstreams:      ptr([]string{}),
		}
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil || *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false, got %v", cfg.UsePrivatePtrResolvers)
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}
