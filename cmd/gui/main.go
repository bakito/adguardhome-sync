package main

import (
	"bufio"
	"context"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

//go:embed ui/*
var uiFS embed.FS

// GUIVersion is shown in the UI and returned via the API.
// Bump this when you release a new GUI build.
var GUIVersion = "0.1.4"

type processStatus struct {
	Running  bool   `json:"running"`
	PID      int    `json:"pid,omitempty"`
	LastExit string `json:"lastExit,omitempty"`
}

type configResponse struct {
	Path     string         `json:"path"`
	Exists   bool           `json:"exists"`
	Writable bool           `json:"writable"`
	YAML     string         `json:"yaml"`
	Data     map[string]any `json:"data"`
	Process  processStatus  `json:"process"`
	Error    string         `json:"error,omitempty"`
	Now      string         `json:"now,omitempty"`
	Version  string         `json:"version,omitempty"`
}

type putConfigRequest struct {
	YAML   string `json:"yaml"`
	Reload bool   `json:"reload"`
}

type logEntry struct {
	ID     int64  `json:"id"`
	TS     string `json:"ts"`
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

type logsResponse struct {
	Entries []logEntry `json:"entries"`
	NextID  int64      `json:"nextId"`
}

type renderRequest struct {
	Data map[string]any `json:"data"`
}
type changesRequest struct {
	YAML string `json:"yaml"`
}

type changeItem struct {
	Op   string `json:"op"`
	Path string `json:"path"`
	From any    `json:"from,omitempty"`
	To   any    `json:"to,omitempty"`
}

type changesResponse struct {
	Changes []changeItem `json:"changes"`
}

type guiServer struct {
	cfgPath string
	bind    string

	guiUser string
	guiPass string

	mu       sync.Mutex
	proc     *managedProc
	lastExit string
	started  time.Time
}

func main() {
	bind := getenv("GUI_BIND", "0.0.0.0:8080")
	s := &guiServer{
		cfgPath: getenv("CONFIG_PATH", "/config/adguardhome-sync.yaml"),
		bind:    bind,
		guiUser: os.Getenv("GUI_USERNAME"),
		guiPass: os.Getenv("GUI_PASSWORD"),
		proc:    newManagedProc(getenv("SYNC_BIN", "/usr/local/bin/adguardhome-sync"), bindPort(bind, 8080)),
		started: time.Now(),
	}

	// Start child only if config exists (so first-run UI still works without a file)
	if _, err := os.Stat(s.cfgPath); err == nil {
		if err := s.proc.Start(s.cfgPath, getenv("SYNC_ARGS", "")); err != nil {
			log.Printf("sync start failed: %v", err)
		}
	} else {
		log.Printf("no config at %s yet; GUI will start without running sync", s.cfgPath)
	}

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/v1/config", s.withAuth(s.handleConfig))
	mux.HandleFunc("/api/v1/reload", s.withAuth(s.handleReload))
	mux.HandleFunc("/api/v1/render", s.withAuth(s.handleRender))
	mux.HandleFunc("/api/v1/logs", s.withAuth(s.handleLogs))
	mux.HandleFunc("/api/v1/changes", s.withAuth(s.handleChanges))

	// UI assets
	sub, _ := fs.Sub(uiFS, "ui")
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/ui/", http.StripPrefix("/ui/", fileServer))
	mux.HandleFunc("/", s.withAuth(s.handleIndex))

	srv := &http.Server{
		Addr:              s.bind,
		Handler:           withNoCache(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown
	ctx, stop := signalContext()
	defer stop()

	go func() {
		<-ctx.Done()
		log.Println("shutting down...")
		shCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		_ = srv.Shutdown(shCtx)
		_ = s.proc.Stop()
	}()

	log.Printf("GUI listening on http://%s", s.bind)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("http server error: %v", err)
	}
}

func (s *guiServer) withAuth(next http.HandlerFunc) http.HandlerFunc {
	if s.guiUser == "" || s.guiPass == "" {
		return next
	}
	return func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(u), []byte(s.guiUser)) != 1 || subtle.ConstantTimeCompare([]byte(p), []byte(s.guiPass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="adguardhome-sync-gui"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *guiServer) handleIndex(w http.ResponseWriter, r *http.Request) {
	// SPA entry
	b, err := uiFS.ReadFile("ui/index.html")
	if err != nil {
		http.Error(w, "ui missing", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func (s *guiServer) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetConfig(w, r)
	case http.MethodPut:
		s.handlePutConfig(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *guiServer) handleGetConfig(w http.ResponseWriter, _ *http.Request) {
	yml, exists := readFileIfExists(s.cfgPath)
	data := map[string]any{}
	if exists {
		if m, err := parseYAMLToMap(yml); err == nil {
			data = m
		}
	}

	resp := configResponse{
		Path:     s.cfgPath,
		Exists:   exists,
		Writable: canWritePath(s.cfgPath),
		YAML:     yml,
		Data:     data,
		Process:  s.proc.Status(),
		Now:      time.Now().Format(time.RFC3339),
		Version:  GUIVersion,
	}
	writeJSON(w, resp)
}

func (s *guiServer) handlePutConfig(w http.ResponseWriter, r *http.Request) {
	var req putConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	if !canWritePath(s.cfgPath) {
		http.Error(w, "config path is not writable (check volume permissions)", http.StatusForbidden)
		return
	}

	// Validate YAML
	if strings.TrimSpace(req.YAML) == "" {
		http.Error(w, "yaml is empty", http.StatusBadRequest)
		return
	}
	m, err := parseYAMLToMap(req.YAML)
	if err != nil {
		http.Error(w, fmt.Sprintf("yaml parse error: %v", err), http.StatusBadRequest)
		return
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.cfgPath), 0o755); err != nil {
		http.Error(w, fmt.Sprintf("mkdir error: %v", err), http.StatusInternalServerError)
		return
	}

	// Write file atomically
	if err := atomicWriteFile(s.cfgPath, []byte(req.YAML), 0o600); err != nil {
		http.Error(w, fmt.Sprintf("write error: %v", err), http.StatusInternalServerError)
		return
	}

	if req.Reload {
		_ = s.proc.Restart(s.cfgPath, getenv("SYNC_ARGS", ""))
	}

	resp := configResponse{
		Path:     s.cfgPath,
		Exists:   true,
		Writable: canWritePath(s.cfgPath),
		YAML:     req.YAML,
		Data:     m,
		Process:  s.proc.Status(),
		Now:      time.Now().Format(time.RFC3339),
		Version:  GUIVersion,
	}
	writeJSON(w, resp)
}

func (s *guiServer) handleReload(w http.ResponseWriter, _ *http.Request) {
	// If config doesn't exist, just keep GUI running
	if _, err := os.Stat(s.cfgPath); err != nil {
		http.Error(w, "config file not found; create and save it first", http.StatusBadRequest)
		return
	}
	if err := s.proc.Restart(s.cfgPath, getenv("SYNC_ARGS", "")); err != nil {
		http.Error(w, fmt.Sprintf("reload failed: %v", err), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"process": s.proc.Status()})
}

func (s *guiServer) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	since := int64(0)
	limit := 400
	if v := r.URL.Query().Get("since"); v != "" {
		_, _ = fmt.Sscan(v, &since)
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		_, _ = fmt.Sscan(v, &limit)
		if limit < 1 {
			limit = 1
		}
		if limit > 2000 {
			limit = 2000
		}
	}
	entries, next := s.proc.Logs().GetSince(since, limit)
	writeJSON(w, logsResponse{Entries: entries, NextID: next})
}

func (s *guiServer) handleRender(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req renderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	yml, err := yaml.Marshal(req.Data)
	if err != nil {
		http.Error(w, fmt.Sprintf("yaml marshal error: %v", err), http.StatusBadRequest)
		return
	}
	// Validate output parses
	if _, err := parseYAMLToMap(string(yml)); err != nil {
		http.Error(w, fmt.Sprintf("rendered yaml invalid: %v", err), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]any{"yaml": string(yml)})
}

func (s *guiServer) handleChanges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req changesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.YAML) == "" {
		writeJSON(w, changesResponse{Changes: []changeItem{}})
		return
	}

	// Base config: current file (if present)
	baseYml, exists := readFileIfExists(s.cfgPath)
	baseMap := map[string]any{}
	if exists {
		if m, err := parseYAMLToMap(baseYml); err == nil {
			baseMap = m
		}
	}

	// Proposed config
	newMap, err := parseYAMLToMap(req.YAML)
	if err != nil {
		http.Error(w, fmt.Sprintf("yaml parse error: %v", err), http.StatusBadRequest)
		return
	}

	changes := make([]changeItem, 0, 64)
	diffAny("", baseMap, newMap, &changes, 240)
	writeJSON(w, changesResponse{Changes: changes})
}

func diffAny(path string, a any, b any, out *[]changeItem, limit int) {
	if len(*out) >= limit {
		return
	}

	if a == nil && b == nil {
		return
	}
	if a == nil {
		*out = append(*out, changeItem{Op: "add", Path: cleanPath(path), To: b})
		return
	}
	if b == nil {
		*out = append(*out, changeItem{Op: "del", Path: cleanPath(path), From: a})
		return
	}

	// Map
	am, aIsMap := a.(map[string]any)
	bm, bIsMap := b.(map[string]any)
	if aIsMap && bIsMap {
		keys := make([]string, 0, len(am)+len(bm))
		seen := map[string]struct{}{}
		for k := range am {
			seen[k] = struct{}{}
		}
		for k := range bm {
			seen[k] = struct{}{}
		}
		for k := range seen {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if len(*out) >= limit {
				return
			}
			av, aok := am[k]
			bv, bok := bm[k]
			sub := joinPath(path, k)
			if !aok {
				*out = append(*out, changeItem{Op: "add", Path: cleanPath(sub), To: bv})
				continue
			}
			if !bok {
				*out = append(*out, changeItem{Op: "del", Path: cleanPath(sub), From: av})
				continue
			}
			diffAny(sub, av, bv, out, limit)
		}
		return
	}

	// Slice
	as, aIsSlice := a.([]any)
	bs, bIsSlice := b.([]any)
	if aIsSlice && bIsSlice {
		max := len(as)
		if len(bs) > max {
			max = len(bs)
		}
		for i := 0; i < max; i++ {
			if len(*out) >= limit {
				return
			}
			sub := joinIndex(path, i)
			var av any
			var bv any
			if i < len(as) {
				av = as[i]
			}
			if i < len(bs) {
				bv = bs[i]
			}
			diffAny(sub, av, bv, out, limit)
		}
		return
	}

	// Scalar
	if !reflect.DeepEqual(a, b) {
		*out = append(*out, changeItem{Op: "mod", Path: cleanPath(path), From: a, To: b})
	}
}

func joinPath(base, key string) string {
	if base == "" {
		return key
	}
	return base + "." + key
}

func joinIndex(base string, idx int) string {
	if base == "" {
		return fmt.Sprintf("[%d]", idx)
	}
	return fmt.Sprintf("%s[%d]", base, idx)
}

func cleanPath(p string) string {
	if p == "" {
		return "(root)"
	}
	return p
}

/* ----------------- process management ----------------- */

type managedProc struct {
	bin     string
	logs    *logStore
	guiPort int

	mu   sync.Mutex
	cmd  *exec.Cmd
	last string
}

func newManagedProc(bin string, guiPort int) *managedProc {
	return &managedProc{bin: bin, logs: newLogStore(2500), guiPort: guiPort}
}

func (p *managedProc) Logs() *logStore { return p.logs }

func (p *managedProc) Status() processStatus {
	p.mu.Lock()
	defer p.mu.Unlock()
	st := processStatus{LastExit: p.last}
	if p.cmd != nil && p.cmd.Process != nil {
		st.Running = p.cmd.ProcessState == nil || !p.cmd.ProcessState.Exited()
		st.PID = p.cmd.Process.Pid
	}
	return st
}

func (p *managedProc) computeChildAPIPort(cfgPath string) string {
	// Highest priority: explicit env override
	if v := strings.TrimSpace(os.Getenv("SYNC_API_PORT")); v != "" {
		return v
	}

	// Next: api.port from YAML (if present)
	yml, ok := readFileIfExists(cfgPath)
	if ok {
		if m, err := parseYAMLToMap(yml); err == nil {
			if apiAny, ok := mapGetCI(m, "api"); ok {
				if api, ok := apiAny.(map[string]any); ok {
					if portAny, ok := mapGetCI(api, "port"); ok {
						if n, ok := asInt(portAny); ok {
							return fmt.Sprint(n)
						}
						if s, ok := portAny.(string); ok && strings.TrimSpace(s) != "" {
							return strings.TrimSpace(s)
						}
					}
				}
			}
		}
	}

	// Default: disable to avoid port clash with GUI (upstream default is 8080)
	return "0"
}

func (p *managedProc) Start(cfgPath string, extraArgs string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cmd != nil && p.cmd.Process != nil && p.cmd.ProcessState == nil {
		return nil // already running
	}

	args := []string{"run", "--config", cfgPath}
	if strings.TrimSpace(extraArgs) != "" {
		args = append(args, strings.Fields(extraArgs)...)
	}

	cmd := exec.Command(p.bin, args...)
	cmd.Stdout = io.MultiWriter(os.Stdout, p.logs.Writer("stdout"))
	cmd.Stderr = io.MultiWriter(os.Stderr, p.logs.Writer("stderr"))

	// NOTE: adguardhome-sync exposes its own API (defaults to :8080). The GUI also
	// listens on :8080, so we disable the sync API by default.
	// Enable it by setting api.port in the YAML to a non-zero, non-conflicting port
	// (e.g. 8081), or by setting SYNC_API_PORT to force a specific port.
	apiPort := p.computeChildAPIPort(cfgPath)
	if p.guiPort != 0 && apiPort != "0" && apiPort == fmt.Sprint(p.guiPort) {
		p.logs.append("stderr", fmt.Sprintf("sync API port %s conflicts with GUI port %d; disabling sync API. Set api.port to e.g. 8081 or set SYNC_API_PORT.", apiPort, p.guiPort))
		apiPort = "0"
	}
	cmd.Env = setEnv(os.Environ(), "API_PORT", apiPort)

	if err := cmd.Start(); err != nil {
		return err
	}

	p.cmd = cmd

	go func() {
		err := cmd.Wait()
		p.mu.Lock()
		defer p.mu.Unlock()
		if err != nil {
			p.last = err.Error()
		} else {
			p.last = "exited 0"
		}
	}()

	return nil
}

/* ----------------- lightweight log capture ----------------- */

type logStore struct {
	mu     sync.Mutex
	max    int
	nextID int64
	items  []logEntry
}

func newLogStore(max int) *logStore {
	if max < 100 {
		max = 100
	}
	return &logStore{max: max, nextID: 1, items: make([]logEntry, 0, max)}
}

func (s *logStore) append(stream, line string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(line) == "" {
		return
	}
	entry := logEntry{
		ID:     s.nextID,
		TS:     time.Now().Format(time.RFC3339Nano),
		Stream: stream,
		Line:   line,
	}
	s.nextID++
	s.items = append(s.items, entry)
	if len(s.items) > s.max {
		over := len(s.items) - s.max
		s.items = append([]logEntry(nil), s.items[over:]...)
	}
}

func (s *logStore) Writer(stream string) io.Writer {
	return &streamWriter{store: s, stream: stream}
}

func (s *logStore) GetSince(since int64, limit int) ([]logEntry, int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit < 1 {
		limit = 1
	}
	if limit > 2000 {
		limit = 2000
	}
	if since <= 0 {
		start := 0
		if len(s.items) > limit {
			start = len(s.items) - limit
		}
		out := append([]logEntry(nil), s.items[start:]...)
		return out, s.nextID
	}
	out := make([]logEntry, 0, limit)
	for _, it := range s.items {
		if it.ID > since {
			out = append(out, it)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, s.nextID
}

type streamWriter struct {
	store  *logStore
	stream string
	mu     sync.Mutex
	buf    []byte
}

func (w *streamWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.buf = append(w.buf, p...)
	reader := strings.NewReader(string(w.buf))
	scanner := bufio.NewScanner(reader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 512*1024)
	lines := make([]string, 0, 8)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if len(w.buf) > 0 && w.buf[len(w.buf)-1] != '\n' {
		if len(lines) > 0 {
			partial := lines[len(lines)-1]
			lines = lines[:len(lines)-1]
			w.buf = []byte(partial)
		} else {
			return len(p), nil
		}
	} else {
		w.buf = w.buf[:0]
	}
	for _, ln := range lines {
		w.store.append(w.stream, ln)
	}
	return len(p), nil
}

func (p *managedProc) Stop() error {
	p.mu.Lock()
	cmd := p.cmd
	p.mu.Unlock()

	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return stopProcess(cmd)
}

func (p *managedProc) Restart(cfgPath string, extraArgs string) error {
	_ = p.Stop()
	time.Sleep(250 * time.Millisecond)
	return p.Start(cfgPath, extraArgs)
}

func stopProcess(cmd *exec.Cmd) error {
	// try graceful first
	_ = cmd.Process.Signal(syscall.SIGTERM)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		return err
	case <-time.After(6 * time.Second):
		_ = cmd.Process.Kill()
		return errors.New("killed after timeout")
	}
}

/* ----------------- helpers ----------------- */

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func bindPort(addr string, def int) int {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return def
	}
	portStr := ""
	if _, p, err := net.SplitHostPort(addr); err == nil {
		portStr = p
	} else {
		if i := strings.LastIndex(addr, ":"); i != -1 && i+1 < len(addr) {
			portStr = addr[i+1:]
		} else {
			portStr = addr
		}
	}
	if n, err := strconv.Atoi(strings.TrimSpace(portStr)); err == nil && n > 0 && n < 65536 {
		return n
	}
	return def
}

func mapGetCI(m map[string]any, key string) (any, bool) {
	if m == nil {
		return nil, false
	}
	if v, ok := m[key]; ok {
		return v, true
	}
	lk := strings.ToLower(key)
	for k, v := range m {
		if strings.ToLower(k) == lk {
			return v, true
		}
	}
	return nil, false
}

func asInt(v any) (int, bool) {
	switch x := v.(type) {
	case int:
		return x, true
	case int64:
		return int(x), true
	case int32:
		return int(x), true
	case float64:
		return int(x), true
	case float32:
		return int(x), true
	case uint:
		return int(x), true
	case uint64:
		return int(x), true
	case uint32:
		return int(x), true
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0, false
		}
		if n, err := strconv.Atoi(s); err == nil {
			return n, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// setEnv returns a copy of env with key set to value (replacing any existing entry).
func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	out = append(out, prefix+value)
	return out
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func withNoCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// helpful when editing UI assets
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func readFileIfExists(path string) (string, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(b), true
}

func canWritePath(path string) bool {
	// If file exists: check if we can open for write without truncation
	if _, err := os.Stat(path); err == nil {
		f, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			return false
		}
		_ = f.Close()
		return true
	}

	// If not: check directory by creating a temp file
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	test := filepath.Join(dir, ".writecheck")
	f, err := os.OpenFile(test, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return false
	}
	_ = f.Close()
	_ = os.Remove(test)
	return true
}

func atomicWriteFile(path string, data []byte, perm fs.FileMode) error {
	dir := filepath.Dir(path)
	tmp := filepath.Join(dir, fmt.Sprintf(".%s.tmp", filepath.Base(path)))
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func parseYAMLToMap(yml string) (map[string]any, error) {
	var raw any
	if err := yaml.Unmarshal([]byte(yml), &raw); err != nil {
		return nil, err
	}
	n := normalize(raw)
	m, ok := n.(map[string]any)
	if !ok {
		return map[string]any{}, nil
	}
	return m, nil
}

func normalize(v any) any {
	switch x := v.(type) {
	case map[any]any:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			m[fmt.Sprint(k)] = normalize(vv)
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, vv := range x {
			m[k] = normalize(vv)
		}
		return m
	case []any:
		out := make([]any, 0, len(x))
		for _, it := range x {
			out = append(out, normalize(it))
		}
		return out
	default:
		return x
	}
}

// signalContext provides a cancelable context that is canceled on SIGINT/SIGTERM.
func signalContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	ch := make(chan os.Signal, 2)
	signalNotify(ch)
	go func() {
		<-ch
		cancel()
	}()
	return ctx, cancel
}

func signalNotify(ch chan os.Signal) {
	// separate function for tiny build if needed
	signalNotifyImpl(ch)
}
