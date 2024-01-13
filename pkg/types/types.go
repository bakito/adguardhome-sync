package types

import (
	"fmt"
	"net/url"

	"go.uber.org/zap"
)

const (
	// DefaultAPIPath default api path
	DefaultAPIPath = "/control"
)

// Config application configuration struct
// +k8s:deepcopy-gen=true
type Config struct {
	Origin          AdGuardInstance   `json:"origin" yaml:"origin" env:"ORIGIN"`
	Replica         *AdGuardInstance  `json:"replica,omitempty" yaml:"replica,omitempty" env:"REPLICA"`
	Replicas        []AdGuardInstance `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Cron            string            `json:"cron,omitempty" yaml:"cron,omitempty" env:"CRON"`
	RunOnStart      bool              `json:"runOnStart,omitempty" yaml:"runOnStart,omitempty" env:"RUN_ON_START" envDefault:"true"`
	PrintConfigOnly bool              `json:"printConfigOnly,omitempty" yaml:"printConfigOnly,omitempty" env:"PRINT_CONFIG_ONLY"`
	ContinueOnError bool              `json:"continueOnError,omitempty" yaml:"continueOnError,omitempty" env:"CONTINUE_ON_ERROR"`
	API             API               `json:"api,omitempty" yaml:"api,omitempty" env:"API"`
	Features        Features          `json:"features,omitempty" yaml:"features,omitempty" env:"FEATURES_"`
}

// API configuration
type API struct {
	Port     int    `json:"port,omitempty" yaml:"port,omitempty" env:"API_PORT" envDefault:"8080"`
	Username string `json:"username,omitempty" yaml:"username,omitempty" env:"API_USERNAME"`
	Password string `json:"password,omitempty" yaml:"password,omitempty" env:"API_PASSWORD"`
	DarkMode bool   `json:"darkMode,omitempty" yaml:"darkMode,omitempty" env:"API_DARK_MODE"`
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

func (cfg *Config) Init() error {
	if err := cfg.Origin.Init(); err != nil {
		return err
	}
	for i := range cfg.Replicas {
		replica := &cfg.Replicas[i]
		if err := replica.Init(); err != nil {
			return err
		}
	}
	return nil
}

// AdGuardInstance AdguardHome config instance
// +k8s:deepcopy-gen=true
type AdGuardInstance struct {
	URL                string `json:"url" yaml:"url" env:"URL"`
	WebURL             string `json:"webURL" yaml:"webURL" env:"WEB_URL"`
	APIPath            string `json:"apiPath,omitempty" yaml:"apiPath,omitempty" env:"API_PATH" envDefault:"/control"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty" env:"USERNAME"`
	Password           string `json:"password,omitempty" yaml:"password,omitempty" env:"PASSWORD"`
	Cookie             string `json:"cookie,omitempty" yaml:"cookie,omitempty" env:"COOKIE"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify" env:"INSECURE_SKIP_VERIFY"`
	AutoSetup          bool   `json:"autoSetup" yaml:"autoSetup" env:"AUTO_SETUP"`
	InterfaceName      string `json:"interfaceName,omitempty" yaml:"interfaceName,omitempty" env:"INTERFACE_NAME"`
	DHCPServerEnabled  *bool  `json:"dhcpServerEnabled,omitempty" yaml:"dhcpServerEnabled,omitempty" env:"DHCP_SERVER_ENABLED"`

	Host    string `json:"-" yaml:"-"`
	WebHost string `json:"-" yaml:"-"`
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

func (i *AdGuardInstance) Init() error {
	u, err := url.Parse(i.URL)
	if err != nil {
		return err
	}
	i.Host = u.Host

	if i.WebURL == "" {
		i.WebHost = i.Host
		i.WebURL = i.URL
	} else {
		u, err := url.Parse(i.WebURL)
		if err != nil {
			return err
		}
		i.WebHost = u.Host
	}
	return nil
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
