package model

import (
	"reflect"
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
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When GatewayIp is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  new(""),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When SubnetMask is nil",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: nil,
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When SubnetMask is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new(""),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeStart is nil",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: nil,
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeStart is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new(""),
				RangeEnd:   new("4.4.4.4"),
			},
			want: false,
		},
		{
			name: "When RangeEnd is nil",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   nil,
			},
			want: false,
		},
		{
			name: "When RangeEnd is \"\"",
			v4: DhcpConfigV4{
				GatewayIp:  new("1.1.1.1"),
				SubnetMask: new("2.2.2.2"),
				RangeStart: new("3.3.3.3"),
				RangeEnd:   new(""),
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
			v6:   DhcpConfigV6{RangeStart: new("")},
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
			UsePrivatePtrResolvers: new(true),
			LocalPtrUpstreams:      nil,
		}
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil || *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false, got %v", cfg.UsePrivatePtrResolvers)
		}
	})

	t.Run("should disable UsePrivatePtrResolvers if resolvers is empty", func(t *testing.T) {
		cfg := &DNSConfig{
			UsePrivatePtrResolvers: new(true),
			LocalPtrUpstreams:      new([]string{}),
		}
		cfg.Sanitize(l)
		if cfg.UsePrivatePtrResolvers == nil || *cfg.UsePrivatePtrResolvers {
			t.Errorf("expected UsePrivatePtrResolvers to be false, got %v", cfg.UsePrivatePtrResolvers)
		}
	})
}

func TestDhcpStatus_cleanV4V6(t *testing.T) {
	tests := []struct {
		name    string
		ds      *DhcpStatus
		checkV4 bool
		checkV6 bool
	}{
		{
			name: "should set V4 and V6 to nil if they are invalid",
			ds: &DhcpStatus{
				V4: &DhcpConfigV4{},
				V6: &DhcpConfigV6{},
			},
			checkV4: false,
			checkV6: false,
		},
		{
			name: "should keep V4 and V6 if they are valid",
			ds: &DhcpStatus{
				V4: &DhcpConfigV4{
					GatewayIp:  new("1.1.1.1"),
					SubnetMask: new("255.255.255.0"),
					RangeStart: new("1.1.1.2"),
					RangeEnd:   new("1.1.1.10"),
				},
				V6: &DhcpConfigV6{
					RangeStart: new("2001:db8::1"),
				},
			},
			checkV4: true,
			checkV6: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.ds.cleanV4V6()
			if (tt.ds.V4 != nil) != tt.checkV4 {
				t.Errorf("V4 nil check failed, got %v, want %v", tt.ds.V4 != nil, tt.checkV4)
			}
			if (tt.ds.V6 != nil) != tt.checkV6 {
				t.Errorf("V6 nil check failed, got %v, want %v", tt.ds.V6 != nil, tt.checkV6)
			}
		})
	}
}

func TestStats_sumUp(t *testing.T) {
	tests := []struct {
		name string
		s1   *[]int
		s2   *[]int
		want *[]int
	}{
		{
			name: "should sum up slices",
			s1:   new([]int{1, 2, 3}),
			s2:   new([]int{4, 5, 6}),
			want: new([]int{5, 7, 9}),
		},
		{
			name: "should handle different lengths",
			s1:   new([]int{1, 2}),
			s2:   new([]int{4, 5, 6}),
			want: new([]int{5, 7}),
		},
		{
			name: "should return target if other is nil",
			s1:   new([]int{1, 2}),
			s2:   nil,
			want: new([]int{1, 2}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := sumUp(tt.s1, tt.s2)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("sumUp() = %v, want %v", res, tt.want)
			}
		})
	}
}

func TestStats_addInt(t *testing.T) {
	tests := []struct {
		name string
		t    *int
		add  *int
		want *int
	}{
		{
			name: "should add int",
			t:    new(1),
			add:  new(2),
			want: new(3),
		},
		{
			name: "should return t if add is nil",
			t:    new(1),
			add:  nil,
			want: new(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := addInt(tt.t, tt.add)
			if !reflect.DeepEqual(res, tt.want) {
				t.Errorf("addInt() = %v, want %v", res, tt.want)
			}
		})
	}
}

func TestPtrEquals(t *testing.T) {
	tests := []struct {
		name string
		a    *int
		b    *int
		want bool
	}{
		{
			name: "both nil",
			a:    nil,
			b:    nil,
			want: true,
		},
		{
			name: "a nil",
			a:    nil,
			b:    new(1),
			want: false,
		},
		{
			name: "b nil",
			a:    new(1),
			b:    nil,
			want: false,
		},
		{
			name: "equal",
			a:    new(1),
			b:    new(1),
			want: true,
		},
		{
			name: "not equal",
			a:    new(1),
			b:    new(2),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ptrEquals(tt.a, tt.b); got != tt.want {
				t.Errorf("ptrEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSConfig_Sort_Private(t *testing.T) {
	cfg := &DNSConfig{
		UpstreamDns:       new([]string{"b", "a"}),
		BootstrapDns:      new([]string{"d", "c"}),
		LocalPtrUpstreams: new([]string{"f", "e"}),
	}
	cfg.Sort()
	if !reflect.DeepEqual(*cfg.UpstreamDns, []string{"a", "b"}) {
		t.Errorf("UpstreamDns sort failed, got %v", *cfg.UpstreamDns)
	}
	if !reflect.DeepEqual(*cfg.BootstrapDns, []string{"c", "d"}) {
		t.Errorf("BootstrapDns sort failed, got %v", *cfg.BootstrapDns)
	}
	if !reflect.DeepEqual(*cfg.LocalPtrUpstreams, []string{"e", "f"}) {
		t.Errorf("LocalPtrUpstreams sort failed, got %v", *cfg.LocalPtrUpstreams)
	}
}

func TestClient_Sort_Private(t *testing.T) {
	cl := &Client{
		Ids:             new([]string{"b", "a"}),
		Tags:            new([]string{"d", "c"}),
		BlockedServices: new([]string{"f", "e"}),
		Upstreams:       new([]string{"h", "g"}),
	}
	cl.Sort()
	if !reflect.DeepEqual(*cl.Ids, []string{"a", "b"}) {
		t.Errorf("Ids sort failed, got %v", *cl.Ids)
	}
	if !reflect.DeepEqual(*cl.Tags, []string{"c", "d"}) {
		t.Errorf("Tags sort failed, got %v", *cl.Tags)
	}
	if !reflect.DeepEqual(*cl.BlockedServices, []string{"e", "f"}) {
		t.Errorf("BlockedServices sort failed, got %v", *cl.BlockedServices)
	}
	if !reflect.DeepEqual(*cl.Upstreams, []string{"g", "h"}) {
		t.Errorf("Upstreams sort failed, got %v", *cl.Upstreams)
	}
}
