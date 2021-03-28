package types

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type Status struct {
	DNSAddresses      []string `json:"dns_addresses"`
	DNSPort           int      `json:"dns_port"`
	HTTPPort          int      `json:"http_port"`
	ProtectionEnabled bool     `json:"protection_enabled"`
	DhcpAvailable     bool     `json:"dhcp_available"`
	Running           bool     `json:"running"`
	Version           string   `json:"version"`
	Language          string   `json:"language"`
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

type FilteringConfig struct {
	Enabled  bool `json:"enabled"`
	Interval int  `json:"interval"`
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

func (s Services) Equals(o Services) bool {
	s.Sort()
	o.Sort()
	if len(s) != len(o) {
		return false
	}
	for i, v := range s {
		if v != o[i] {
			return false
		}
	}
	return true
}
