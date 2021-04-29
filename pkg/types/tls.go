package types

// https://ha.bakito.net:3000/control/tls/status

type TLSConfigSettings struct {
	Enabled         bool   `yaml:"enabled" json:"enabled"`                                 // Enabled is the encryption (DOT/DOH/HTTPS) status
	ServerName      string `yaml:"server_name" json:"server_name,omitempty"`               // ServerName is the hostname of your HTTPS/TLS server
	ForceHTTPS      bool   `yaml:"force_https" json:"force_https,omitempty"`               // ForceHTTPS: if true, forces HTTP->HTTPS redirect
	PortHTTPS       int    `yaml:"port_https" json:"port_https,omitempty"`                 // HTTPS port. If 0, HTTPS will be disabled
	PortDNSOverTLS  int    `yaml:"port_dns_over_tls" json:"port_dns_over_tls,omitempty"`   // DNS-over-TLS port. If 0, DOT will be disabled
	PortDNSOverQUIC int    `yaml:"port_dns_over_quic" json:"port_dns_over_quic,omitempty"` // DNS-over-QUIC port. If 0, DoQ will be disabled

	// PortDNSCrypt is the port for DNSCrypt requests.  If it's zero,
	// DNSCrypt is disabled.
	PortDNSCrypt int `yaml:"port_dnscrypt" json:"port_dnscrypt"`
	// DNSCryptConfigFile is the path to the DNSCrypt config file.  Must be
	// set if PortDNSCrypt is not zero.
	//
	// See https://github.com/AdguardTeam/dnsproxy and
	// https://github.com/ameshkov/dnscrypt.
	DNSCryptConfigFile string `yaml:"dnscrypt_config_file" json:"dnscrypt_config_file"`

	// Allow DOH queries via unencrypted HTTP (e.g. for reverse proxying)
	AllowUnencryptedDOH bool `yaml:"allow_unencrypted_doh" json:"allow_unencrypted_doh"`
}
