package sync

import (
	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"go.uber.org/zap"
)

// Sync config from origin to replica
func Sync(cfg *types.Config) {
	l := log.GetLogger("sync")
	oc, err := client.New(cfg.Origin)
	if err != nil {
		l.With("error", err, "url", cfg.Origin.URL).Error("Error creating origin client")
		return
	}

	l = l.With("from", oc.Host())

	o := &origin{}

	o.status, err = oc.Status()
	if err != nil {
		l.With("error", err).Error("Error getting origin status")
		return
	}

	o.rewrites, err = oc.RewriteList()
	if err != nil {
		l.With("error", err).Error("Error getting origin rewrites")
		return
	}

	o.services, err = oc.Services()
	if err != nil {
		l.With("error", err).Error("Error getting origin services")
		return
	}

	o.filters, err = oc.Filtering()
	if err != nil {
		l.With("error", err).Error("Error getting origin filters")
		return
	}
	o.clients, err = oc.Clients()
	if err != nil {
		l.With("error", err).Error("Error getting origin clients")
		return
	}

	replicas := cfg.UniqueReplicas()
	for _, replica := range replicas {
		syncTo(l, o, replica)
	}
}

func syncTo(l *zap.SugaredLogger, o *origin, replica types.AdGuardInstance) {

	rc, err := client.New(replica)
	if err != nil {
		l.With("error", err, "url", replica.URL).Error("Error creating replica client")
	}

	rl := l.With("to", rc.Host())
	rl.Info("Start sync")

	rs, err := rc.Status()
	if err != nil {
		l.With("error", err).Error("Error getting replica status")
		return
	}

	if o.status.Version != rs.Version {
		l.With("originVersion", o.status.Version, "replicaVersion", rs.Version).Warn("Versions do not match")
	}

	err = syncRewrites(o.rewrites, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing rewrites")
		return
	}
	err = syncFilters(o.filters, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing filters")
		return
	}

	err = syncServices(o.services, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing services")
		return
	}

	if err = syncClients(o.clients, rc); err != nil {
		l.With("error", err).Error("Error syncing clients")
		return
	}

	rl.Info("Sync done")
}

func syncServices(os *types.Services, replica client.Client) error {
	rs, err := replica.Services()
	if err != nil {
		return err
	}

	if !os.Equals(rs) {
		if err := replica.SetServices(*os); err != nil {
			return err
		}
	}
	return nil
}

func syncFilters(of *types.FilteringStatus, replica client.Client) error {
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

func syncRewrites(or *types.RewriteEntries, replica client.Client) error {

	replicaRewrites, err := replica.RewriteList()
	if err != nil {
		return err
	}

	a, r := replicaRewrites.Merge(or)

	if err = replica.AddRewriteEntries(a...); err != nil {
		return err
	}
	if err = replica.DeleteRewriteEntries(r...); err != nil {
		return err
	}
	return nil
}

func syncClients(oc *types.Clients, replica client.Client) error {
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

type origin struct {
	status   *types.Status
	rewrites *types.RewriteEntries
	services *types.Services
	filters  *types.FilteringStatus
	clients  *types.Clients
}
