package sync

import (
	// import embed for html page
	_ "embed"

	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bakito/adguardhome-sync/pkg/log"
)

var (
	//go:embed index.html
	index []byte
	//go:embed favicon.ico
	favicon []byte
)

func (w *worker) handleSync(rw http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		l.With("remote-addr", req.RemoteAddr).Info("Starting sync from API")
		w.sync()
	default:
		http.Error(rw, "only POST allowed", http.StatusBadRequest)
	}
}

func (w *worker) handleRoot(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "text/html")
	_, _ = rw.Write(index)
}

func (w *worker) handleFavicon(rw http.ResponseWriter, _ *http.Request) {
	rw.Header().Set("Content-Type", "image/x-icon")
	_, _ = rw.Write(favicon)
}

func (w *worker) handleLogs(rw http.ResponseWriter, _ *http.Request) {
	_, _ = rw.Write([]byte(strings.Join(log.Logs(), "")))
}

func (w *worker) basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {

		rw.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		username, password, authOK := r.BasicAuth()
		if !authOK {
			http.Error(rw, "Not authorized", 401)
			return
		}

		if username != w.cfg.API.Username || password != w.cfg.API.Password {
			http.Error(rw, "Not authorized", 401)
			return
		}

		h.ServeHTTP(rw, r)
	}
}

func use(h http.HandlerFunc, middleware ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, m := range middleware {
		h = m(h)
	}

	return h
}

func (w *worker) listenAndServe() {
	l.With("port", w.cfg.API.Port).Info("Starting API server")

	ctx, cancel := context.WithCancel(context.Background())
	mux := http.NewServeMux()
	httpServer := &http.Server{
		Addr:        fmt.Sprintf(":%d", w.cfg.API.Port),
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	var mw []func(http.HandlerFunc) http.HandlerFunc
	if w.cfg.API.Username != "" && w.cfg.API.Password != "" {
		mw = append(mw, w.basicAuth)
	}

	mux.HandleFunc("/api/v1/sync", use(w.handleSync, mw...))
	mux.HandleFunc("/api/v1/logs", use(w.handleLogs, mw...))
	mux.HandleFunc("/favicon.ico", use(w.handleFavicon, mw...))
	mux.HandleFunc("/", use(w.handleRoot, mw...))

	go func() {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
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

	gracefullCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if w.cron != nil {
		l.Info("Stopping cron")
		w.cron.Stop()
	}

	if err := httpServer.Shutdown(gracefullCtx); err != nil {
		l.With("error", err).Error("Shutdown error")
		defer os.Exit(1)
	} else {
		l.Info("API server stopped")
	}

	// manually cancel context if not using httpServer.RegisterOnShutdown(cancel)
	cancel()

	defer os.Exit(0)
}
