package sync

import (
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/utils"
)

var (
	actionProfileInfo = func(ac *actionContext) error {
		if pro, err := ac.client.ProfileInfo(); err != nil {
			return err
		} else if merged := pro.ShouldSyncFor(ac.origin.profileInfo, ac.cfg.Features.Theme); merged != nil {
			return ac.client.SetProfileInfo(merged)
		}
		return nil
	}
	actionProtection = func(ac *actionContext) error {
		if ac.origin.status.ProtectionEnabled != ac.replicaStatus.ProtectionEnabled {
			return ac.client.ToggleProtection(ac.origin.status.ProtectionEnabled)
		}
		return nil
	}
	actionParental = func(ac *actionContext) error {
		if rp, err := ac.client.Parental(); err != nil {
			return err
		} else if ac.origin.parental != rp {
			return ac.client.ToggleParental(ac.origin.parental)
		}
		return nil
	}
	actionSafeSearchConfig = func(ac *actionContext) error {
		if ssc, err := ac.client.SafeSearchConfig(); err != nil {
			return err
		} else if !ac.origin.safeSearch.Equals(ssc) {
			return ac.client.SetSafeSearchConfig(ac.origin.safeSearch)
		}
		return nil
	}
	actionSafeBrowsing = func(ac *actionContext) error {
		if rs, err := ac.client.SafeBrowsing(); err != nil {
			return err
		} else if ac.origin.safeBrowsing != rs {
			if err = ac.client.ToggleSafeBrowsing(ac.origin.safeBrowsing); err != nil {
				return err
			}
		}
		return nil
	}
	actionQueryLogConfig = func(ac *actionContext) error {
		qlc, err := ac.client.QueryLogConfig()
		if err != nil {
			return err
		}
		if !ac.origin.queryLogConfig.Equals(qlc) {
			return ac.client.SetQueryLogConfig(ac.origin.queryLogConfig)
		}
		return nil
	}
	actionStatsConfig = func(ac *actionContext) error {
		sc, err := ac.client.StatsConfig()
		if err != nil {
			return err
		}
		if !sc.Equals(ac.origin.statsConfig) {
			return ac.client.SetStatsConfig(ac.origin.statsConfig)
		}
		return nil
	}
	actionDNSRewrites = func(ac *actionContext) error {
		replicaRewrites, err := ac.client.RewriteList()
		if err != nil {
			return err
		}

		a, r, d := replicaRewrites.Merge(ac.origin.rewrites)

		if err = ac.client.DeleteRewriteEntries(r...); err != nil {
			return err
		}
		if err = ac.client.AddRewriteEntries(a...); err != nil {
			return err
		}

		for _, dupl := range d {
			ac.rl.With("domain", dupl.Domain, "answer", dupl.Answer).Warn("Skipping duplicated rewrite from source")
		}
		return nil
	}
	actionFilters = func(ac *actionContext) error {
		rf, err := ac.client.Filtering()
		if err != nil {
			return err
		}

		if err = syncFilterType(ac.rl, ac.origin.filters.Filters, rf.Filters, false, ac.client, ac.cfg.ContinueOnError); err != nil {
			return err
		}
		if err = syncFilterType(ac.rl, ac.origin.filters.WhitelistFilters, rf.WhitelistFilters, true, ac.client, ac.cfg.ContinueOnError); err != nil {
			return err
		}

		if utils.PtrToString(ac.origin.filters.UserRules) != utils.PtrToString(rf.UserRules) {
			return ac.client.SetCustomRules(ac.origin.filters.UserRules)
		}

		if !utils.PtrEquals(ac.origin.filters.Enabled, rf.Enabled) ||
			!utils.PtrEquals(ac.origin.filters.Interval, rf.Interval) {
			return ac.client.ToggleFiltering(*ac.origin.filters.Enabled, *ac.origin.filters.Interval)
		}
		return nil
	}

	actionBlockedServicesSchedule = func(ac *actionContext) error {
		rbss, err := ac.client.BlockedServicesSchedule()
		if err != nil {
			return err
		}

		if !ac.origin.blockedServicesSchedule.Equals(rbss) {
			return ac.client.SetBlockedServicesSchedule(ac.origin.blockedServicesSchedule)
		}
		return nil
	}
	actionClientSettings = func(ac *actionContext) error {
		rc, err := ac.client.Clients()
		if err != nil {
			return err
		}

		a, u, r := rc.Merge(ac.origin.clients)

		for _, client := range r {
			if err := ac.client.DeleteClient(client); err != nil {
				ac.rl.With("client-name", client.Name, "error", err).Error("error deleting client setting")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}

		for _, client := range a {
			if err := ac.client.AddClient(client); err != nil {
				ac.rl.With("client-name", client.Name, "error", err).Error("error adding client setting")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}

		for _, client := range u {
			if err := ac.client.UpdateClient(client); err != nil {
				ac.rl.With("client-name", client.Name, "error", err).Error("error updating client setting")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}

		return nil
	}

	actionDNSAccessLists = func(ac *actionContext) error {
		al, err := ac.client.AccessList()
		if err != nil {
			return err
		}
		if !al.Equals(ac.origin.accessList) {
			return ac.client.SetAccessList(ac.origin.accessList)
		}
		return nil
	}
	actionDNSServerConfig = func(ac *actionContext) error {
		dc, err := ac.client.DNSConfig()
		if err != nil {
			return err
		}

		// dc.Sanitize(ac.rl)

		if !dc.Equals(ac.origin.dnsConfig) {
			if err = ac.client.SetDNSConfig(ac.origin.dnsConfig); err != nil {
				return err
			}
		}
		return nil
	}
	actionDHCPServerConfig = func(ac *actionContext) error {
		if ac.origin.dhcpServerConfig.HasConfig() {
			sc, err := ac.client.DhcpConfig()
			if err != nil {
				return err
			}
			origClone := ac.origin.dhcpServerConfig.Clone()
			if ac.replica.InterfaceName != "" {
				// overwrite interface name
				origClone.InterfaceName = utils.Ptr(ac.replica.InterfaceName)
			}
			if ac.replica.DHCPServerEnabled != nil {
				// overwrite dhcp enabled
				origClone.Enabled = ac.replica.DHCPServerEnabled
			}

			if !sc.CleanAndEquals(origClone) {
				return ac.client.SetDhcpConfig(origClone)
			}
		}
		return nil
	}
	actionDHCPStaticLeases = func(ac *actionContext) error {
		sc, err := ac.client.DhcpConfig()
		if err != nil {
			return err
		}

		a, r := model.MergeDhcpStaticLeases(sc.StaticLeases, ac.origin.dhcpServerConfig.StaticLeases)

		for _, lease := range r {
			if err := ac.client.DeleteDHCPStaticLease(lease); err != nil {
				ac.rl.With("hostname", lease.Hostname, "error", err).Error("error deleting dhcp static lease")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}

		for _, lease := range a {
			if err := ac.client.AddDHCPStaticLease(lease); err != nil {
				ac.rl.With("hostname", lease.Hostname, "error", err).Error("error adding dhcp static lease")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}
		return nil
	}
	tlsConfig = func(ac *actionContext) error {
		tlsc, err := ac.client.TLSConfig()
		if err != nil {
			return err
		}

		if !tlsc.Equals(ac.origin.tlsConfig) {
			if err := ac.client.SetTLSConfig(ac.origin.tlsConfig); err != nil {
				ac.rl.With("enabled", ac.origin.tlsConfig.Enabled, "error", err).Error("error setting tls config")
				if !ac.cfg.ContinueOnError {
					return err
				}
			}
		}
		return nil
	}
)

func syncFilterType(
	rl *zap.SugaredLogger,
	of *[]model.Filter,
	rFilters *[]model.Filter,
	whitelist bool,
	replica client.Client,
	continueOnError bool,
) error {
	fa, fu, fd := model.MergeFilters(rFilters, of)

	for _, f := range fd {
		if err := replica.DeleteFilter(whitelist, f); err != nil {
			rl.With("filter", f.Name, "url", f.Url, "whitelist", whitelist, "error", err).Error("error deleting filter")
			if !continueOnError {
				return err
			}
		}
	}

	for _, f := range fa {
		if err := replica.AddFilter(whitelist, f); err != nil {
			rl.With("filter", f.Name, "url", f.Url, "whitelist", whitelist, "error", err).Error("error adding filter")
			if !continueOnError {
				return err
			}
		}
	}

	for _, f := range fu {
		if err := replica.UpdateFilter(whitelist, f); err != nil {
			rl.With("filter", f.Name, "url", f.Url, "whitelist", whitelist, "error", err).Error("error updating filter")
			if !continueOnError {
				return err
			}
		}
	}

	if len(fa) > 0 || len(fu) > 0 {
		if err := replica.RefreshFilters(whitelist); err != nil {
			return err
		}
	}
	return nil
}
