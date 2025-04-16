package metrics

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/bakito/adguardhome-sync/pkg/client/model"
	"github.com/bakito/adguardhome-sync/pkg/log"
)

const StatsTotal = "total"

var (
	l = log.GetLogger("metrics")

	// avgProcessingTime - Average processing time for a DNS query.
	avgProcessingTime = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "avg_processing_time",
			Namespace: "adguard",
			Help:      "This represent the average processing time for a DNS query in s",
		},
		[]string{"hostname"},
	)

	// dnsQueries - Number of DNS queries.
	dnsQueries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_dns_queries",
			Namespace: "adguard",
			Help:      "Number of DNS queries",
		},
		[]string{"hostname"},
	)

	// blockedFiltering - Number of DNS queries blocked.
	blockedFiltering = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_blocked_filtering",
			Namespace: "adguard",
			Help:      "This represent the number of domains blocked",
		},
		[]string{"hostname"},
	)

	// parentalFiltering - Number of DNS queries replaced by parental control.
	parentalFiltering = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_replaced_parental",
			Namespace: "adguard",
			Help:      "This represent the number of domains blocked (parental)",
		},
		[]string{"hostname"},
	)

	// safeBrowsingFiltering - Number of DNS queries replaced by safe browsing.
	safeBrowsingFiltering = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_replaced_safebrowsing",
			Namespace: "adguard",
			Help:      "This represent the number of domains blocked (safe browsing)",
		},
		[]string{"hostname"},
	)

	// safeSearchFiltering - Number of DNS queries replaced by safe search.
	safeSearchFiltering = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "num_replaced_safesearch",
			Namespace: "adguard",
			Help:      "This represent the number of domains blocked (safe search)",
		},
		[]string{"hostname"},
	)

	// topQueries - The number of top queries.
	topQueries = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_queried_domains",
			Namespace: "adguard",
			Help:      "This represent the top queried domains",
		},
		[]string{"hostname", "domain"},
	)

	// topBlocked - The number of top domains blocked.
	topBlocked = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_blocked_domains",
			Namespace: "adguard",
			Help:      "This represent the top bloacked domains",
		},
		[]string{"hostname", "domain"},
	)

	// topClients - The number of top clients.
	topClients = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "top_clients",
			Namespace: "adguard",
			Help:      "This represent the top clients",
		},
		[]string{"hostname", "client"},
	)

	// queryTypes - The type of DNS Queries (A, AAAA...)
	queryTypes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "query_types",
			Namespace: "adguard",
			Help:      "This represent the DNS query types",
		},
		[]string{"hostname", "type"},
	)

	// running - If Adguard is running.
	running = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "running",
			Namespace: "adguard",
			Help:      "This represent if Adguard is running",
		},
		[]string{"hostname"},
	)

	// protectionEnabled - If Adguard protection is enabled.
	protectionEnabled = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "protection_enabled",
			Namespace: "adguard",
			Help:      "This represent if Adguard Protection is enabled",
		},
		[]string{"hostname"},
	)
	// aghsSyncDuration - the sync curation in seconds.
	aghsSyncDuration = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "sync_duration_seconds",
			Namespace: "adguard_home_sync",
			Help:      "This represents the duration of the last sync in seconds",
		},
		[]string{"hostname"},
	)
	// aghsSyncSuccessful - the sync result.
	aghsSyncSuccessful = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:      "sync_successful",
			Namespace: "adguard_home_sync",
			Help:      "This represents the whether the last sync was successful",
		},
		[]string{"hostname"},
	)
	stats = OverallStats{}
)

// Init initializes all Prometheus metrics made available by AdGuard  exporter.
func Init() {
	initMetric("avg_processing_time", avgProcessingTime)
	initMetric("num_dns_queries", dnsQueries)
	initMetric("num_blocked_filtering", blockedFiltering)
	initMetric("num_replaced_parental", parentalFiltering)
	initMetric("num_replaced_safebrowsing", safeBrowsingFiltering)
	initMetric("num_replaced_safesearch", safeSearchFiltering)
	initMetric("top_queried_domains", topQueries)
	initMetric("top_blocked_domains", topBlocked)
	initMetric("top_clients", topClients)
	initMetric("query_types", queryTypes)
	initMetric("running", running)
	initMetric("protection_enabled", protectionEnabled)
	initMetric("sync_duration_seconds", aghsSyncDuration)
	initMetric("sync_successful", aghsSyncSuccessful)
}

