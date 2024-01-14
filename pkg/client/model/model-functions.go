package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/jinzhu/copier"
)

// Clone the config
func (c *DhcpStatus) Clone() *DhcpStatus {
	clone := &DhcpStatus{}
	_ = copier.Copy(clone, c)
	return clone
}

func (c *DhcpStatus) cleanV4V6() {
	if c.V4 != nil && !c.V4.isValid() {
		c.V4 = nil
	}
	if c.V6 != nil && !c.V6.isValid() {
		c.V6 = nil
	}
}

// CleanAndEquals dhcp server config equal check where V4 and V6 are cleaned in advance
func (c *DhcpStatus) CleanAndEquals(o *DhcpStatus) bool {
	c.cleanV4V6()
	o.cleanV4V6()
	return c.Equals(o)
}

// Equals dhcp server config equal check
func (c *DhcpStatus) Equals(o *DhcpStatus) bool {
	return utils.JsonEquals(c, o)
}

func (c *DhcpStatus) HasConfig() bool {
	return (c.V4 != nil && c.V4.isValid()) || (c.V6 != nil && c.V6.isValid())
}

func (j DhcpConfigV4) isValid() bool {
	return j.GatewayIp != nil && *j.GatewayIp != "" &&
		j.SubnetMask != nil && *j.SubnetMask != "" &&
		j.RangeStart != nil && *j.RangeStart != "" &&
		j.RangeEnd != nil && *j.RangeEnd != ""
}

func (j DhcpConfigV6) isValid() bool {
	return j.RangeStart != nil && *j.RangeStart != ""
}

type DhcpStaticLeases []DhcpStaticLease

