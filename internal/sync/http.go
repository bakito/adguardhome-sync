package sync

import (
	"context"
	// go embed blank import.
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bakito/adguardhome-sync/internal/log"
	"github.com/bakito/adguardhome-sync/internal/metrics"
	"github.com/bakito/adguardhome-sync/internal/sync/static"
	"github.com/bakito/adguardhome-sync/version"
)

func (w *worker) handleSync(c *gin.Context) {
	l.With("remote-addr", c.Request.RemoteAddr).Info("Starting sync from API")
	w.sync()
}

func (w *worker) handleRoot(c *gin.Context) {
	total, dns, blocked, malware, adult := metrics.StatsGraph()

	c.HTML(http.StatusOK, "index.html", map[string]any{
		"DarkMode":   w.cfg.API.DarkMode,
		"Metrics":    w.cfg.API.Metrics.Enabled,
		"Version":    version.Version,
		"Build":      version.Build,
		"SyncStatus": w.status(),
		"Stats": map[string]any{
			"Labels":            getLast24Hours(),
			"DNS":               dns,
			"Blocked":           blocked,
			"BlockedPercentage": percent(total.NumBlockedFiltering, total.NumDnsQueries),
			"Malware":           malware,
			"MalwarePercentage": percent(total.NumReplacedSafebrowsing, total.NumDnsQueries),
			"Adult":             adult,
			"AdultPercentage":   percent(total.NumReplacedParental, total.NumDnsQueries),
			"TotalDNS":          total.NumDnsQueries,
			"TotalBlocked":      total.NumBlockedFiltering,
			"TotalMalware":      total.NumReplacedSafebrowsing,
			"TotalAdult":        total.NumReplacedParental,
		},
	},
	)
}

func percent(a, b *int) string {
	if a == nil || b == nil || *b == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", (float64(*a)*100.0)/float64(*b))
}

func (*worker) handleLogs(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain", []byte(strings.Join(log.Logs(), "")))
}

func (*worker) handleClearLogs(c *gin.Context) {
	log.Clear()
	c.Status(http.StatusOK)
}

func (w *worker) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, w.status())
}

func (w *worker) handleHealthz(c *gin.Context) {
	status := w.status()

	if status.Origin.Status != "success" {
		c.Status(http.StatusServiceUnavailable)
		return
	}

	for _, replica := range status.Replicas {
		if replica.Status != "success" {
			c.Status(http.StatusServiceUnavailable)
			return
		}
	}

	c.Status(http.StatusOK)
}

func (w *worker) listenAndServe() {
	sl := l.With("port", w.cfg.API.Port)
	if w.cfg.API.TLS.Enabled() {
		c, k := w.cfg.API.TLS.Certs()
		sl = sl.With("tls-cert", c).With("tls-key", k)
	}
	sl.Info("Starting API server")

	ctx, cancel := context.WithCancel(context.Background())

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	r.HEAD("/healthz", w.handleHealthz)
	r.GET("/healthz", w.handleHealthz)

	var group gin.IRouter = r
	if w.cfg.API.Username != "" && w.cfg.API.Password != "" {
		group = r.Group("/", gin.BasicAuth(map[string]string{w.cfg.API.Username: w.cfg.API.Password}))
	}

	group.POST("/api/v1/sync", w.handleSync)
	group.GET("/api/v1/logs", w.handleLogs)
	group.POST("/api/v1/clear-logs", w.handleClearLogs)
	group.GET("/api/v1/status", w.handleStatus)
	static.HandleResources(group, w.cfg.API.DarkMode)
	group.GET("/", w.handleRoot)
	if w.cfg.API.Metrics.Enabled {
		group.GET("/metrics", metrics.Handler())

		go w.startScraping()
	}

	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", w.cfg.API.Port),
		Handler:           r,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 1 * time.Second,
	}

	r.SetHTMLTemplate(template.Must(template.New("index.html").Parse(static.Index())))

	go func() {
		var err error
		if w.cfg.API.TLS.Enabled() {
			err = httpServer.ListenAndServeTLS(w.cfg.API.TLS.Certs())
		} else {
			err = httpServer.ListenAndServe()
		}

		if !errors.Is(err, http.ErrServerClosed) {
			l.With("error", err).Fatalf("HTTP server ListenAndServe")
		}
	}()

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,  // kill -SIGHUP XXXX
		syscall.SIGINT,  // kill -SIGINT XXXX or Ctrl+c
		syscall.SIGQUIT, // kill -SIGQUIT XXXX
	)

	<-signalChan
	l.Info("os.Interrupt - shutting down...")

	go func() {
		<-signalChan
		l.Fatal("os.Kill - terminating...")
	}()

	gracefulCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if w.cron != nil {
		l.Info("Stopping cron")
		w.cron.Stop()
	}

	if err := httpServer.Shutdown(gracefulCtx); err != nil {
		l.With("error", err).Error("Shutdown error")
		defer os.Exit(1)
	} else {
		l.Info("API server stopped")
	}

	// manually cancel context if not using httpServer.RegisterOnShutdown(cancel)
	cancel()
}

type syncStatus struct {
	SyncRunning bool            `json:"syncRunning"`
	Origin      replicaStatus   `json:"origin"`
	Replicas    []replicaStatus `json:"replicas"`
}

type replicaStatus struct {
	Host              string `json:"host"`
	URL               string `json:"url"`
	Status            string `json:"status"`
	Error             string `json:"error,omitempty"`
	ProtectionEnabled *bool  `json:"protection_enabled"`
}

func getLast24Hours() []string {
	var result []string
	currentTime := time.Now()

	// Loop to get the last 24 hours
	for i := range 24 {
		// Calculate the time for the current hour in the loop
		timeInstance := currentTime.Add(time.Duration(-i) * time.Hour)
		timeInstance = timeInstance.Truncate(time.Hour)

		// Format the time as "14 Dec 17:00"
		formattedTime := timeInstance.Format("02 Jan 15:04")
		result = append(result, formattedTime)
	}

	// Reverse the slice to get the correct order (from oldest to latest)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
