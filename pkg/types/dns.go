package types

import (
	"encoding/json"
	"net"
	"sort"
)

// https://ha.bakito.net:3000/control/dns_config
// {"bootstrap_dns":["1.1.1.1:53"],"upstream_mode":"parallel","upstream_dns":["https://dns10.quad9.net/dns-query"]}
// {"bootstrap_dns":["1.1.1.1:53"],"upstream_mode":"","upstream_dns":["https://dns10.quad9.net/dns-query"]}
// {"bootstrap_dns":["1.1.1.1:53"],"upstream_mode":"fastest_addr","upstream_dns":["https://dns10.quad9.net/dns-query"]}

// {"ratelimit":20,"blocking_mode":"default","blocking_ipv4":"0.0.0.0","blocking_ipv6":"::","edns_cs_enabled":true,"disable_ipv6":false,"dnssec_enabled":false}
// {"cache_size":4194304,"cache_ttl_max":0,"cache_ttl_min":0}

// https://ha.bakito.net:3000/control/access/set
// {"allowed_clients":["2.2.2.2"],"disallowed_clients":["1.1.1.1"],"blocked_hosts":["version.bind","id.server","hostname.bind"]}
// https://ha.bakito.net:3000/control/access/list
// {"allowed_clients":[],"disallowed_clients":[],"blocked_hosts":["version.bind","id.server","hostname.bind"]}

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
	ResolveClients    bool     `json:"resolve_clients"`
	LocalPTRUpstreams []string `json:"local_ptr_upstreams,omitempty"`
}

// Equal dns config equal check
func (c *DNSConfig) Equal(o *DNSConfig) bool {
	c.Sort()
	o.Sort()

	a, _ := json.Marshal(c)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

// Sort sort dns config
func (c *DNSConfig) Sort() {
	sort.Strings(c.Upstreams)
	sort.Strings(c.Bootstraps)
	sort.Strings(c.LocalPTRUpstreams)
}

type AccessList struct {
	AllowedClients    []string `json:"allowed_clients"`
	DisallowedClients []string `json:"disallowed_clients"`
	BlockedHosts      []string `json:"blocked_hosts"`
}

// Equal access list equal check
func (al *AccessList) Equal(o *AccessList) bool {
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
