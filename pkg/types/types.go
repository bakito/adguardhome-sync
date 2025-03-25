package types

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// DefaultAPIPath default api path.
	DefaultAPIPath = "/control"
)

// Config application configuration struct
// +k8s:deepcopy-gen=true
type Config struct {
	Origin          AdGuardInstance   `env:"ORIGIN"            json:"origin"                    yaml:"origin"`
	Replica         *AdGuardInstance  `env:"REPLICA"           json:"replica,omitempty"         yaml:"replica,omitempty"`
	Replicas        []AdGuardInstance `                        json:"replicas,omitempty"        yaml:"replicas,omitempty"        faker:"slice_len=2"`
	Cron            string            `env:"CRON"              json:"cron,omitempty"            yaml:"cron,omitempty"`
	RunOnStart      bool              `env:"RUN_ON_START"      json:"runOnStart,omitempty"      yaml:"runOnStart,omitempty"`
	PrintConfigOnly bool              `env:"PRINT_CONFIG_ONLY" json:"printConfigOnly,omitempty" yaml:"printConfigOnly,omitempty"`
	ContinueOnError bool              `env:"CONTINUE_ON_ERROR" json:"continueOnError,omitempty" yaml:"continueOnError,omitempty"`
	API             API               `env:"API"               json:"api,omitempty"             yaml:"api,omitempty"`
	Features        Features          `env:"FEATURES_"         json:"features,omitempty"        yaml:"features,omitempty"`
}

// API configuration.
type API struct {
	Port     int     `env:"API_PORT"      json:"port,omitempty"     yaml:"port,omitempty"`
	Username string  `env:"API_USERNAME"  json:"username,omitempty" yaml:"username,omitempty"`
	Password string  `env:"API_PASSWORD"  json:"password,omitempty" yaml:"password,omitempty"`
	DarkMode bool    `env:"API_DARK_MODE" json:"darkMode,omitempty" yaml:"darkMode,omitempty"`
	Metrics  Metrics `env:"API_METRICS"   json:"metrics,omitempty"  yaml:"metrics,omitempty"`
	TLS      TLS     `env:"API_TLS"       json:"tls,omitempty"      yaml:"tls,omitempty"`
}

// Metrics configuration.
type Metrics struct {
	Enabled        bool          `env:"API_METRICS_ENABLED"         json:"enabled,omitempty"        yaml:"enabled,omitempty"`
	ScrapeInterval time.Duration `env:"API_METRICS_SCRAPE_INTERVAL" json:"scrapeInterval,omitempty" yaml:"scrapeInterval,omitempty"`
	QueryLogLimit  int           `env:"API_METRICS_QUERY_LOG_LIMIT" json:"queryLogLimit,omitempty"  yaml:"queryLogLimit,omitempty"`
}

// TLS configuration.
type TLS struct {
	CertDir  string `env:"API_TLS_CERT_DIR"  json:"certDir,omitempty"  yaml:"certDir,omitempty"`
	CertName string `env:"API_TLS_CERT_NAME" json:"certName,omitempty" yaml:"certName,omitempty"`
	KeyName  string `env:"API_TLS_KEY_NAME"  json:"keyName,omitempty"  yaml:"keyName,omitempty"`
}

func (t TLS) Enabled() bool {
	return strings.TrimSpace(t.CertDir) != ""
}

func (t TLS) Certs() (cert, key string) {
	cert = filepath.Join(t.CertDir, defaultIfEmpty(t.CertName, "tls.crt"))
	key = filepath.Join(t.CertDir, defaultIfEmpty(t.KeyName, "tls.key"))
	return cert, key
}

func defaultIfEmpty(val, fallback string) string {
	if strings.TrimSpace(val) == "" {
		return fallback
	}
	return val
}

// Mask maks username and password.
func (a *API) Mask() {
	a.Username = mask(a.Username)
	a.Password = mask(a.Password)
}

// UniqueReplicas get unique replication instances.
func (cfg *Config) UniqueReplicas() []AdGuardInstance {
	dedup := make(map[string]AdGuardInstance)
	if cfg.Replica != nil && cfg.Replica.URL != "" {
		if cfg.Replica.APIPath == "" {
			cfg.Replica.APIPath = DefaultAPIPath
		}
		dedup[cfg.Replica.Key()] = *cfg.Replica
	}
	for _, replica := range cfg.Replicas {
		if replica.APIPath == "" {
			replica.APIPath = DefaultAPIPath
		}
		if replica.URL != "" {
			dedup[replica.Key()] = replica
		}
	}

	var r []AdGuardInstance
	for _, replica := range dedup {
		r = append(r, replica)
	}
	return r
}

// Log the current config.
func (cfg *Config) Log(l *zap.SugaredLogger) {
	c := cfg.mask()
	l.With("config", c).Debug("Using config")
}

func (cfg *Config) mask() *Config {
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
	c.API.Mask()
	return c
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
	URL                string `env:"URL"                  faker:"url" json:"url"                         yaml:"url"`
	WebURL             string `env:"WEB_URL"              faker:"url" json:"webURL"                      yaml:"webURL"`
	APIPath            string `env:"API_PATH"                         json:"apiPath,omitempty"           yaml:"apiPath,omitempty"`
	Username           string `env:"USERNAME"                         json:"username,omitempty"          yaml:"username,omitempty"`
	Password           string `env:"PASSWORD"                         json:"password,omitempty"          yaml:"password,omitempty"`
	Cookie             string `env:"COOKIE"                           json:"cookie,omitempty"            yaml:"cookie,omitempty"`
	InsecureSkipVerify bool   `env:"INSECURE_SKIP_VERIFY"             json:"insecureSkipVerify"          yaml:"insecureSkipVerify"`
	AutoSetup          bool   `env:"AUTO_SETUP"                       json:"autoSetup"                   yaml:"autoSetup"`
	InterfaceName      string `env:"INTERFACE_NAME"                   json:"interfaceName,omitempty"     yaml:"interfaceName,omitempty"`
	DHCPServerEnabled  *bool  `env:"DHCP_SERVER_ENABLED"              json:"dhcpServerEnabled,omitempty" yaml:"dhcpServerEnabled,omitempty"`

	Host    string `json:"-" yaml:"-"`
	WebHost string `json:"-" yaml:"-"`
}

// Key AdGuardInstance key.
func (i *AdGuardInstance) Key() string {
	return fmt.Sprintf("%s#%s", i.URL, i.APIPath)
}

// Mask maks username and password.
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
	if len(s) < 3 {
		return strings.Repeat("*", len(s))
	}
	mask := strings.Repeat("*", len(s)-2)
	return fmt.Sprintf("%v%s%v", string(s[0]), mask, string(s[len(s)-1]))
}

// Protection API struct.
type Protection struct {
	ProtectionEnabled bool `json:"protection_enabled"`
}

// InstallConfig AdguardHome install config.
type InstallConfig struct {
	Web      InstallPort `json:"web"`
	DNS      InstallPort `json:"dns"`
	Username string      `json:"username"`
	Password string      `json:"password"`
}

// InstallPort AdguardHome install config port.
type InstallPort struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Status     string `json:"status"`
	CanAutofix bool   `json:"can_autofix"`
}
