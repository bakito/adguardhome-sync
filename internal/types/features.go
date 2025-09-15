package types

import (
	"go.uber.org/zap"
)

func NewFeatures(enabled bool) Features {
	return Features{
		DNS: DNS{
			AccessLists:  enabled,
			ServerConfig: enabled,
			Rewrites:     enabled,
		},
		DHCP: DHCP{
			ServerConfig: enabled,
			StaticLeases: enabled,
		},
		GeneralSettings: enabled,
		QueryLogConfig:  enabled,
		StatsConfig:     enabled,
		ClientSettings:  enabled,
		Services:        enabled,
		Filters:         enabled,
		Theme:           enabled,
		TLSConfig:       enabled,
	}
}

// Features feature flags.
type Features struct {
	DNS             DNS  `json:"dns"             yaml:"dns"`
	DHCP            DHCP `json:"dhcp"            yaml:"dhcp"`
	GeneralSettings bool `json:"generalSettings" yaml:"generalSettings" documentation:"Sync general settings" env:"FEATURES_GENERAL_SETTINGS"`
	QueryLogConfig  bool `json:"queryLogConfig"  yaml:"queryLogConfig"  documentation:"Sync query log config" env:"FEATURES_QUERY_LOG_CONFIG"`
	StatsConfig     bool `json:"statsConfig"     yaml:"statsConfig"     documentation:"Sync stats config"     env:"FEATURES_STATS_CONFIG"`
	ClientSettings  bool `json:"clientSettings"  yaml:"clientSettings"  documentation:"Sync client settings"  env:"FEATURES_CLIENT_SETTINGS"`
	Services        bool `json:"services"        yaml:"services"        documentation:"Sync services"         env:"FEATURES_SERVICES"`
	Filters         bool `json:"filters"         yaml:"filters"         documentation:"Sync filters"          env:"FEATURES_FILTERS"`
	Theme           bool `json:"theme"           yaml:"theme"           documentation:"Sync the web UI theme" env:"FEATURES_THEME"`
	TLSConfig       bool `json:"tlsConfig"       yaml:"tlsConfig"       documentation:"Sync the TLS config"   env:"FEATURES_TLS_CONFIG"`
}

// DHCP features.
type DHCP struct {
	ServerConfig bool `documentation:"Sync DHCP server config" env:"FEATURES_DHCP_SERVER_CONFIG" json:"serverConfig" yaml:"serverConfig"`
	StaticLeases bool `documentation:"Sync DHCP static leases" env:"FEATURES_DHCP_STATIC_LEASES" json:"staticLeases" yaml:"staticLeases"`
}

// DNS features.
type DNS struct {
	AccessLists  bool `documentation:"Sync DNS access lists"  env:"FEATURES_DNS_ACCESS_LISTS"  json:"accessLists"  yaml:"accessLists"`
	ServerConfig bool `documentation:"Sync DNS server config" env:"FEATURES_DNS_SERVER_CONFIG" json:"serverConfig" yaml:"serverConfig"`
	Rewrites     bool `documentation:"Sync DNS rewrites"      env:"FEATURES_DNS_REWRITES"      json:"rewrites"     yaml:"rewrites"`
}

// LogDisabled log all disabled features.
func (f *Features) LogDisabled(l *zap.SugaredLogger) {
	features := f.collectDisabled()

	if len(features) > 0 {
		l.With("features", features).Info("Disabled features")
	}
}

func (f *Features) collectDisabled() []string {
	var features []string
	if !f.DHCP.ServerConfig {
		features = append(features, "DHCP.ServerConfig")
	}
	if !f.DHCP.StaticLeases {
		features = append(features, "DHCP.StaticLeases")
	}
	if !f.DNS.AccessLists {
		features = append(features, "DNS.AccessLists")
	}
	if !f.DNS.ServerConfig {
		features = append(features, "DNS.ServerConfig")
	}
	if !f.DNS.Rewrites {
		features = append(features, "DNS.Rewrites")
	}
	if !f.GeneralSettings {
		features = append(features, "GeneralSettings")
	}
	if !f.QueryLogConfig {
		features = append(features, "QueryLogConfig")
	}
	if !f.StatsConfig {
		features = append(features, "StatsConfig")
	}
	if !f.ClientSettings {
		features = append(features, "ClientSettings")
	}
	if !f.Services {
		features = append(features, "BlockedServices")
	}
	if !f.Filters {
		features = append(features, "Filters")
	}
	if !f.TLSConfig {
		features = append(features, "TLSConfig")
	}
	return features
}
