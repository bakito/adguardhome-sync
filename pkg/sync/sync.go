package sync

import (
	"errors"
	"fmt"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var (
	l = log.GetLogger("sync")
)

// Sync config from origin to replica
func Sync(cfg *types.Config) error {

	if cfg.Origin.URL == "" {
		return fmt.Errorf("origin URL is required")
	}

	if len(cfg.UniqueReplicas()) == 0 {
		return fmt.Errorf("no replicas configured")
	}

	cfg.Origin.SkipAutoSetup = true

	w := &worker{
		cfg: cfg,
		createClient: func(ai types.AdGuardInstance) (client.Client, error) {
			return client.New(ai)
		},
	}
	if cfg.Cron != "" {
		w.cron = cron.New()
		cl := l.With("version", version.Version, "cron", cfg.Cron)
		_, err := w.cron.AddFunc(cfg.Cron, func() {
			w.sync()
		})
		if err != nil {
			cl.With("error", err).Error("Error during cron job setup")
			return err
		}
		cl.Info("Setup cronjob")
		if cfg.API.Port != 0 {
			w.cron.Start()
		} else {
			w.cron.Run()
		}
	} else {
		w.sync()
	}
	if cfg.API.Port != 0 {
		w.listenAndServe()
	}
	return nil
}

type worker struct {
	cfg          *types.Config
	running      bool
	cron         *cron.Cron
	createClient func(instance types.AdGuardInstance) (client.Client, error)
}

func (w *worker) sync() {
	if w.running {
		l.Info("Sync already running")
		return
	}
	w.running = true
	defer func() { w.running = false }()

	oc, err := w.createClient(w.cfg.Origin)
	if err != nil {
		l.With("error", err, "url", w.cfg.Origin.URL).Error("Error creating origin client")
		return
	}

	sl := l.With("from", oc.Host())

	o := &origin{}
	o.status, err = oc.Status()
	if err != nil {
		sl.With("error", err).Error("Error getting origin status")
		return
	}

	o.parental, err = oc.Parental()
	if err != nil {
		sl.With("error", err).Error("Error getting parental status")
		return
	}
	o.safeSearch, err = oc.SafeSearch()
	if err != nil {
		sl.With("error", err).Error("Error getting safe search status")
		return
	}
	o.safeBrowsing, err = oc.SafeBrowsing()
	if err != nil {
		sl.With("error", err).Error("Error getting safe browsing status")
		return
	}

	o.rewrites, err = oc.RewriteList()
	if err != nil {
		sl.With("error", err).Error("Error getting origin rewrites")
		return
	}

	o.services, err = oc.Services()
	if err != nil {
		sl.With("error", err).Error("Error getting origin services")
		return
	}

	o.filters, err = oc.Filtering()
	if err != nil {
		sl.With("error", err).Error("Error getting origin filters")
		return
	}
	o.clients, err = oc.Clients()
	if err != nil {
		sl.With("error", err).Error("Error getting origin clients")
		return
	}
	o.queryLogConfig, err = oc.QueryLogConfig()
	if err != nil {
		sl.With("error", err).Error("Error getting query log config")
		return
	}
	o.statsConfig, err = oc.StatsConfig()
	if err != nil {
		sl.With("error", err).Error("Error getting stats config")
		return
	}

	replicas := w.cfg.UniqueReplicas()
	for _, replica := range replicas {
		w.syncTo(sl, o, replica)
	}
}

func (w *worker) syncTo(l *zap.SugaredLogger, o *origin, replica types.AdGuardInstance) {

	rc, err := w.createClient(replica)
	if err != nil {
		l.With("error", err, "url", replica.URL).Error("Error creating replica client")
		return
	}

	rl := l.With("to", rc.Host())
	rl.Info("Start sync")

	rs, err := rc.Status()
	if err != nil {
		if !replica.SkipAutoSetup && errors.Is(err, client.SetupNeededError) {
			if err = rc.Setup(); err != nil {
				l.With("error", err).Error("Error setup adguard home")
				return
			}
			rs, err = rc.Status()
		}
		if err != nil {
			l.With("error", err).Error("Error getting replica status")
			return
		}
	}

	if o.status.Version != rs.Version {
		l.With("originVersion", o.status.Version, "replicaVersion", rs.Version).Warn("Versions do not match")
	}

	err = w.syncGeneralSettings(o, rs, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing general settings")
		return
	}

	err = w.syncConfigs(o, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing configs")
		return
	}

	err = w.syncRewrites(o.rewrites, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing rewrites")
		return
	}
	err = w.syncFilters(o.filters, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing filters")
		return
	}

	err = w.syncServices(o.services, rc)
	if err != nil {
		l.With("error", err).Error("Error syncing services")
		return
	}

	if err = w.syncClients(o.clients, rc); err != nil {
		l.With("error", err).Error("Error syncing clients")
		return
	}

	rl.Info("Sync done")
}

func (w *worker) syncServices(os types.Services, replica client.Client) error {
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

func (w *worker) syncFilters(of *types.FilteringStatus, replica client.Client) error {
	rf, err := replica.Filtering()
	if err != nil {
		return err
	}

	if err = w.syncFilterType(of.Filters, rf.Filters, false, replica); err != nil {
		return err
	}
	if err = w.syncFilterType(of.WhitelistFilters, rf.WhitelistFilters, true, replica); err != nil {
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

func (w *worker) syncFilterType(of types.Filters, rFilters types.Filters, whitelist bool, replica client.Client) error {
	fa, fu, fd := rFilters.Merge(of)

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

	if err := replica.DeleteFilters(whitelist, fd...); err != nil {
		return err
	}
	return nil
}

func (w *worker) syncRewrites(or *types.RewriteEntries, replica client.Client) error {

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

func (w *worker) syncClients(oc *types.Clients, replica client.Client) error {
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

func (w *worker) syncGeneralSettings(o *origin, rs *types.Status, replica client.Client) error {
	if o.status.ProtectionEnabled != rs.ProtectionEnabled {
		if err := replica.ToggleProtection(o.status.ProtectionEnabled); err != nil {
			return err
		}
	}
	if rp, err := replica.Parental(); err != nil {
		return err
	} else if o.parental != rp {
		if err = replica.ToggleParental(o.parental); err != nil {
			return err
		}
	}
	if rs, err := replica.SafeSearch(); err != nil {
		return err
	} else if o.safeSearch != rs {
		if err = replica.ToggleSafeSearch(o.safeSearch); err != nil {
			return err
		}
	}
	if rs, err := replica.SafeBrowsing(); err != nil {
		return err
	} else if o.safeBrowsing != rs {
		if err = replica.ToggleSafeBrowsing(o.safeBrowsing); err != nil {
			return err
		}
	}
	return nil
}

func (w *worker) syncConfigs(o *origin, replica client.Client) error {
	qlc, err := replica.QueryLogConfig()
	if err != nil {
		return err
	}
	if !o.queryLogConfig.Equals(qlc) {
		if err = replica.SetQueryLogConfig(o.queryLogConfig.Enabled, o.queryLogConfig.Interval, o.queryLogConfig.AnonymizeClientIP); err != nil {
			return err
		}
	}

	sc, err := replica.StatsConfig()
	if err != nil {
		return err
	}
	if o.statsConfig.Interval != sc.Interval {
		if err = replica.SetStatsConfig(o.statsConfig.Interval); err != nil {
			return err
		}
	}

	return nil
}

type origin struct {
	status         *types.Status
	rewrites       *types.RewriteEntries
	services       types.Services
	filters        *types.FilteringStatus
	clients        *types.Clients
	queryLogConfig *types.QueryLogConfig
	statsConfig    *types.IntervalConfig
	parental       bool
	safeSearch     bool
	safeBrowsing   bool
}