func initMetric(name string, metric *prometheus.GaugeVec) {
	prometheus.MustRegister(metric)
	l.With("name", name).Info("New Prometheus metric registered")
}

func UpdateInstances(iml InstanceMetricsList) {
	for _, im := range iml.Metrics {
		updateMetrics(im)
		stats[im.HostName] = im.Stats
	}

	l.Debug("updated")
}

func UpdateResult(host string, ok bool, duration float64) {
	if ok {
		aghsSyncSuccessful.WithLabelValues(host).Set(1)
	} else {
		aghsSyncSuccessful.WithLabelValues(host).Set(0)
	}
	aghsSyncDuration.WithLabelValues(host).Set(duration)
}

func updateMetrics(im InstanceMetrics) {
	// Status
	isRunning := 0
	if im.Status.Running {
		isRunning = 1
	}
	running.WithLabelValues(im.HostName).Set(float64(isRunning))

	isProtected := 0
	if im.Status.ProtectionEnabled {
		isProtected = 1
	}
	protectionEnabled.WithLabelValues(im.HostName).Set(float64(isProtected))

	// Stats
	avgProcessingTime.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.AvgProcessingTime))
	dnsQueries.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.NumDnsQueries))
	blockedFiltering.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.NumBlockedFiltering))
	parentalFiltering.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.NumReplacedParental))
	safeBrowsingFiltering.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.NumReplacedSafebrowsing))
	safeSearchFiltering.WithLabelValues(im.HostName).Set(safeMetric(im.Stats.NumReplacedSafesearch))

	if im.Stats.TopQueriedDomains != nil {
		for _, tq := range *im.Stats.TopQueriedDomains {
			for domain, value := range tq.AdditionalProperties {
				topQueries.WithLabelValues(im.HostName, domain).Set(float64(value))
			}
		}
	}
	if im.Stats.TopBlockedDomains != nil {
		for _, tb := range *im.Stats.TopBlockedDomains {
			for domain, value := range tb.AdditionalProperties {
				topBlocked.WithLabelValues(im.HostName, domain).Set(float64(value))
			}
		}
	}
	if im.Stats.TopClients != nil {
		for _, tc := range *im.Stats.TopClients {
			for source, value := range tc.AdditionalProperties {
				topClients.WithLabelValues(im.HostName, source).Set(float64(value))
			}
		}
	}

	// LogQuery
	m := make(map[string]int)
	if im.QueryLog != nil && im.QueryLog.Data != nil {
		logdata := *im.QueryLog.Data
		for _, ld := range logdata {
			if ld.Answer != nil {
				dnsanswer := *ld.Answer
				if len(dnsanswer) > 0 {
					for _, dnsa := range dnsanswer {
						dnsType := *dnsa.Type
						m[dnsType]++
					}
				}
			}
		}
	}

	for key, value := range m {
		queryTypes.WithLabelValues(im.HostName, key).Set(float64(value))
	}
}

type InstanceMetricsList struct {
	Metrics []InstanceMetrics `faker:"slice_len=5"`
}

type InstanceMetrics struct {
	HostName string
	Status   *model.ServerStatus
	Stats    *model.Stats
	QueryLog *model.QueryLog
}

type OverallStats map[string]*model.Stats

func (os OverallStats) consolidate() OverallStats {
	consolidated := OverallStats{StatsTotal: model.NewStats()}
	for host, stats := range os {
		consolidated[host] = stats
		consolidated[StatsTotal].Add(stats)
	}
	return consolidated
}

func safeMetric[T int | float64 | float32](v *T) float64 {
	if v == nil {
		return 0
	}
	return float64(*v)
}

func getStats() OverallStats {
	return stats.consolidate()
}

func (os OverallStats) Total() *model.Stats {
	return os[StatsTotal]
}
