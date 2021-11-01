package types

import (
	"fmt"
	"go.uber.org/zap"
	"strings"
)

// Features feature flags
type Features struct {
	DNS             DNS  `json:"dns" yaml:"dns"`
	DHCP            DHCP `json:"dhcp" yaml:"dhcp"`
	GeneralSettings bool `json:"generalSettings" yaml:"generalSettings"`
	QueryLogConfig  bool `json:"queryLogConfig" yaml:"queryLogConfig"`
	StatsConfig     bool `json:"statsConfig" yaml:"statsConfig"`
	ClientSettings  bool `json:"clientSettings" yaml:"clientSettings"`
	Services        bool `json:"services" yaml:"services"`
	Filters         bool `json:"filters" yaml:"filters"`
}

// DHCP features
type DHCP struct {
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig"`
	StaticLeases bool `json:"staticLeases" yaml:"staticLeases"`
}

// DNS features
type DNS struct {
	AccessLists  bool `json:"accessLists" yaml:"accessLists"`
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig"`
	Rewrites     bool `json:"rewrites" yaml:"rewrites"`
}

func (f *Features) LogDisabled(l *zap.SugaredLogger) {
	var features []string
	if !f.DHCP.ServerConfig {
		features = append(features, "DHCP.ServerConfig")
	}
	if !f.DHCP.StaticLeases {
		features = append(features, "DHCP.StaticLeases")
	}

	if len(features) > 0 {
		l.With("features", fmt.Sprintf("[%s]", strings.Join(features, ","))).Info("Disabled features")
	}
}
