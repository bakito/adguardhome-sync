package types

import (
	"go.uber.org/zap"
)

// Features feature flags
type Features struct {
	DNS             DNS  `json:"dns" yaml:"dns"`
	DHCP            DHCP `json:"dhcp" yaml:"dhcp"`
	GeneralSettings bool `json:"generalSettings" yaml:"generalSettings" env:"FEATURES_GENERAL_SETTINGS" envDefault:"true"`
	QueryLogConfig  bool `json:"queryLogConfig" yaml:"queryLogConfig" env:"FEATURES_QUERY_LOG_CONFIG" envDefault:"true"`
	StatsConfig     bool `json:"statsConfig" yaml:"statsConfig" env:"FEATURES_STATS_CONFIG" envDefault:"true"`
	ClientSettings  bool `json:"clientSettings" yaml:"clientSettings" env:"FEATURES_CLIENT_SETTINGS" envDefault:"true"`
	Services        bool `json:"services" yaml:"services" env:"FEATURES_SERVICES" envDefault:"true"`
	Filters         bool `json:"filters" yaml:"filters" env:"FEATURES_FILTERS" envDefault:"true"`
}

// DHCP features
type DHCP struct {
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" env:"FEATURES_DHCP_SERVER_CONFIG" envDefault:"true"`
	StaticLeases bool `json:"staticLeases" yaml:"staticLeases" env:"FEATURES_DHCP_STATIC_LEASES" envDefault:"true"`
}

// DNS features
type DNS struct {
	AccessLists  bool `json:"accessLists" yaml:"accessLists" env:"FEATURES_DNS_ACCESS_LISTS" envDefault:"true"`
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" env:"FEATURES_DNS_SERVER_CONFIG" envDefault:"true"`
	Rewrites     bool `json:"rewrites" yaml:"rewrites" env:"FEATURES_DNS_REWRITES" envDefault:"true"`
}

// LogDisabled log all disabled features
func (f *Features) LogDisabled(l *zap.SugaredLogger) {
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

	if len(features) > 0 {
		l.With("features", features).Info("Disabled features")
	}
}
