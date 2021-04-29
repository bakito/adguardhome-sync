package types

import (
	"net"
	"time"
)

type DHCPServerConfigJSON struct {
	V4            *V4ServerConfJSON `json:"v4"`
	V6            *V6ServerConfJSON `json:"v6"`
	InterfaceName string            `json:"interface_name"`
	Enabled       bool              `json:"enabled"`
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
	LeaseDuration uint32 `json:"lease_duration"`
}

// https://ha.bakito.net:3000/control/dhcp/status

// https://ha.bakito.net:3000/control/dhcp/set_config
// set {"interface_name":"docker0","v4":{"gateway_ip":"172.17.0.1","range_start":"172.17.0.100","range_end":"172.17.0.200","subnet_mask":"255.255.255.0","lease_duration":888888}}

// dhcpStatusResponse is the response for /control/dhcp/status endpoint.
type DHCPStatusResponse struct {
	Enabled      bool         `json:"enabled"`
	IfaceName    string       `json:"interface_name"`
	V4           V4ServerConf `json:"v4"`
	V6           V6ServerConf `json:"v6"`
	Leases       []Lease      `json:"leases"`
	StaticLeases []Lease      `json:"static_leases"`
}

// V4ServerConf - server configuration
type V4ServerConf struct {
	GatewayIP  net.IP `yaml:"gateway_ip" json:"gateway_ip"`
	SubnetMask net.IP `yaml:"subnet_mask" json:"subnet_mask"`

	// The first & the last IP address for dynamic leases
	// Bytes [0..2] of the last allowed IP address must match the first IP
	RangeStart net.IP `yaml:"range_start" json:"range_start"`
	RangeEnd   net.IP `yaml:"range_end" json:"range_end"`

	LeaseDuration uint32 `yaml:"lease_duration" json:"lease_duration"` // in seconds
}

// V6ServerConf - server configuration
type V6ServerConf struct {
	// The first IP address for dynamic leases
	// The last allowed IP address ends with 0xff byte
	RangeStart net.IP `yaml:"range_start" json:"range_start"`

	LeaseDuration uint32 `yaml:"lease_duration" json:"lease_duration"` // in seconds

	RASLAACOnly  bool `yaml:"ra_slaac_only" json:"-"`  // send ICMPv6.RA packets without MO flags
	RAAllowSLAAC bool `yaml:"ra_allow_slaac" json:"-"` // send ICMPv6.RA packets with MO flags
}

// Lease contains the necessary information about a DHCP lease
type Lease struct {
	HWAddr   net.HardwareAddr `json:"mac"`
	IP       net.IP           `json:"ip"`
	Hostname string           `json:"hostname"`

	// Lease expiration time
	// 1: static lease
	Expiry time.Time `json:"expires"`
}

// POST https://ha.bakito.net:3000/control/dhcp/add_static_lease
// {"mac":"00:80:41:ae:fd:7e","ip":"1.1.2.3","hostname":"dddd"}

// POST https://ha.bakito.net:3000/control/dhcp/remove_static_lease
// {"ip":"1.1.2.3","mac":"00:80:41:ae:fd:7e","hostname":"dddd"}
