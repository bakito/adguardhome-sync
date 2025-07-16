package sync

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"

	"github.com/bakito/adguardhome-sync/pkg/client"
	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/metrics"
	"github.com/bakito/adguardhome-sync/pkg/types"
	"github.com/bakito/adguardhome-sync/pkg/utils"
	"github.com/bakito/adguardhome-sync/pkg/versions"
	"github.com/bakito/adguardhome-sync/version"
)

var (
	l                       = log.GetLogger("sync")
	fixVersionCompareRegExp = regexp.MustCompile(`[^0-9.]`)
)

// Sync config from origin to replica.
func Sync(cfg *types.Config) error {
	if cfg.Origin.URL == "" {
		return errors.New("origin URL is required")
	}

	if len(cfg.UniqueReplicas()) == 0 {
		return errors.New("no replicas configured")
	}

	l.With(
		"version", version.Version,
		"build", version.Build,
		"os", runtime.GOOS,
		"arch", runtime.GOARCH,
	).Info("AdGuardHome sync")
	cfg.Log(l)
	cfg.Features.LogDisabled(l)
	cfg.Origin.AutoSetup = false

	w := &worker{
		cfg:          cfg,
		createClient: client.New,
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
			runOnStartAsync(cfg, w)
			w.cron.Run()
		}
	}
	if cfg.API.Port != 0 {
		runOnStartAsync(cfg, w)
		w.listenAndServe()
	} else if cfg.RunOnStart {
		l.Info("Running sync on startup")
		w.sync()
	}

	return nil
}

func runOnStartAsync(cfg *types.Config, w *worker) {
	if cfg.RunOnStart {
		go func() {
			l.Info("Running sync on startup")
			w.sync()
		}()
	}
}

type worker struct {
	cfg          *types.Config
	running      bool
	cron         *cron.Cron
	createClient func(instance types.AdGuardInstance) (client.Client, error)
	actions      []syncAction
}

func (w *worker) status() *syncStatus {
	syncStatus := &syncStatus{
		Origin: w.getStatus(*w.cfg.Origin),
	}

	for _, replica := range w.cfg.Replicas {
		st := w.getStatus(replica)
		if w.running {
			st.Status = "info"
		}
		syncStatus.Replicas = append(syncStatus.Replicas, st)
	}

	sort.Slice(syncStatus.Replicas, func(i, j int) bool {
		return syncStatus.Replicas[i].Host < syncStatus.Replicas[j].Host
	})

	syncStatus.SyncRunning = w.running

	return syncStatus
}

func (w *worker) healthz() int {
	status := w.status()

	if status.Origin.Status != "success" {
		return 503
	}

	for _, replica := range status.Replicas {
		if replica.Status != "success" {
			return 503
		}
	}

	return 200
}

func (w *worker) getStatus(inst types.AdGuardInstance) replicaStatus {
	st := replicaStatus{Host: inst.WebHost, URL: inst.WebURL}

	oc, err := w.createClient(inst)
	if err != nil {
		l.With("error", err, "url", w.cfg.Origin.URL).Error("Error creating origin client")
		st.Status = "danger"
		st.Error = err.Error()
		return st
	}
	sl := l.With("from", inst.WebHost)
	status, err := oc.Status()
	if err != nil {
		if errors.Is(err, client.ErrSetupNeeded) {
			st.Status = "warning"
			st.Error = err.Error()
			return st
		}
		sl.With("error", err).Error("Error getting origin status")
		st.Status = "danger"
		st.Error = err.Error()
		return st
	}
	st.Status = "success"
	st.ProtectionEnabled = utils.Ptr(status.ProtectionEnabled)
	return st
}

func (w *worker) sync() {
	if w.running {
		l.Info("Sync already running")
		return
	}
	w.running = true
	defer func() {
		w.running = false
	}()

	oc, err := w.createClient(*w.cfg.Origin)
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
		sl.With("error", err, "version", o.status.Version).
			Errorf("Origin AdGuard Home version must be >= %s", versions.MinAgh)
		return
	}

	sl.With("version", o.status.Version).Info("Connected to origin")

	o.profileInfo, err = oc.ProfileInfo()
	if err != nil {
		sl.With("error", err).Error("Error getting profileInfo info")
		return
	}

	o.parental, err = oc.Parental()
	if err != nil {
		sl.With("error", err).Error("Error getting parental status")
		return
	}
	o.safeSearch, err = oc.SafeSearchConfig()
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

	o.blockedServicesSchedule, err = oc.BlockedServicesSchedule()
	if err != nil {
		sl.With("error", err).Error("Error getting origin blocked services schedule")
		return
	}

	o.filters, err = oc.Filtering()
	if err != nil {
		sl.With("error", err).Error("Error getting origin actionFilters")
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
		o.dhcpServerConfig, err = oc.DhcpConfig()
		if err != nil {
			sl.With("error", err).Error("Error getting dhcp server config")
			return
		}
	}

	w.actions = setupActions(w.cfg)

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
	start := time.Now()
	withError := false
	delta := time.Since(start).Seconds()
	defer func() {
		metrics.UpdateResult(rc.Host(), !withError, delta)
		doneLog := rl.With("duration", fmt.Sprintf("%vs", delta))
		if withError {
			doneLog.Error("Sync done")
		} else {
			doneLog.Info("Sync done")
		}
	}()

	replicaStatus, err := w.statusWithSetup(rl, replica, rc)
	if err != nil {
		rl.With("error", err).Error("Error getting replica status")
		withError = true
		return
	}

	replicaStatus.Version = fixVersionCompareRegExp.ReplaceAllString(replicaStatus.Version, "")
	o.status.Version = fixVersionCompareRegExp.ReplaceAllString(o.status.Version, "")

	rl.With("version", replicaStatus.Version).Info("Connected to replica")

	if versions.IsNewerThan(versions.MinAgh, replicaStatus.Version) {
		rl.With("error", err, "version", replicaStatus.Version).
			Errorf("Replica AdGuard Home version must be >= %s", versions.MinAgh)
		withError = true
		return
	}

	if o.status.Version != replicaStatus.Version {
		rl.With("originVersion", o.status.Version, "replicaVersion", replicaStatus.Version).
			Warn("Versions do not match")
	}

	ac := &actionContext{
		cfg:           w.cfg,
		rl:            rl,
		origin:        o,
		replicaStatus: replicaStatus,
		client:        rc,
		replica:       replica,
	}

	for _, action := range w.actions {
		if err := action.sync(ac); err != nil {
			rl.With("error", err).Errorf("Error syncing %s", action.name())
			withError = true
			if !w.cfg.ContinueOnError {
				return
			}
		}
	}
}

func (w *worker) statusWithSetup(
	rl *zap.SugaredLogger,
	replica types.AdGuardInstance,
	rc client.Client,
) (*model.ServerStatus, error) {
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

type origin struct {
	status                  *model.ServerStatus
	rewrites                *model.RewriteEntries
	blockedServicesSchedule *model.BlockedServicesSchedule
	filters                 *model.FilterStatus
	clients                 *model.Clients
	queryLogConfig          *model.QueryLogConfigWithIgnored
	statsConfig             *model.GetStatsConfigResponse
	accessList              *model.AccessList
	dnsConfig               *model.DNSConfig
	dhcpServerConfig        *model.DhcpStatus
	parental                bool
	safeSearch              *model.SafeSearchConfig
	profileInfo             *model.ProfileInfo
	safeBrowsing            bool
}
