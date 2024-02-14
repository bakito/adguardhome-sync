package sync

import (
	"time"

	"github.com/bakito/adguardhome-sync/pkg/metrics"
	"github.com/bakito/adguardhome-sync/pkg/types"
)

func (w *worker) startScraping() {
	if w.cfg.API.Metrics.ScrapeInterval == 0 {
		w.cfg.API.Metrics.ScrapeInterval = 30 * time.Second
	}
	for range time.Tick(w.cfg.API.Metrics.ScrapeInterval) {
		var ims []metrics.InstanceMetrics

		ims = append(ims, w.getMetrics(w.cfg.Origin))
		for _, replica := range w.cfg.Replicas {
			ims = append(ims, w.getMetrics(replica))
		}
		metrics.Update(ims...)
	}
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
