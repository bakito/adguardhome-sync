package model

import (
	"testing"

	"github.com/bakito/adguardhome-sync/internal/log"
)

func ptr[T any](v T) *T {
	return &v
}

func TestDhcpConfigV4_isValid(t *testing.T) {
	tests := []struct {
		name string
		v4   DhcpConfigV4
	}{
		{
			name: "When GatewayIp is nil",
			v4: DhcpConfigV4{
				GatewayIp:  nil,
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When GatewayIp is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr(""),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When SubnetMask is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: nil,
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When SubnetMask is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr(""),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When RangeStart is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: nil,
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When RangeStart is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr(""),
				RangeEnd:   ptr("4.4.4.4"),
			},
		},
		{
			name: "When RangeEnd is nil",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   nil,
			},
		},
		{
			name: "When RangeEnd is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  ptr("1.1.1.1"),
				SubnetMask: ptr("2.2.2.2"),
				RangeStart: ptr("3.3.3.3"),
				RangeEnd:   ptr(""),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v4.isValid() {
				t.Error("expected isValid to be false")
			}
		})
	}
}

func TestDhcpConfigV6_isValid(t *testing.T) {
	tests := []struct {
		name string
		v6   DhcpConfigV6
	}{
		{
			name: "When RangeStart is nil",
			v6:   DhcpConfigV6{RangeStart: nil},
		},
		{
			name: "When RangeStart is \"\"",
			v6:   DhcpConfigV6{RangeStart: ptr("")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.v6.isValid() {
				t.Errorf("expected isValid to be false")
			}
		})
	}
}

func TestDNSConfig_Sanitize(t *testing.T) {
	l := log.GetLogger("test")

	t.Run("should disable UsePrivatePtrResolvers resolvers is nil ", func(t *testing.T) {
		cfg := &DNSConfig{
			UsePrivatePtrResolvers: ptr(true),
		}
		cfg.LocalPtrUpstreams = nil
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil {
			t.Fatal("expected UsePrivatePtrResolvers to be non-nil")
		}
		if *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false")
		}
	})
	t.Run("should disable UsePrivatePtrResolvers resolvers is empty ", func(t *testing.T) {
		cfg := &DNSConfig{
			UsePrivatePtrResolvers: ptr(true),
		}
		cfg.LocalPtrUpstreams = ptr([]string{})
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil {
			t.Fatal("expected UsePrivatePtrResolvers to be non-nil")
		}
		if *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false")
		}
	})
}
