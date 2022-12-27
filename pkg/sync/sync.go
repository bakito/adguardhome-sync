package sync

import (
	"errors"
	"fmt"
	"time"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/versions"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

var l = log.GetLogger("sync")

// Sync config from origin to replica
func Sync(cfg *types.Config) error {
	if cfg.Origin.URL == "" {
		return fmt.Errorf("origin URL is required")
	}

	if len(cfg.UniqueReplicas()) == 0 {
		return fmt.Errorf("no replicas configured")
	}

	l.With("version", version.Version, "build", version.Build).Info("AdGuardHome sync")
	cfg.Features.LogDisabled(l)
	cfg.Origin.AutoSetup = false

	w := &worker{
		cfg: cfg,
		createClient: func(ai types.AdGuardInstance) (client.Client, error) {
			return client.New(ai)
		},
	}
	if cfg.Cron != "" {
		w.cron = cron.New()
		cl := l.With("cron", cfg.Cron)
		sched, err := cron.ParseStandard(cfg.Cron)
		if err != nil {
			cl.With("error", err).Error("Error parsing cron expression")
			return err
		}
		cl = cl.With("next-execution", sched.Next(time.Now()))
		_, err = w.cron.AddFunc(cfg.Cron, func() {
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
	}
	if cfg.API.Port != 0 {
		if cfg.RunOnStart {
			go func() {
				l.Info("Running sync on startup")
				w.sync()
			}()
		}
		w.listenAndServe()
	} else if cfg.RunOnStart {
		l.Info("Running sync on startup")
		w.sync()
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

	if versions.IsNewerThan(versions.MinAgh, o.status.Version) {
		sl.With("error", err, "version", o.status.Version).Errorf("Origin AdGuard Home version must be >= %s", versions.MinAgh)
		return
	}

	sl.With("version", o.status.Version).Info("Connected to origin")

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

	o.accessList, err = oc.AccessList()
	if err != nil {
		sl.With("error", err).Error("Error getting access list")
		return
	}

	o.dnsConfig, err = oc.DNSConfig()
	if err != nil {
		sl.With("error", err).Error("Error getting dns config")
		return
	}

	if w.cfg.Features.DHCP.ServerConfig || w.cfg.Features.DHCP.StaticLeases {
		o.dhcpServerConfig, err = oc.DHCPServerConfig()
		if err != nil {
			sl.With("error", err).Error("Error getting dhcp server config")
			return
		}
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

	rs, err := w.statusWithSetup(rl, replica, rc)
	if err != nil {
		rl.With("error", err).Error("Error getting replica status")
		return
	}

	rl.With("version", rs.Version).Info("Connected to replica")

	if versions.IsNewerThan(versions.MinAgh, rs.Version) {
		rl.With("error", err, "version", rs.Version).Errorf("Replica AdGuard Home version must be >= %s", versions.MinAgh)
		return
	}

	if versions.IsSame(rs.Version, versions.IncompatibleAPI) {
		rl.With("error", err, "version", rs.Version).Errorf("Replica AdGuard Home runs with an incompatible API - Please ugrade to version %s or newer", versions.FixedIncompatibleAPI)
		return
	}

	if o.status.Version != rs.Version {
		rl.With("originVersion", o.status.Version, "replicaVersion", rs.Version).Warn("Versions do not match")
	}

	err = w.syncGeneralSettings(o, rs, rc)
	if err != nil {
		rl.With("error", err).Error("Error syncing general settings")
		return
	}

	err = w.syncConfigs(o, rc)
	if err != nil {
		rl.With("error", err).Error("Error syncing configs")
		return
	}

	err = w.syncRewrites(rl, o.rewrites, rc)
	if err != nil {
		rl.With("error", err).Error("Error syncing rewrites")
		return
	}
	err = w.syncFilters(o.filters, rc)
	if err != nil {
		rl.With("error", err).Error("Error syncing filters")
		return
	}

	err = w.syncServices(o.services, rc)
	if err != nil {
		rl.With("error", err).Error("Error syncing services")
		return
	}

	if err = w.syncClients(o.clients, rc); err != nil {
		rl.With("error", err).Error("Error syncing clients")
		return
	}

	if err = w.syncDNS(o.accessList, o.dnsConfig, rc); err != nil {
		rl.With("error", err).Error("Error syncing dns")
		return
	}

	if w.cfg.Features.DHCP.ServerConfig || w.cfg.Features.DHCP.StaticLeases {
		if err = w.syncDHCPServer(o.dhcpServerConfig, rc, replica); err != nil {
			rl.With("error", err).Error("Error syncing dns")
			return
		}
	}

	rl.Info("Sync done")
}

func (w *worker) statusWithSetup(rl *zap.SugaredLogger, replica types.AdGuardInstance, rc client.Client) (*types.Status, error) {
	rs, err := rc.Status()
	if err != nil {
		if replica.AutoSetup && errors.Is(err, client.ErrSetupNeeded) {
			if serr := rc.Setup(); serr != nil {
				rl.With("error", serr).Error("Error setup AdGuardHome")
				return nil, err
			}
			return rc.Status()
		}
		return nil, err
	}
	return rs, err
}

func (w *worker) syncServices(os types.Services, replica client.Client) error {
	if w.cfg.Features.Services {
		rs, err := replica.Services()
		if err != nil {
			return err
		}

		if !os.Equals(rs) {
			if err := replica.SetServices(os); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *worker) syncFilters(of *types.FilteringStatus, replica client.Client) error {
	if w.cfg.Features.Filters {
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
	}
	return nil
}

func (w *worker) syncFilterType(of types.Filters, rFilters types.Filters, whitelist bool, replica client.Client) error {
	fa, fu, fd := rFilters.Merge(of)

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

func (w *worker) syncRewrites(rl *zap.SugaredLogger, or *types.RewriteEntries, replica client.Client) error {
	if w.cfg.Features.DNS.Rewrites {
		replicaRewrites, err := replica.RewriteList()
		if err != nil {
			return err
		}

		a, r, d := replicaRewrites.Merge(or)

		if err = replica.DeleteRewriteEntries(r...); err != nil {
			return err
		}
		if err = replica.AddRewriteEntries(a...); err != nil {
			return err
		}

		for _, dupl := range d {
			rl.With("domain", dupl.Domain, "answer", dupl.Answer).Warn("Skipping duplicated rewrite from source")
		}
	}

	return nil
}

func (w *worker) syncClients(oc *types.Clients, replica client.Client) error {
	if w.cfg.Features.ClientSettings {
		rc, err := replica.Clients()
		if err != nil {
			return err
		}

		a, u, r := rc.Merge(oc)

		if err = replica.DeleteClients(r...); err != nil {
			return err
		}
		if err = replica.AddClients(a...); err != nil {
			return err
		}
		if err = replica.UpdateClients(u...); err != nil {
			return err
		}
	}
	return nil
}

func (w *worker) syncGeneralSettings(o *origin, rs *types.Status, replica client.Client) error {
	if w.cfg.Features.GeneralSettings {
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
	}
	return nil
}

func (w *worker) syncConfigs(o *origin, rc client.Client) error {
	if w.cfg.Features.QueryLogConfig {
		qlc, err := rc.QueryLogConfig()
		if err != nil {
			return err
		}
		if !o.queryLogConfig.Equals(qlc) {
			if err = rc.SetQueryLogConfig(o.queryLogConfig.Enabled, o.queryLogConfig.Interval, o.queryLogConfig.AnonymizeClientIP); err != nil {
				return err
			}
		}
	}
	if w.cfg.Features.StatsConfig {
		sc, err := rc.StatsConfig()
		if err != nil {
			return err
		}
		if o.statsConfig.Interval != sc.Interval {
			if err = rc.SetStatsConfig(o.statsConfig.Interval); err != nil {
				return err
			}
		}
	}

	return nil
}

func (w *worker) syncDNS(oal *types.AccessList, odc *types.DNSConfig, rc client.Client) error {
	if w.cfg.Features.DNS.AccessLists {
		al, err := rc.AccessList()
		if err != nil {
			return err
		}
		if !al.Equals(oal) {
			if err = rc.SetAccessList(oal); err != nil {
				return err
			}
		}
	}
	if w.cfg.Features.DNS.ServerConfig {
		dc, err := rc.DNSConfig()
		if err != nil {
			return err
		}
		if !dc.Equals(odc) {
			if err = rc.SetDNSConfig(odc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *worker) syncDHCPServer(osc *types.DHCPServerConfig, rc client.Client, replica types.AdGuardInstance) error {
	if !w.cfg.Features.DHCP.ServerConfig && !w.cfg.Features.DHCP.StaticLeases {
		return nil
	}
	sc, err := rc.DHCPServerConfig()
	if w.cfg.Features.DHCP.ServerConfig {
		if err != nil {
			return err
		}
		origClone := osc.Clone()
		if replica.InterfaceName != "" {
			// overwrite interface name
			origClone.InterfaceName = replica.InterfaceName
		}
		if !sc.Equals(origClone) {
			if err = rc.SetDHCPServerConfig(origClone); err != nil {
				return err
			}
		}
	}

	if w.cfg.Features.DHCP.StaticLeases {
		a, r := sc.StaticLeases.Merge(osc.StaticLeases)

		if err = rc.DeleteDHCPStaticLeases(r...); err != nil {
			return err
		}
		if err = rc.AddDHCPStaticLeases(a...); err != nil {
			return err
		}
	}
	return nil
}

type origin struct {
	status           *types.Status
	rewrites         *types.RewriteEntries
	services         types.Services
	filters          *types.FilteringStatus
	clients          *types.Clients
	queryLogConfig   *types.QueryLogConfig
	statsConfig      *types.IntervalConfig
	accessList       *types.AccessList
	dnsConfig        *types.DNSConfig
	dhcpServerConfig *types.DHCPServerConfig
	parental         bool
	safeSearch       bool
	safeBrowsing     bool
}
