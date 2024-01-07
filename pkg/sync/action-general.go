package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/utils"
)

var (
	actionProfileInfo = func(ac *actionContext) error {
		if pro, err := ac.client.ProfileInfo(); err != nil {
			return err
		} else if merged := pro.ShouldSyncFor(ac.o.profileInfo); merged != nil {
			return ac.client.SetProfileInfo(merged)
		}
		return nil
	}
	actionProtection = func(ac *actionContext) error {
		if ac.o.status.ProtectionEnabled != ac.rs.ProtectionEnabled {
			return ac.client.ToggleProtection(ac.o.status.ProtectionEnabled)
		}
		return nil
	}
	actionParental = func(ac *actionContext) error {
		if rp, err := ac.client.Parental(); err != nil {
			return err
		} else if ac.o.parental != rp {
			return ac.client.ToggleParental(ac.o.parental)
		}
		return nil
	}
	actionSafeSearchConfig = func(ac *actionContext) error {
		if ssc, err := ac.client.SafeSearchConfig(); err != nil {
			return err
		} else if !ac.o.safeSearch.Equals(ssc) {
			return ac.client.SetSafeSearchConfig(ac.o.safeSearch)
		}
		return nil
	}
	actionSafeBrowsing = func(ac *actionContext) error {
		if rs, err := ac.client.SafeBrowsing(); err != nil {
			return err
		} else if ac.o.safeBrowsing != rs {
			if err = ac.client.ToggleSafeBrowsing(ac.o.safeBrowsing); err != nil {
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
		if !ac.o.queryLogConfig.Equals(qlc) {
			return ac.client.SetQueryLogConfig(ac.o.queryLogConfig)
		}
		return nil
	}
	actionStatsConfig = func(ac *actionContext) error {
		sc, err := ac.client.StatsConfig()
		if err != nil {
			return err
		}
		if ac.o.statsConfig.Interval != sc.Interval {
			return ac.client.SetStatsConfig(ac.o.statsConfig)
		}
		return nil
	}
	dnsRewrites = func(ac *actionContext) error {
		replicaRewrites, err := ac.client.RewriteList()
		if err != nil {
			return err
		}

		a, r, d := replicaRewrites.Merge(ac.o.rewrites)

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
	filters = func(ac *actionContext) error {
		rf, err := ac.client.Filtering()
		if err != nil {
			return err
		}

		if err = syncFilterType(ac.o.filters.Filters, rf.Filters, false, ac.client); err != nil {
			return err
		}
		if err = syncFilterType(ac.o.filters.WhitelistFilters, rf.WhitelistFilters, true, ac.client); err != nil {
			return err
		}

		if utils.PtrToString(ac.o.filters.UserRules) != utils.PtrToString(rf.UserRules) {
			return ac.client.SetCustomRules(ac.o.filters.UserRules)
		}

		if ac.o.filters.Enabled != rf.Enabled || ac.o.filters.Interval != rf.Interval {
			if err = ac.client.ToggleFiltering(*ac.o.filters.Enabled, *ac.o.filters.Interval); err != nil {
				return err
			}
		}
		return nil
	}
)

func syncFilterType(of *[]model.Filter, rFilters *[]model.Filter, whitelist bool, replica client.Client) error {
	fa, fu, fd := model.MergeFilters(rFilters, of)

	if err := replica.DeleteFilters(whitelist, fd...); err != nil {
		return err
	}
	if err := replica.AddFilters(whitelist, fa...); err != nil {
		return err
	}
	if err := replica.UpdateFilters(whitelist, fu...); err != nil {
		return err
	}

	if len(fa) > 0 || len(fu) > 0 {
		if err := replica.RefreshFilters(whitelist); err != nil {
			return err
		}
	}
	return nil
}
