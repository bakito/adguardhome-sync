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
	l.With("scrape-interval", w.cfg.API.Metrics.ScrapeInterval).Info("setup metrics")
	w.scrape()
	for range time.Tick(w.cfg.API.Metrics.ScrapeInterval) {
		w.scrape()
	}
}

func (w *worker) scrape() {
	var ims []metrics.InstanceMetrics

	ims = append(ims, w.getMetrics(w.cfg.Origin))
	for _, replica := range w.cfg.Replicas {
		ims = append(ims, w.getMetrics(replica))
	}
	metrics.Update(ims...)
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
	return
}
