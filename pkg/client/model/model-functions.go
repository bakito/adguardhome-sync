package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"

	"github.com/bakito/adguardhome-sync/pkg/utils"
)

// Clone the config.
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

// CleanAndEquals dhcp server config equal check where V4 and V6 are cleaned in advance.
func (c *DhcpStatus) CleanAndEquals(o *DhcpStatus) bool {
	c.cleanV4V6()
	o.cleanV4V6()
	return c.Equals(o)
}

// Equals dhcp server config equal check.
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

// MergeDhcpStaticLeases the leases.
func MergeDhcpStaticLeases(l, other *[]DhcpStaticLease) (DhcpStaticLeases, DhcpStaticLeases) {
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

// Equals dns config equal check.
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

// Sort sort dns config.
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

// Equals access list equal check.
func (al *AccessList) Equals(o *AccessList) bool {
	return EqualsStringSlice(al.AllowedClients, o.AllowedClients, true) &&
		EqualsStringSlice(al.DisallowedClients, o.DisallowedClients, true) &&
		EqualsStringSlice(al.BlockedHosts, o.BlockedHosts, true)
}

func EqualsStringSlice(a, b *[]string, sortIt bool) bool {
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

// Sort clients.
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

// PrepareDiff so we skip it in diff.
func (cl *Client) PrepareDiff() *string {
	var tz *string
	bss := cl.BlockedServicesSchedule
	if bss != nil && bss.Mon == nil && bss.Tue == nil && bss.Wed == nil &&
		bss.Thu == nil && bss.Fri == nil && bss.Sat == nil && bss.Sun == nil {
		tz = cl.BlockedServicesSchedule.TimeZone
		cl.BlockedServicesSchedule.TimeZone = nil
	}
	return tz
}

// AfterDiff reset after diff.
func (cl *Client) AfterDiff(tz *string) {
	if cl.BlockedServicesSchedule != nil {
		cl.BlockedServicesSchedule.TimeZone = tz
	}
}

// Equals Clients equal check.
func (cl *Client) Equals(o *Client) bool {
	cl.Sort()
	o.Sort()

	bssCl := cl.PrepareDiff()
	bssO := o.PrepareDiff()

	defer func() {
		cl.AfterDiff(bssCl)
		o.AfterDiff(bssO)
	}()

	return utils.JsonEquals(cl, o)
}

// Add ac client.
func (clients *Clients) Add(cl Client) {
	if clients.Clients == nil {
		clients.Clients = &ClientsArray{cl}
	} else {
		a := append(*clients.Clients, cl)
		clients.Clients = &a
	}
}

// Merge merge Clients.
func (clients *Clients) Merge(other *Clients) ([]*Client, []*Client, []*Client) {
	current := make(map[string]*Client)
	if clients.Clients != nil {
		cc := *clients.Clients
		for _, client := range cc {
			current[*client.Name] = &client
		}
	}

	expected := make(map[string]*Client)
	if other.Clients != nil {
		oc := *other.Clients
		for _, client := range oc {
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

// Key RewriteEntry key.
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

// RewriteEntries list of RewriteEntry.
type RewriteEntries []RewriteEntry

// Merge RewriteEntries.
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

func MergeFilters(this, other *[]Filter) ([]Filter, []Filter, []Filter) {
	if this == nil && other == nil {
		return nil, nil, nil
	}

	current := make(map[string]*Filter)

	var adds []Filter
	var updates []Filter
	var removes []Filter
	if this != nil {
		for _, fi := range *this {
			current[fi.Url] = &fi
		}
	}

	if other != nil {
		for _, rr := range *other {
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

// Equals Filter equal check.
func (f *Filter) Equals(o *Filter) bool {
	return f.Enabled == o.Enabled && f.Url == o.Url && f.Name == o.Name
}

type QueryLogConfigWithIgnored struct {
	QueryLogConfig

	// Ignored List of host names, which should not be written to log
	Ignored []string `json:"ignored,omitempty"`
}

// Equals QueryLogConfig equal check.
func (qlc *QueryLogConfigWithIgnored) Equals(o *QueryLogConfigWithIgnored) bool {
	return utils.JsonEquals(qlc, o)
}

// Equals QueryLogConfigInterval equal check.
func (qlc *QueryLogConfigInterval) Equals(o *QueryLogConfigInterval) bool {
	return ptrEquals(qlc, o)
}

func ptrEquals[T comparable](a, b *T) bool {
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

// EnableConfig API struct.
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

func (pi *ProfileInfo) Equals(o *ProfileInfo, withTheme bool) bool {
	return pi.Language == o.Language && (!withTheme || pi.Theme == o.Theme)
}

func (pi *ProfileInfo) ShouldSyncFor(o *ProfileInfo, withTheme bool) *ProfileInfo {
	if pi.Equals(o, withTheme) {
		return nil
	}
	merged := &ProfileInfo{Name: pi.Name, Language: pi.Language, Theme: pi.Theme}
	if o.Language != "" {
		merged.Language = o.Language
	}
	if withTheme && o.Theme != "" {
		merged.Theme = o.Theme
	}
	if merged.Name == "" || merged.Language == "" || merged.Equals(pi, false) {
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

func (c *DNSConfig) Sanitize(l *zap.SugaredLogger) {
	// disable UsePrivatePtrResolvers if not configured
	// https://github.com/AdguardTeam/AdGuardHome/issues/6820
	if c.UsePrivatePtrResolvers != nil && *c.UsePrivatePtrResolvers &&
		(c.LocalPtrUpstreams == nil || len(*c.LocalPtrUpstreams) == 0) {
		l.Warn(
			"disabling replica 'Use private reverse DNS resolvers' as no 'Private reverse DNS servers' are configured on origin",
		)
		c.UsePrivatePtrResolvers = utils.Ptr(false)
	}
}

// Equals GetStatsConfigResponse equal check.
func (sc *GetStatsConfigResponse) Equals(o *GetStatsConfigResponse) bool {
	return utils.JsonEquals(sc, o)
}

func NewStats() *Stats {
	return &Stats{
		NumBlockedFiltering:     ptr.To(0),
		NumReplacedParental:     ptr.To(0),
		NumReplacedSafesearch:   ptr.To(0),
		NumReplacedSafebrowsing: ptr.To(0),
		NumDnsQueries:           ptr.To(0),

		BlockedFiltering:     ptr.To(make([]int, 24)),
		DnsQueries:           ptr.To(make([]int, 24)),
		ReplacedParental:     ptr.To(make([]int, 24)),
		ReplacedSafebrowsing: ptr.To(make([]int, 24)),
	}
}

func (s *Stats) Add(other *Stats) {
	s.NumBlockedFiltering = addInt(s.NumBlockedFiltering, other.NumBlockedFiltering)
	s.NumReplacedSafebrowsing = addInt(s.NumReplacedSafebrowsing, other.NumReplacedSafebrowsing)
	s.NumDnsQueries = addInt(s.NumDnsQueries, other.NumDnsQueries)
	s.NumReplacedSafesearch = addInt(s.NumReplacedSafesearch, other.NumReplacedSafesearch)
	s.NumReplacedParental = addInt(s.NumReplacedParental, other.NumReplacedParental)

	s.BlockedFiltering = sumUp(s.BlockedFiltering, other.BlockedFiltering)
	s.DnsQueries = sumUp(s.DnsQueries, other.DnsQueries)
	s.ReplacedParental = sumUp(s.ReplacedParental, other.ReplacedParental)
	s.ReplacedSafebrowsing = sumUp(s.ReplacedSafebrowsing, other.ReplacedSafebrowsing)
}

func addInt(t, add *int) *int {
	if add != nil {
		return ptr.To(*t + *add)
	}
	return t
}

func sumUp(t, o *[]int) *[]int {
	if o != nil {
		tt := *t
		oo := *o
		var sum []int
		for i := range tt {
			if len(oo) >= i {
				sum = append(sum, tt[i]+oo[i])
			}
		}
		return &sum
	}
	return t
}
