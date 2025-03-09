package sync

import (
	"time"

	"github.com/bakito/adguardhome-sync/pkg/metrics"
	"github.com/bakito/adguardhome-sync/pkg/types"
)

func (w *worker) startScraping() {
	metrics.Init()
	if w.cfg.API.Metrics.ScrapeInterval == 0 {
		w.cfg.API.Metrics.ScrapeInterval = 30 * time.Second
	}
	if w.cfg.API.Metrics.QueryLogLimit == 0 {
		w.cfg.API.Metrics.QueryLogLimit = 10_000
	}
	l.With(
		"scrape-interval", w.cfg.API.Metrics.ScrapeInterval,
		"query-log-limit", w.cfg.API.Metrics.QueryLogLimit,
	).Info("setup metrics")
	w.scrape()
	for range time.Tick(w.cfg.API.Metrics.ScrapeInterval) {
		w.scrape()
	}
}

func (w *worker) scrape() {
	var iml metrics.InstanceMetricsList

	iml.Metrics = append(iml.Metrics, w.getMetrics(w.cfg.Origin))
	for _, replica := range w.cfg.Replicas {
		iml.Metrics = append(iml.Metrics, w.getMetrics(replica))
	}
	metrics.Update(iml)
}

func (w *worker) getMetrics(inst types.AdGuardInstance) (im metrics.InstanceMetrics) {
	client, err := w.createClient(inst)
	if err != nil {
		l.With("error", err, "url", w.cfg.Origin.URL).Error("Error creating origin client")
		return
	}

	im.HostName = inst.Host
	im.Status, _ = client.Status()
	im.Stats, _ = client.Stats()
	im.QueryLog, _ = client.QueryLog(w.cfg.API.Metrics.QueryLogLimit)
	return
}