// MergeDhcpStaticLeases the leases
func MergeDhcpStaticLeases(l *[]DhcpStaticLease, other *[]DhcpStaticLease) (DhcpStaticLeases, DhcpStaticLeases) {
	var thisLeases []DhcpStaticLease
	var otherLeases []DhcpStaticLease

	if l != nil {
		thisLeases = *l
	}
	if other != nil {
		otherLeases = *other
	}
	current := make(map[string]DhcpStaticLease)

	var adds DhcpStaticLeases
	var removes DhcpStaticLeases
	for _, le := range thisLeases {
		current[le.Mac] = le
	}

	for _, le := range otherLeases {
		if _, ok := current[le.Mac]; ok {
			delete(current, le.Mac)
		} else {
			adds = append(adds, le)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes
}

// Equals dns config equal check
func (c *DNSConfig) Equals(o *DNSConfig) bool {
	cc := c.Clone()
	oo := o.Clone()
	cc.Sort()
	oo.Sort()

	return utils.JsonEquals(cc, oo)
}

func (c *DNSConfig) Clone() *DNSConfig {
	return utils.Clone(c, &DNSConfig{})
}

// Sort sort dns config
func (c *DNSConfig) Sort() {
	if c.UpstreamDns != nil {
		sort.Strings(*c.UpstreamDns)
	}

	if c.UpstreamDns != nil {
		sort.Strings(*c.BootstrapDns)
	}

	if c.UpstreamDns != nil {
		sort.Strings(*c.LocalPtrUpstreams)
	}
}

// Equals access list equal check
func (al *AccessList) Equals(o *AccessList) bool {
	return EqualsStringSlice(al.AllowedClients, o.AllowedClients, true) &&
		EqualsStringSlice(al.DisallowedClients, o.DisallowedClients, true) &&
		EqualsStringSlice(al.BlockedHosts, o.BlockedHosts, true)
}

func EqualsStringSlice(a *[]string, b *[]string, sortIt bool) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	aa := *a
	bb := *b
	if sortIt {
		sort.Strings(aa)
		sort.Strings(bb)
	}
	if len(aa) != len(bb) {
		return false
	}
	for i, v := range aa {
		if v != bb[i] {
			return false
		}
	}
	return true
}

// Sort clients
func (cl *Client) Sort() {
	if cl.Ids != nil {
		sort.Strings(*cl.Ids)
	}
	if cl.Tags != nil {
		sort.Strings(*cl.Tags)
	}
	if cl.BlockedServices != nil {
		sort.Strings(*cl.BlockedServices)
	}
	if cl.Upstreams != nil {
		sort.Strings(*cl.Upstreams)
	}
}

// Equals Clients equal check
func (cl *Client) Equals(o *Client) bool {
	cl.Sort()
	o.Sort()

	return utils.JsonEquals(cl, o)
}

// Add ac client
func (clients *Clients) Add(cl Client) {
	if clients.Clients == nil {
		clients.Clients = &ClientsArray{cl}
	} else {
		a := append(*clients.Clients, cl)
		clients.Clients = &a
	}
}

// Merge merge Clients
func (clients *Clients) Merge(other *Clients) ([]*Client, []*Client, []*Client) {
	current := make(map[string]*Client)
	if clients.Clients != nil {
		cc := *clients.Clients
		for i := range cc {
			client := cc[i]
			current[*client.Name] = &client
		}
	}

	expected := make(map[string]*Client)
	if other.Clients != nil {
		oc := *other.Clients
		for i := range oc {
			client := oc[i]
			expected[*client.Name] = &client
		}
	}

	var adds []*Client
	var removes []*Client
	var updates []*Client

	for _, cl := range expected {
		if oc, ok := current[*cl.Name]; ok {
			if !cl.Equals(oc) {
				updates = append(updates, cl)
			}
			delete(current, *cl.Name)
		} else {
			adds = append(adds, cl)
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, updates, removes
}

// Key RewriteEntry key
func (re *RewriteEntry) Key() string {
	var d string
	var a string
	if re.Domain != nil {
		d = *re.Domain
	}
	if re.Answer != nil {
		a = *re.Answer
	}
	return fmt.Sprintf("%s#%s", d, a)
}

// RewriteEntries list of RewriteEntry
type RewriteEntries []RewriteEntry

// Merge RewriteEntries
func (rwe *RewriteEntries) Merge(other *RewriteEntries) (RewriteEntries, RewriteEntries, RewriteEntries) {
	current := make(map[string]RewriteEntry)

	var adds RewriteEntries
	var removes RewriteEntries
	var duplicates RewriteEntries
	processed := make(map[string]bool)
	for _, rr := range *rwe {
		if _, ok := processed[rr.Key()]; !ok {
			current[rr.Key()] = rr
			processed[rr.Key()] = true
		} else {
			// remove duplicate
			removes = append(removes, rr)
		}
	}

	for _, rr := range *other {
		if _, ok := current[rr.Key()]; ok {
			delete(current, rr.Key())
		} else {
			if _, ok := processed[rr.Key()]; !ok {
				adds = append(adds, rr)
				processed[rr.Key()] = true
			} else {
				//	skip duplicate
				duplicates = append(duplicates, rr)
			}
		}
	}

	for _, rr := range current {
		removes = append(removes, rr)
	}

	return adds, removes, duplicates
}

func MergeFilters(this *[]Filter, other *[]Filter) ([]Filter, []Filter, []Filter) {
	if this == nil && other == nil {
		return nil, nil, nil
	}

	current := make(map[string]*Filter)

	var adds []Filter
	var updates []Filter
	var removes []Filter
	if this != nil {
		for i := range *this {
			fi := (*this)[i]
			current[fi.Url] = &fi
		}
	}

	if other != nil {
		for i := range *other {
			rr := (*other)[i]
			if c, ok := current[rr.Url]; ok {
				if !c.Equals(&rr) {
					updates = append(updates, rr)
				}
				delete(current, rr.Url)
			} else {
				adds = append(adds, rr)
			}
		}
	}

	for _, rr := range current {
		removes = append(removes, *rr)
	}

	return adds, updates, removes
}

// Equals Filter equal check
func (f *Filter) Equals(o *Filter) bool {
	return f.Enabled == o.Enabled && f.Url == o.Url && f.Name == o.Name
}

// Equals QueryLogConfig equal check
func (qlc *QueryLogConfig) Equals(o *QueryLogConfig) bool {
	return ptrEquals(qlc.Enabled, o.Enabled) &&
		ptrEquals(qlc.AnonymizeClientIp, o.AnonymizeClientIp) &&
		qlc.Interval.Equals(o.Interval)
}

// Equals QueryLogConfigInterval equal check
func (qlc *QueryLogConfigInterval) Equals(o *QueryLogConfigInterval) bool {
	return ptrEquals(qlc, o)
}

func ptrEquals[T comparable](a *T, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	var aa T
	if a != nil {
		aa = *a
	}
	var bb T
	if b != nil {
		bb = *b
	}

	return aa == bb
}

// EnableConfig API struct
type EnableConfig struct {
	Enabled bool `json:"enabled"`
}

func (ssc *SafeSearchConfig) Equals(o *SafeSearchConfig) bool {
	return ptrEquals(ssc.Enabled, o.Enabled) &&
		ptrEquals(ssc.Bing, o.Bing) &&
		ptrEquals(ssc.Duckduckgo, o.Duckduckgo) &&
		ptrEquals(ssc.Google, o.Google) &&
		ptrEquals(ssc.Pixabay, o.Pixabay) &&
		ptrEquals(ssc.Yandex, o.Yandex) &&
		ptrEquals(ssc.Youtube, o.Youtube)
}

func (pi *ProfileInfo) Equals(o *ProfileInfo) bool {
	return pi.Language == o.Language &&
		pi.Theme == o.Theme
}

func (pi *ProfileInfo) ShouldSyncFor(o *ProfileInfo) *ProfileInfo {
	if pi.Equals(o) {
		return nil
	}
	merged := &ProfileInfo{Name: pi.Name, Language: pi.Language, Theme: pi.Theme}
	if o.Language != "" {
		merged.Language = o.Language
	}
	if o.Theme != "" {
		merged.Theme = o.Theme
	}
	if merged.Name == "" || merged.Language == "" || merged.Theme == "" || merged.Equals(pi) {
		return nil
	}
	return merged
}

func (bss *BlockedServicesSchedule) Equals(o *BlockedServicesSchedule) bool {
	return utils.JsonEquals(bss, o)
}

func (bss *BlockedServicesSchedule) ServicesString() string {
	return ArrayString(bss.Ids)
}

func ArrayString(a *[]string) string {
	if a == nil {
		return "[]"
	}
	sorted := *a
	sort.Strings(sorted)
	return fmt.Sprintf("[%s]", strings.Join(sorted, ","))
}
