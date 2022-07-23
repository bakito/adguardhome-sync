package types

import (
	"encoding/json"
	"net"
	"sort"
)

// DNSConfig dns config
// +k8s:deepcopy-gen=true
type DNSConfig struct {
	Upstreams     []string `json:"upstream_dns,omitempty"`
	UpstreamsFile string   `json:"upstream_dns_file"`
	Bootstraps    []string `json:"bootstrap_dns,omitempty"`

	ProtectionEnabled bool     `json:"protection_enabled"`
	RateLimit         uint32   `json:"ratelimit"`
	BlockingMode      string   `json:"blocking_mode,omitempty"`
	BlockingIPv4      net.IP   `json:"blocking_ipv4,omitempty"`
	BlockingIPv6      net.IP   `json:"blocking_ipv6"`
	EDNSCSEnabled     bool     `json:"edns_cs_enabled"`
	DNSSECEnabled     bool     `json:"dnssec_enabled"`
	DisableIPv6       bool     `json:"disable_ipv6"`
	UpstreamMode      string   `json:"upstream_mode,omitempty"`
	CacheSize         uint32   `json:"cache_size"`
	CacheMinTTL       uint32   `json:"cache_ttl_min"`
	CacheMaxTTL       uint32   `json:"cache_ttl_max"`
	CacheOptimistic   bool     `json:"cache_optimistic"`
	ResolveClients    bool     `json:"resolve_clients"`
	LocalPTRUpstreams []string `json:"local_ptr_upstreams,omitempty"`
}

// Equals dns config equal check
func (c *DNSConfig) Equals(o *DNSConfig) bool {
	cc := c.DeepCopy()
	oo := o.DeepCopy()
	cc.Sort()
	oo.Sort()

	a, _ := json.Marshal(cc)
	b, _ := json.Marshal(oo)
	return string(a) == string(b)
}

// Sort sort dns config
func (c *DNSConfig) Sort() {
	sort.Strings(c.Upstreams)
	sort.Strings(c.Bootstraps)
	sort.Strings(c.LocalPTRUpstreams)
}

// AccessList access list
type AccessList struct {
	AllowedClients    []string `json:"allowed_clients"`
	DisallowedClients []string `json:"disallowed_clients"`
	BlockedHosts      []string `json:"blocked_hosts"`
}

// Equals access list equal check
func (al *AccessList) Equals(o *AccessList) bool {
	return equals(al.AllowedClients, o.AllowedClients) &&
		equals(al.DisallowedClients, o.DisallowedClients) &&
		equals(al.BlockedHosts, o.BlockedHosts)
}

// Sort sort access list
func (al *AccessList) Sort() {
	sort.Strings(al.AllowedClients)
	sort.Strings(al.DisallowedClients)
	sort.Strings(al.BlockedHosts)
}
