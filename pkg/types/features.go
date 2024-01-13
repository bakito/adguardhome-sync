package types

import (
	"go.uber.org/zap"
)

// Features feature flags
type Features struct {
	DNS             DNS  `json:"dns" yaml:"dns" mapstructure:"dns"`
	DHCP            DHCP `json:"dhcp" yaml:"dhcp" mapstructure:"dhcp"`
	GeneralSettings bool `json:"generalSettings" yaml:"generalSettings" mapstructure:"generalSettings"`
	QueryLogConfig  bool `json:"queryLogConfig" yaml:"queryLogConfig" mapstructure:"queryLogConfig"`
	StatsConfig     bool `json:"statsConfig" yaml:"statsConfig" mapstructure:"statsConfig"`
	ClientSettings  bool `json:"clientSettings" yaml:"clientSettings" mapstructure:"clientSettings"`
	Services        bool `json:"services" yaml:"services" mapstructure:"services"`
	Filters         bool `json:"filters" yaml:"filters" mapstructure:"filters"`
}

// DHCP features
type DHCP struct {
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" mapstructure:"serverConfig"`
	StaticLeases bool `json:"staticLeases" yaml:"staticLeases" mapstructure:"staticLeases"`
}

// DNS features
type DNS struct {
	AccessLists  bool `json:"accessLists" yaml:"accessLists" mapstructure:"accessLists"`
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig" mapstructure:"serverConfig"`
	Rewrites     bool `json:"rewrites" yaml:"rewrites" mapstructure:"rewrites"`
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
