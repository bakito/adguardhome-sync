package types

import (
	"fmt"

	"go.uber.org/zap"
)

const (
	// DefaultAPIPath default api path
	DefaultAPIPath = "/control"
)

// Config application configuration struct
// +k8s:deepcopy-gen=true
type Config struct {
	Origin          AdGuardInstance   `json:"origin" yaml:"origin"`
	Replica         *AdGuardInstance  `json:"replica,omitempty" yaml:"replica,omitempty"`
	Replicas        []AdGuardInstance `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Cron            string            `json:"cron,omitempty" yaml:"cron,omitempty"`
	RunOnStart      bool              `json:"runOnStart,omitempty" yaml:"runOnStart,omitempty"`
	PrintConfigOnly bool              `json:"printConfigOnly,omitempty" yaml:"printConfigOnly,omitempty"`
	API             API               `json:"api,omitempty" yaml:"api,omitempty"`
	Features        Features          `json:"features,omitempty" yaml:"features,omitempty"`
}

// API configuration
type API struct {
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	DarkMode bool   `json:"darkMode,omitempty" yaml:"darkMode,omitempty"`
}

// UniqueReplicas get unique replication instances
func (cfg *Config) UniqueReplicas() []AdGuardInstance {
	dedup := make(map[string]AdGuardInstance)
	if cfg.Replica != nil && cfg.Replica.URL != "" {
		dedup[cfg.Replica.Key()] = *cfg.Replica
	}
	for _, replica := range cfg.Replicas {
		if replica.URL != "" {
			dedup[replica.Key()] = replica
		}
	}

	var r []AdGuardInstance
	for _, replica := range dedup {
		if replica.APIPath == "" {
			replica.APIPath = DefaultAPIPath
		}
		r = append(r, replica)
	}
	return r
}

// Log the current config
func (cfg *Config) Log(l *zap.SugaredLogger) {
	c := cfg.DeepCopy()
	c.Origin.Mask()
	if c.Replica != nil {
		if c.Replica.URL == "" {
			c.Replica = nil
		} else {
			c.Replica.Mask()
		}
	}
	for i := range c.Replicas {
		c.Replicas[i].Mask()
	}
	l.With("config", c).Debug("Using config")
}

// AdGuardInstance AdguardHome config instance
// +k8s:deepcopy-gen=true
type AdGuardInstance struct {
	URL                string `json:"url" yaml:"url"`
	APIPath            string `json:"apiPath,omitempty" yaml:"apiPath,omitempty"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty"`
	Password           string `json:"password,omitempty" yaml:"password,omitempty"`
	Cookie             string `json:"cookie,omitempty" yaml:"cookie,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	AutoSetup          bool   `json:"autoSetup" yaml:"autoSetup"`
	InterfaceName      string `json:"interfaceName,omitempty" yaml:"interfaceName,omitempty"`
	DHCPServerEnabled  *bool  `json:"dhcpServerEnabled,omitempty" yaml:"dhcpServerEnabled,omitempty"`
}

// Key AdGuardInstance key
func (i *AdGuardInstance) Key() string {
	return fmt.Sprintf("%s#%s", i.URL, i.APIPath)
}

// Mask maks username and password
func (i *AdGuardInstance) Mask() {
	i.Username = mask(i.Username)
	i.Password = mask(i.Password)
}

func mask(s string) string {
	if s == "" {
		return "***"
	}
	return fmt.Sprintf("%v***%v", string(s[0]), string(s[len(s)-1]))
}

// Protection API struct
type Protection struct {
	ProtectionEnabled bool `json:"protection_enabled"`
}

// InstallConfig AdguardHome install config
type InstallConfig struct {
	Web      InstallPort `json:"web"`
	DNS      InstallPort `json:"dns"`
	Username string      `json:"username"`
	Password string      `json:"password"`
}

// InstallPort AdguardHome install config port
type InstallPort struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	CanAutofix bool   `json:"can_autofix"`
}
