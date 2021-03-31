package types

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Config struct {
	Origin   AdGuardInstance   `json:"origin" yaml:"origin"`
	Replica  *AdGuardInstance  `json:"replica,omitempty" yaml:"replica,omitempty"`
	Replicas []AdGuardInstance `json:"replicas,omitempty" yaml:"replicas,omitempty"`
	Cron     string            `json:"cron,omitempty" yaml:"cron,omitempty"`
	API      API               `json:"api,omitempty" yaml:"api,omitempty"`
}

type API struct {
	Port     int    `json:"port,omitempty" yaml:"port,omitempty"`
	Username string `json:"username,omitempty" yaml:"username,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

func (cfg *Config) UniqueReplicas() []AdGuardInstance {
	dedup := make(map[string]AdGuardInstance)
	if cfg.Replica != nil {
		dedup[cfg.Replica.Key()] = *cfg.Replica
	}
	for _, replica := range cfg.Replicas {
		dedup[replica.Key()] = replica
	}

	var r []AdGuardInstance
	for _, replica := range dedup {
		r = append(r, replica)
	}
	return r
}

type AdGuardInstance struct {
	URL                string `json:"url" yaml:"url"`
	APIPath            string `json:"apiPath,omitempty" yaml:"apiPath,omitempty"`
	Username           string `json:"username,omitempty" yaml:"username,omitempty"`
	Password           string `json:"password,omitempty" yaml:"password,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
}

func (i *AdGuardInstance) Key() string {
	return fmt.Sprintf("%s%s", i.URL, i.APIPath)
}

type Protection struct {
	ProtectionEnabled bool `json:"protection_enabled"`
}

type Status struct {
	Protection
	DNSAddresses  []string `json:"dns_addresses"`
	DNSPort       int      `json:"dns_port"`
	HTTPPort      int      `json:"http_port"`
	DhcpAvailable bool     `json:"dhcp_available"`
	Running       bool     `json:"running"`
	Version       string   `json:"version"`
	Language      string   `json:"language"`
}

type RewriteEntries []RewriteEntry

func (rwe *RewriteEntries) Merge(other *RewriteEntries) (RewriteEntries, RewriteEntries) {
	current := make(map[string]RewriteEntry)

	var adds RewriteEntries
	var removes RewriteEntries
	for _, rr := range *rwe {
		current[rr.Key()] = rr
	}

	for _, rr := range *other {
		if _, ok := current[rr.Key()]; ok {
			delete(current, rr.Key())
		} else {
			adds = append(adds, rr)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}

type RewriteEntry struct {
	Domain string `json:"domain"`
	Answer string `json:"answer"`
}

func (re *RewriteEntry) Key() string {
	return fmt.Sprintf("%s#%s", re.Domain, re.Answer)
}

type Filters []Filter

type Filter struct {
	ID          int       `json:"id"`
	Enabled     bool      `json:"enabled"`
	URL         string    `json:"url"`  // needed for add
	Name        string    `json:"name"` // needed for add
	RulesCount  int       `json:"rules_count"`
	LastUpdated time.Time `json:"last_updated"`
	Whitelist   bool      `json:"whitelist"` // needed for add
}

type FilteringStatus struct {
	FilteringConfig
	Filters          Filters   `json:"filters"`
	WhitelistFilters Filters   `json:"whitelist_filters"`
	UserRules        UserRules `json:"user_rules"`
}

type UserRules []string

func (ur UserRules) String() string {
	return strings.Join(ur, "\n")
}

type FeatureStatus struct {
	Enabled bool `json:"enabled"`
}

type FilteringConfig struct {
	FeatureStatus
	Interval int `json:"interval"`
}

type RefreshFilter struct {
	Whitelist bool `json:"whitelist"`
}

func (fs *Filters) Merge(other Filters) (Filters, Filters) {
	current := make(map[string]Filter)

	var adds Filters
	var removes Filters
	for _, f := range *fs {
		current[f.URL] = f
	}

	for _, rr := range other {
		if _, ok := current[rr.URL]; ok {
			delete(current, rr.URL)
		} else {
			adds = append(adds, rr)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}

type Services []string

func (s Services) Sort() {
	sort.Strings(s)
}

func (s *Services) Equals(o *Services) bool {
	s.Sort()
	o.Sort()
	return equals(*s, *o)
}

type Clients struct {
	Clients     []Client `json:"clients"`
	AutoClients []struct {
		IP        string `json:"ip"`
		Name      string `json:"name"`
		Source    string `json:"source"`
		WhoisInfo struct {
		} `json:"whois_info"`
	} `json:"auto_clients"`
	SupportedTags []string `json:"supported_tags"`
}

type Client struct {
	Ids             []string `json:"ids"`
	Tags            []string `json:"tags"`
	BlockedServices []string `json:"blocked_services"`
	Upstreams       []string `json:"upstreams"`

	UseGlobalSettings        bool   `json:"use_global_settings"`
	UseGlobalBlockedServices bool   `json:"use_global_blocked_services"`
	Name                     string `json:"name"`
	FilteringEnabled         bool   `json:"filtering_enabled"`
	ParentalEnabled          bool   `json:"parental_enabled"`
	SafesearchEnabled        bool   `json:"safesearch_enabled"`
	SafebrowsingEnabled      bool   `json:"safebrowsing_enabled"`
	Disallowed               bool   `json:"disallowed"`
	DisallowedRule           string `json:"disallowed_rule"`
}

func (cl *Client) Sort() {
	sort.Strings(cl.Ids)
	sort.Strings(cl.Tags)
	sort.Strings(cl.BlockedServices)
	sort.Strings(cl.Upstreams)
}

func (cl *Client) Equal(o *Client) bool {
	cl.Sort()
	o.Sort()

	a, _ := json.Marshal(cl)
	b, _ := json.Marshal(o)
	return string(a) == string(b)
}

func (clients *Clients) Merge(other *Clients) ([]Client, []Client, []Client) {
	current := make(map[string]Client)
	for _, client := range clients.Clients {
		current[client.Name] = client
	}

	expected := make(map[string]Client)
	for _, client := range other.Clients {
		expected[client.Name] = client
	}

	var adds []Client
	var removes []Client
	var updates []Client

	for _, cl := range expected {
		if oc, ok := current[cl.Name]; ok {
			if !cl.Equal(&oc) {
				updates = append(updates, cl)
			}
			delete(current, cl.Name)
		} else {
			adds = append(adds, cl)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, updates, removes
}

type ClientUpdate struct {
	Name string `json:"name"`
	Data Client `json:"data"`
}

func equals(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
