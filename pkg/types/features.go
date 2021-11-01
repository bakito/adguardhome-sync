package types

import (
	"fmt"
	"go.uber.org/zap"
	"strings"
)

// Features feature flags
type Features struct {
	DHCP DHCP `json:"dhcp" yaml:"dhcp"`
}

// DHCP features
type DHCP struct {
	ServerConfig bool `json:"serverConfig" yaml:"serverConfig"`
	StaticLeases bool `json:"staticLeases" yaml:"staticLeases"`
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
