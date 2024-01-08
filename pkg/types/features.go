package types

import (
	"go.uber.org/zap"
)

// Features feature flags
type Features struct {
	DNS             DNS  `json:"dns" yaml:"dns"`
	DHCP            DHCP `json:"dhcp" yaml:"dhcp"`
	GeneralSettings bool `json:"generalSettings" yaml:"generalSettings" mapstructure:"GENERAL_SETTINGS"`
	QueryLogConfig  bool `json:"queryLogConfig" yaml:"queryLogConfig" mapstructure:"QUERY_LOG_CONFIG"`
	StatsConfig     bool `json:"statsConfig" yaml:"statsConfig" mapstructure:"STATS_CONFIG"`
	ClientSettings  bool `json:"clientSettings" yaml:"clientSettings" mapstructure:"CLIENT_SETTINGS"`
	Services        bool `json:"services" yaml:"services"`
	Filters         bool `json:"filters" yaml:"filters"`
}

// DHCP features
type DHCP struct {
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" mapstructure:"SERVER_CONFIG"`
	StaticLeases bool `json:"staticLeases" yaml:"staticLeases" mapstructure:"STATIC_LEASES"`
}

// DNS features
type DNS struct {
	AccessLists  bool `json:"accessLists" yaml:"accessLists" mapstructure:"ACCESS_LISTS"`
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" mapstructure:"SERVER_CONFIG"`
	Rewrites     bool `json:"rewrites" yaml:"rewrites"`
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
