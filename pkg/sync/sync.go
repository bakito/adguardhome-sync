package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/log"
)

// Sync config from origin to replica
func Sync(origin client.Client, replica client.Client) error {

	l := log.GetLogger("sync").With("from", origin.Host(), "to", replica.Host())
	l.Info("Start sync")

	os, err := origin.Status()
	if err != nil {
		return err
	}

	rs, err := replica.Status()
	if err != nil {
		return err
	}

	if os.Version != rs.Version {
		panic("Versions do not match")
	}

	err = syncRewrites(origin, replica)
	if err != nil {
		return err
	}
	err = syncFilters(origin, replica)
	if err != nil {
		return err
	}

	err = syncServices(origin, replica)
	if err != nil {
		return err
	}

	if err = syncClients(origin, replica); err != nil {
		return err
	}

	l.Info("Sync done")
	return nil
}

func syncServices(origin client.Client, replica client.Client) error {
	os, err := origin.Services()
	if err != nil {
		return err
	}
	rs, err := replica.Services()
	if err != nil {
		return err
	}

	if !os.Equals(rs) {
		if err := replica.SetServices(os); err != nil {
			return err
		}
	}
	return nil
}

func syncFilters(origin client.Client, replica client.Client) error {
	of, err := origin.Filtering()
	if err != nil {
		return err
	}
	rf, err := replica.Filtering()
	if err != nil {
		return err
	}

	fa, fd := rf.Filters.Merge(of.Filters)

	if err = replica.AddFilters(false, fa...); err != nil {
		return err
	}

	if len(fa) > 0 {
		if err = replica.RefreshFilters(false); err != nil {
			return err
		}
	}

	if err = replica.DeleteFilters(false, fd...); err != nil {
		return err
	}

	fa, fd = rf.WhitelistFilters.Merge(of.WhitelistFilters)
	if err = replica.AddFilters(true, fa...); err != nil {
		return err
	}

	if len(fa) > 0 {
		if err = replica.RefreshFilters(true); err != nil {
			return err
		}
	}

	if err = replica.DeleteFilters(true, fd...); err != nil {
		return err
	}

	if of.UserRules.String() != rf.UserRules.String() {
		return replica.SetCustomRules(of.UserRules)
	}

	if of.Enabled != rf.Enabled || of.Interval != rf.Interval {
		if err = replica.ToggleFiltering(of.Enabled, of.Interval); err != nil {
			return err
		}
	}
	return nil
}

func syncRewrites(origin client.Client, replica client.Client) error {
	originRewrites, err := origin.RewriteList()
	if err != nil {
		return err
	}
	replicaRewrites, err := replica.RewriteList()
	if err != nil {
		return err
	}

	a, r := replicaRewrites.Merge(originRewrites)

	if err = replica.AddRewriteEntries(a...); err != nil {
		return err
	}
	if err = replica.DeleteRewriteEntries(r...); err != nil {
		return err
	}
	return nil
}

func syncClients(origin client.Client, replica client.Client) error {
	oc, err := origin.Clients()
	if err != nil {
		return err
	}
	rc, err := replica.Clients()
	if err != nil {
		return err
	}

	a, u, r := rc.Merge(oc)

	if err = replica.AddClients(a...); err != nil {
		return err
	}
	if err = replica.UpdateClients(u...); err != nil {
		return err
	}
	if err = replica.DeleteClients(r...); err != nil {
		return err
	}
	return nil
}
