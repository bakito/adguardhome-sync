package model

import (
	"encoding/json"

	"github.com/jinzhu/copier"
)

// Clone the config
func (c *DhcpStatus) Clone() *DhcpStatus {
	clone := &DhcpStatus{}
	_ = copier.Copy(clone, c)
	return clone
}

// Equals dhcp server config equal check
func (c *DhcpStatus) Equals(o *DhcpStatus) bool {
	a, _ := json.Marshal(c)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

func (c *DhcpStatus) HasConfig() bool {
	return (c.V4 != nil && c.V4.isValid()) || (c.V6 != nil && c.V6.isValid())
}

func (j DhcpConfigV4) isValid() bool {
	return j.GatewayIp != nil && j.SubnetMask != nil && j.RangeStart != nil && j.RangeEnd != nil
}

func (j DhcpConfigV6) isValid() bool {
	return j.RangeStart != nil
}

type DhcpStaticLeases []DhcpStaticLease

// MergeDhcpStaticLeases the leases
func MergeDhcpStaticLeases(l *[]DhcpStaticLease, other *[]DhcpStaticLease) (DhcpStaticLeases, DhcpStaticLeases) {
	var thisLeases []DhcpStaticLease
	var otherLeases []DhcpStaticLease

	if l != nil {
		thisLeases = *l
	}
	if other != nil {
		otherLeases = *other
	}
	current := make(map[string]DhcpStaticLease)

	var adds DhcpStaticLeases
	var removes DhcpStaticLeases
	for _, le := range thisLeases {
		current[le.Mac] = le
	}

	for _, le := range otherLeases {
		if _, ok := current[le.Mac]; ok {
			delete(current, le.Mac)
		} else {
			adds = append(adds, le)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}
