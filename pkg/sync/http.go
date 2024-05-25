package sync

import (
	"context"
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

	"github.com/bakito/adguardhome-sync/pkg/log"
	"github.com/bakito/adguardhome-sync/pkg/metrics"
	"github.com/bakito/adguardhome-sync/version"
	"github.com/gin-gonic/gin"
)

var (
	//go:embed index.html
	index []byte
	//go:embed favicon.ico
	favicon []byte
)

func (w *worker) handleSync(c *gin.Context) {
	l.With("remote-addr", c.Request.RemoteAddr).Info("Starting sync from API")
	w.sync()
}

func (w *worker) handleRoot(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", map[string]interface{}{
		"DarkMode":   w.cfg.API.DarkMode,
		"Version":    version.Version,
		"Build":      version.Build,
		"SyncStatus": w.status(),
	},
	)
}

func (w *worker) handleFavicon(c *gin.Context) {
	c.Data(http.StatusOK, "image/x-icon", favicon)
}

func (w *worker) handleLogs(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain", []byte(strings.Join(log.Logs(), "")))
}

func (w *worker) handleClearLogs(c *gin.Context) {
	log.Clear()
	c.Status(http.StatusOK)
}

func (w *worker) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, w.status())
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
	if w.cfg.API.Username != "" && w.cfg.API.Password != "" {
		r.Use(gin.BasicAuth(map[string]string{w.cfg.API.Username: w.cfg.API.Password}))
	}
	httpServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", w.cfg.API.Port),
		Handler:           r,
		BaseContext:       func(_ net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: 1 * time.Second,
	}

	r.SetHTMLTemplate(template.Must(template.New("index.html").Parse(string(index))))
	r.POST("/api/v1/sync", w.handleSync)
	r.GET("/api/v1/logs", w.handleLogs)
	r.POST("/api/v1/clear-logs", w.handleClearLogs)
	r.GET("/api/v1/status", w.handleStatus)
	r.GET("/favicon.ico", w.handleFavicon)
	r.GET("/", w.handleRoot)
	if w.cfg.API.Metrics.Enabled {
		r.GET("/metrics", metrics.Handler())

		go w.startScraping()
	}

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

	defer os.Exit(0)
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
