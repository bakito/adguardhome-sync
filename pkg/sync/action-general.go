package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
)

var (
	actionProfileInfo = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		if pro, err := client.ProfileInfo(); err != nil {
			return err
		} else if merged := pro.ShouldSyncFor(o.profileInfo); merged != nil {
			return client.SetProfileInfo(merged)
		}
		return nil
	}
	actionProtection = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		if o.status.ProtectionEnabled != rs.ProtectionEnabled {
			return client.ToggleProtection(o.status.ProtectionEnabled)
		}
		return nil
	}
	actionParental = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		if rp, err := client.Parental(); err != nil {
			return err
		} else if o.parental != rp {
			return client.ToggleParental(o.parental)
		}
		return nil
	}
	actionSafeSearchConfig = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		if ssc, err := client.SafeSearchConfig(); err != nil {
			return err
		} else if !o.safeSearch.Equals(ssc) {
			return client.SetSafeSearchConfig(o.safeSearch)
		}
		return nil
	}
	actionSafeBrowsing = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		if rs, err := client.SafeBrowsing(); err != nil {
			return err
		} else if o.safeBrowsing != rs {
			if err = client.ToggleSafeBrowsing(o.safeBrowsing); err != nil {
				return err
			}
		}
		return nil
	}
	actionQueryLogConfig = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		qlc, err := client.QueryLogConfig()
		if err != nil {
			return err
		}
		if !o.queryLogConfig.Equals(qlc) {
			return client.SetQueryLogConfig(o.queryLogConfig)
		}
		return nil
	}
	actionStatsConfig = func(o *origin, client client.Client, rs *model.ServerStatus) error {
		sc, err := client.StatsConfig()
		if err != nil {
			return err
		}
		if o.statsConfig.Interval != sc.Interval {
			return client.SetStatsConfig(o.statsConfig)
		}
		return nil
	}
)
