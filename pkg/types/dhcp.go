package types

import (
	"encoding/json"
	"net"
	"time"
)

type DHCPServerConfig struct {
	V4            *V4ServerConfJSON `json:"v4"`
	V6            *V6ServerConfJSON `json:"v6"`
	InterfaceName string            `json:"interface_name"`
	Enabled       bool              `json:"enabled"`

	Leases       Leases `json:"leases,omitempty"`
	StaticLeases Leases `json:"static_leases,omitempty"`
}

// Equals dhcp server config equal check
func (c *DHCPServerConfig) Equals(o *DHCPServerConfig) bool {
	a, _ := json.Marshal(c)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

type V4ServerConfJSON struct {
	GatewayIP     net.IP `json:"gateway_ip"`
	SubnetMask    net.IP `json:"subnet_mask"`
	RangeStart    net.IP `json:"range_start"`
	RangeEnd      net.IP `json:"range_end"`
	LeaseDuration uint32 `json:"lease_duration"`
}

type V6ServerConfJSON struct {
	RangeStart    net.IP `json:"range_start"`
	RangeEnd      net.IP `json:"range_end"`
	LeaseDuration uint32 `json:"lease_duration"`
}

type Leases []Lease

// Merge the leases
func (l Leases) Merge(other Leases) ([]Lease, []Lease) {
	current := make(map[string]Lease)

	var adds Leases
	var removes Leases
	for _, le := range l {
		current[le.HWAddr] = le
	}

	for _, le := range other {
		if _, ok := current[le.HWAddr]; ok {
			delete(current, le.HWAddr)
		} else {
			adds = append(adds, le)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}

// Lease contains the necessary information about a DHCP lease
type Lease struct {
	HWAddr   string `json:"mac"`
	IP       net.IP `json:"ip"`
	Hostname string `json:"hostname"`

	// Lease expiration time
	// 1: static lease
	Expiry time.Time `json:"expires"`
}
