const $ = (id) => document.getElementById(id);

const state = {
  path: "",
  exists: false,
  writable: false,
  yaml: "",
  data: {},
  proc: { running: false, pid: null, lastExit: "" },
  version: "" ,
  changes: [] ,
};

let lastLogId = 0;
let logsPaused = false;

let changesDebounce = null;
let lastChangesYaml = "";

function renderChanges(changes, errMsg=""){
  const list = $("changesList");
  const count = $("changesCount");
  if (errMsg){
    count.textContent = "!";
    list.innerHTML = "";
    const div = document.createElement("div");
    div.className = "changeItem err";
    div.textContent = errMsg;
    list.appendChild(div);
    return;
  }
  const arr = Array.isArray(changes) ? changes : [];
  state.changes = arr;
  count.textContent = String(arr.length);
  list.innerHTML = "";
  if (!arr.length){
    list.textContent = "No changes.";
    return;
  }
  for (const c of arr){
    const op = (c.op||c.Op||"mod").toLowerCase();
    const path = c.path||c.Path||"";
    const from = c.from ?? c.From;
    const to = c.to ?? c.To;
    const line = document.createElement("div");
    line.className = "changeItem " + (op === "add" ? "add" : op === "del" ? "del" : "mod");
    const f = from === undefined ? "" : safeVal(from);
    const t = to === undefined ? "" : safeVal(to);
    if (op === "add") line.textContent = `+ ${path}: ${t}`;
    else if (op === "del") line.textContent = `- ${path}: ${f}`;
    else line.textContent = `~ ${path}: ${f} → ${t}`;
    list.appendChild(line);
  }
}

function safeVal(v){
  try {
    if (typeof v === "string") return v.length > 120 ? (v.slice(0,117)+"…") : v;
    const s = JSON.stringify(v);
    return s.length > 120 ? (s.slice(0,117)+"…") : s;
  } catch {
    return String(v);
  }
}

async function updateChanges(force=false){
  const y = $("yaml").value || "";
  if (!force && y === lastChangesYaml) return;
  lastChangesYaml = y;
  if (!y.trim()){
    renderChanges([]);
    return;
  }
  try {
    const r = await api("/api/v1/changes", { method:"POST", body: JSON.stringify({ yaml: y }) });
    renderChanges(r.changes || []);
  } catch (e){
    renderChanges([], e.message || "Failed to compute changes.");
  }
}

function updateChangesDebounced(force=false){
  if (changesDebounce) clearTimeout(changesDebounce);
  changesDebounce = setTimeout(() => updateChanges(force), force ? 50 : 250);
}


function fmtBool(v){ return v ? "yes" : "no"; }

function setMsg(t, kind="") {
  const el = $("msg");
  el.textContent = t || "";
  if (kind === "ok") el.style.color = "var(--ok)";
  else if (kind === "bad") el.style.color = "var(--danger)";
  else el.style.color = "var(--muted)";
}

function badge(kind, text) {
  const b = $("badge");
  b.className = "badge " + (kind === "ok" ? "badge-ok" : kind === "bad" ? "badge-bad" : "badge-warn");
  b.textContent = text;
}

function esc(s){ return (s||"").replace(/[&<>"]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;'}[c])); }

function prettyOrigin(u){
  if(!u) return "—";
  try {
    const x = new URL(u);
    const host = x.hostname + (x.port ? (":"+x.port) : "");
    return host || u;
  } catch {
    // best-effort for values without scheme
    return String(u).replace(/^https?:\/\//i, "").split("/")[0];
  }
}

function renderStatus() {
  $("cfgPath").textContent = state.path || "—";
  $("cfgExists").textContent = state.exists ? "yes" : "no";
  $("cfgWritable").textContent = state.writable ? "yes" : "no";
  $("procState").textContent = state.proc.running ? "running" : "stopped";
  $("procPid").textContent = state.proc.pid ?? "—";
  $("procExit").textContent = state.proc.lastExit || "—";

  if (!state.writable) badge("bad", "Not writable");
  else if (!state.exists) badge("warn", "No config yet");
  else badge("ok", "Ready");

  $("yaml").value = state.yaml || "";

  // tiles
  const origin = getObj(state.data, "origin") || {};
  const originUrl = origin.url || "—";
  const k = replicasKey(state.data);
  const reps = Array.isArray(state.data[k]) ? state.data[k] : [];
  const features = getObj(state.data, "features") || null;

  $("tileCfg").textContent = state.exists ? "Loaded" : "Not set";
  $("tileCfgSub").textContent = `Writable: ${fmtBool(state.writable)}`;
  $("tileOrigin").textContent = prettyOrigin(originUrl);
  $("tileCron").textContent = (state.data && (state.data.cron || state.data.CRON)) ? `Cron: ${state.data.cron || state.data.CRON}` : "Cron: (default)";
  $("tileReplicas").textContent = `${reps.length}`;

  if (features && typeof features === "object") {
    const keys = Object.keys(features);
    const enabled = keys.filter(k => features[k] !== false).length;
    $("tileFeatures").textContent = `Features: ${enabled}/${keys.length}`;
  } else {
    $("tileFeatures").textContent = "Features: default";
  }

  $("tileProc").textContent = state.proc.running ? "Running" : "Stopped";
  $("tileProcSub").textContent = state.proc.running ? `PID ${state.proc.pid ?? "—"}` : (state.proc.lastExit || "—");
}

function getObj(root, key) {
  if (!root || typeof root !== "object") return null;
  if (root[key] && typeof root[key] === "object") return root[key];
  // tolerate different casing
  const k = Object.keys(root).find(x => x.toLowerCase() === key.toLowerCase());
  if (k) return root[k];
  return null;
}

function setObj(root, key, value) {
  if (!root || typeof root !== "object") return;
  root[key] = value;
}

function replicasKey(root) {
  if (!root || typeof root !== "object") return "replica";
  if (Array.isArray(root.replica)) return "replica";
  if (Array.isArray(root.replicas)) return "replicas";
  // fallback: find first array of objects
  const k = Object.keys(root).find(x => Array.isArray(root[x]) && root[x].every(y => typeof y === "object"));
  return k || "replica";
}

function renderReplicas() {
  const box = $("replicaList");
  box.innerHTML = "";
  const k = replicasKey(state.data);
  const arr = Array.isArray(state.data[k]) ? state.data[k] : [];
  arr.forEach((r, idx) => {
    const url = (r && (r.url || r.URL || r.host)) ? (r.url || r.URL || r.host) : "";
    const user = (r && (r.username || r.user)) ? (r.username || r.user) : "";
    const item = document.createElement("div");
    item.className = "rep";
    item.innerHTML = `
      <div class="meta">
        <div class="u">${esc(url || "(no url)")}</div>
        <div class="s">${esc(user || "")}</div>
      </div>
      <div class="repBtns">
        <button class="btnMini" data-act="edit" data-idx="${idx}">Edit</button>
        <button class="btnMini danger" data-act="rm" data-idx="${idx}">Remove</button>
      </div>
    `;
    item.querySelectorAll("button.btnMini").forEach(btn => {
      btn.addEventListener("click", () => {
        const act = btn.dataset.act;
        const i = Number(btn.dataset.idx);
        if (act === "rm") {
          arr.splice(i, 1);
          state.data[k] = arr;
          setMsg("Replica removed (not saved yet).", "");
          renderReplicas();
          return;
        }
        if (act === "edit") {
          openReplicaDialog(i);
        }
      });
    });
    box.appendChild(item);
  });
}

function openReplicaDialog(editIdx=null) {
  const k = replicasKey(state.data);
  const arr = Array.isArray(state.data[k]) ? state.data[k] : [];
  const rep = (editIdx !== null && arr[editIdx]) ? arr[editIdx] : { url:"", username:"", password:"" };
  $("repTitle").textContent = editIdx === null ? "Add replica" : "Edit replica";
  $("repUrl").value = rep.url || rep.URL || rep.host || "";
  $("repUser").value = rep.username || rep.user || "";
  $("repPass").value = rep.password || "";
  $("dlgReplica").dataset.editIdx = (editIdx === null) ? "" : String(editIdx);
  $("dlgReplica").showModal();
}

function applyQuickToData() {
  const root = state.data || {};
  const origin = getObj(root, "origin") || {};
  if ($("originUrl").value.trim()) origin.url = $("originUrl").value.trim();
  if ($("originUser").value.trim()) origin.username = $("originUser").value.trim();
  if ($("originPass").value.trim()) origin.password = $("originPass").value;
  setObj(root, "origin", origin);
  state.data = root;
  setMsg("Quick fields applied (not saved yet).", "");
  renderReplicas();
}

function syncQuickFromData() {
  const root = state.data || {};
  const origin = getObj(root, "origin") || {};
  $("originUrl").value = origin.url || "";
  $("originUser").value = origin.username || "";
  $("originPass").value = origin.password || "";
  renderReplicas();
}

async function api(path, opts={}) {
  const res = await fetch(path, {
    // Ensure auth works behind reverse proxies and avoid any caching of API responses.
    credentials: "same-origin",
    cache: "no-store",
    headers: { "Content-Type": "application/json", ...(opts.headers||{}) },
    ...opts
  });
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch { /* ignore */ }
  if (!res.ok) {
    const msg = (data && data.error) ? data.error : `${res.status} ${res.statusText}`;
    throw new Error(msg);
  }
  return data;
}

async function load() {
  setMsg("");
  const r = await api("/api/v1/config");
  state.path = r.path;
  state.version = r.version || "";
  if ($("guiVer")) $("guiVer").textContent = state.version ? (`GUI v${state.version}`) : "";
  state.exists = r.exists;
  state.writable = r.writable;
  state.yaml = r.yaml || "";
  state.baseYaml = state.yaml;
  state.data = r.data || {};
  state.baseData = JSON.parse(JSON.stringify(state.data || {}));
  state.proc = r.process || state.proc;
  renderStatus();
  syncQuickFromData();
  updateChangesDebounced(true);
}

async function save(reload=false) {
  state.yaml = $("yaml").value;
  const payload = { yaml: state.yaml, reload };
  const r = await api("/api/v1/config", { method:"PUT", body: JSON.stringify(payload) });
  state.exists = r.exists;
  state.writable = r.writable;
  state.yaml = r.yaml || state.yaml;
  state.data = r.data || state.data;
  state.proc = r.process || state.proc;
  renderStatus();
  syncQuickFromData();
  state.baseYaml = state.yaml;
  state.baseData = JSON.parse(JSON.stringify(state.data || {}));
  updateChangesDebounced(true);
  setMsg(reload ? "Saved and reloaded." : "Saved.", "ok");
}

async function reload() {
  const r = await api("/api/v1/reload", { method:"POST", body:"{}" });
  state.proc = r.process || state.proc;
  renderStatus();
  setMsg("Reload triggered.", "ok");
}

function ensureDataFromYaml() {
  // Best-effort: ask backend to parse it for us (dry-run)
  // This keeps frontend tiny (no YAML parser in JS).
}

$("btnRefresh").addEventListener("click", () => load().catch(e => setMsg(e.message, "bad")));
$("btnSave").addEventListener("click", () => save(false).catch(e => setMsg(e.message, "bad")));
$("btnSaveReload").addEventListener("click", () => save(true).catch(e => setMsg(e.message, "bad")));
$("btnReload").addEventListener("click", () => reload().catch(e => setMsg(e.message, "bad")));

$("btnApplyQuick").addEventListener("click", () => {
  applyQuickToData();
  // send data back to server to get YAML (dry-run) + update editor
  api("/api/v1/render", { method:"POST", body: JSON.stringify({ data: state.data }) })
    .then(r => {
      state.yaml = r.yaml || state.yaml;
      $("yaml").value = state.yaml;
      setMsg("Updated YAML from quick edit (not saved yet).", "");
      updateChangesDebounced();
    })
    .catch(e => setMsg(e.message, "bad"));
});

$("btnAddReplica").addEventListener("click", () => {
  openReplicaDialog(null);
});

$("dlgReplica").addEventListener("close", () => {
  if ($("dlgReplica").returnValue !== "ok") return;
  const rep = {
    url: $("repUrl").value.trim(),
    username: $("repUser").value.trim(),
    password: $("repPass").value
  };
  const k = replicasKey(state.data);
  const arr = Array.isArray(state.data[k]) ? state.data[k] : [];
  const editIdxRaw = $("dlgReplica").dataset.editIdx;
  const editIdx = editIdxRaw !== "" ? Number(editIdxRaw) : null;
  if (editIdx !== null && !Number.isNaN(editIdx) && arr[editIdx]) arr[editIdx] = rep;
  else arr.push(rep);
  state.data[k] = arr;
  renderReplicas();
  setMsg(editIdx !== null ? "Replica updated (not saved yet)." : "Replica added (not saved yet).", "");
  updateChangesDebounced();
});



/* ---------------- cron + api editors ---------------- */

function delKeyCI(obj, key){
  if (!obj || typeof obj !== 'object') return;
  if (obj[key] !== undefined) { delete obj[key]; return; }
  const lk = key.toLowerCase();
  for (const k of Object.keys(obj)) {
    if (k.toLowerCase() === lk) { delete obj[k]; return; }
  }
}

function getObjCI(obj, key){
  if (!obj || typeof obj !== 'object') return null;
  if (obj[key] && typeof obj[key] === 'object') return obj[key];
  const lk = key.toLowerCase();
  for (const k of Object.keys(obj)) {
    if (k.toLowerCase() === lk && obj[k] && typeof obj[k] === 'object') return obj[k];
  }
  return null;
}

function num(v, def){
  const n = Number(v);
  return Number.isFinite(n) ? n : def;
}

function showCronFields(mode){
  const fields = document.querySelectorAll('#cronFields .field');
  fields.forEach(f => {
    const show = (f.getAttribute('data-show') || '').split(/\s+/).filter(Boolean);
    f.classList.toggle('hidden', show.length && !show.includes(mode));
  });
}

function buildCronExprFromUI(){
  const mode = $("cronMode").value;
  if (mode === 'off') return '';
  if (mode === 'min') {
    const n = Math.max(1, Math.min(59, Math.floor(num($("cronEveryMin").value, 10))));
    return `*/${n} * * * *`;
  }
  if (mode === 'hour') return '0 * * * *';
  if (mode === 'hours') {
    const n = Math.max(1, Math.min(23, Math.floor(num($("cronEveryHours").value, 2))));
    const m = Math.max(0, Math.min(59, Math.floor(num($("cronHourMinute").value, 0))));
    return `${m} */${n} * * *`;
  }
  if (mode === 'daily') {
    const h = Math.max(0, Math.min(23, Math.floor(num($("cronAtHour").value, 3))));
    const m = Math.max(0, Math.min(59, Math.floor(num($("cronAtMinute").value, 0))));
    return `${m} ${h} * * *`;
  }
  if (mode === 'weekly') {
    const h = Math.max(0, Math.min(23, Math.floor(num($("cronAtHour").value, 3))));
    const m = Math.max(0, Math.min(59, Math.floor(num($("cronAtMinute").value, 0))));
    const d = String($("cronWeekday").value || '1');
    return `${m} ${h} * * ${d}`;
  }
  // custom
  return ($("cronCustom").value || '').trim();
}

function updateCronPreview(){
  const expr = buildCronExprFromUI();
  $("cronExpr").value = expr || '(disabled)';
}

function parseCronIntoUI(expr){
  const e = (expr || '').trim();
  // defaults
  $("cronMode").value = 'off';
  $("cronEveryMin").value = '10';
  $("cronEveryHours").value = '2';
  $("cronHourMinute").value = '0';
  $("cronAtHour").value = '3';
  $("cronAtMinute").value = '0';
  $("cronWeekday").value = '1';
  $("cronCustom").value = e;

  if (!e) {
    $("cronMode").value = 'off';
  } else {
    const parts = e.split(/\s+/);
    if (parts.length === 5) {
      const [m,h,dom,mon,dow] = parts;
      // */N * * * *
      const mm = m.match(/^\*\/(\d{1,2})$/);
      if (mm && h==='*' && dom==='*' && mon==='*' && dow==='*') {
        $("cronMode").value = 'min';
        $("cronEveryMin").value = mm[1];
      } else if (m==='0' && h==='*' && dom==='*' && mon==='*' && dow==='*') {
        $("cronMode").value = 'hour';
      } else {
        const hh = h.match(/^\*\/(\d{1,2})$/);
        if (hh && dom==='*' && mon==='*' && dow==='*') {
          $("cronMode").value = 'hours';
          $("cronEveryHours").value = hh[1];
          $("cronHourMinute").value = String(num(m,0));
        } else if (dom==='*' && mon==='*' && dow==='*') {
          // daily
          $("cronMode").value = 'daily';
          $("cronAtHour").value = String(num(h,3));
          $("cronAtMinute").value = String(num(m,0));
        } else if (dom==='*' && mon==='*') {
          // weekly
          $("cronMode").value = 'weekly';
          $("cronAtHour").value = String(num(h,3));
          $("cronAtMinute").value = String(num(m,0));
          $("cronWeekday").value = String(dow);
        } else {
          $("cronMode").value = 'custom';
          $("cronCustom").value = e;
        }
      }
    } else {
      $("cronMode").value = 'custom';
      $("cronCustom").value = e;
    }
  }

  showCronFields($("cronMode").value);
  updateCronPreview();
}

function updateYamlFromData(infoMsg){
  api("/api/v1/render", { method:"POST", body: JSON.stringify({ data: state.data }) })
    .then(r => {
      state.yaml = r.yaml || state.yaml;
      $("yaml").value = state.yaml;
      if (infoMsg) setMsg(infoMsg, "");
      updateChangesDebounced();
      renderStatus();
    })
    .catch(e => setMsg(e.message, "bad"));
}

$("btnEditCron").addEventListener("click", () => {
  const cron = (state.data && (state.data.cron || state.data.CRON)) || '';
  parseCronIntoUI(cron);
  $("cronRunOnStart").checked = !!(state.data && (state.data.runOnStart || state.data.RUN_ON_START));
  $("dlgCron").showModal();
});

$("cronMode").addEventListener("change", () => {
  showCronFields($("cronMode").value);
  updateCronPreview();
});

["cronEveryMin","cronEveryHours","cronHourMinute","cronAtHour","cronAtMinute","cronWeekday","cronCustom"].forEach(id => {
  const el = $(id);
  if (!el) return;
  el.addEventListener("input", () => updateCronPreview());
});

$("dlgCron").addEventListener("close", () => {
  if ($("dlgCron").returnValue !== "ok") return;
  const mode = $("cronMode").value;
  const expr = buildCronExprFromUI();
  const runOnStart = $("cronRunOnStart").checked;

  if (!state.data || typeof state.data !== 'object') state.data = {};

  if (mode === 'off' || !expr) delKeyCI(state.data, 'cron');
  else state.data.cron = expr;

  if (runOnStart) state.data.runOnStart = true;
  else delKeyCI(state.data, 'runOnStart');

  updateYamlFromData("Cron updated (not saved yet).\nClick Save or Save & Reload.");
});

$("btnEditApi").addEventListener("click", () => {
  const apiObj = getObjCI(state.data, 'api') || {};
  const port = apiObj.port ?? apiObj.Port ?? 0;
  const enabled = Number(port) > 0;
  $("apiEnabled").checked = enabled;
  $("apiPort").value = enabled ? String(port) : '';
  $("apiUser").value = apiObj.username ?? apiObj.Username ?? '';
  $("apiPass").value = apiObj.password ?? apiObj.Password ?? '';
  $("apiDark").checked = !!(apiObj.darkMode ?? apiObj.DarkMode);

  const metrics = (apiObj.metrics && typeof apiObj.metrics==='object') ? apiObj.metrics : {};
  $("apiMetricsEnabled").checked = !!(metrics.enabled ?? metrics.Enabled);
  $("apiScrapeInterval").value = metrics.scrapeInterval ?? metrics.ScrapeInterval ?? '30s';
  $("apiQueryLogLimit").value = metrics.queryLogLimit ?? metrics.QueryLogLimit ?? '';

  // sensible default when enabling
  if (!enabled) $("apiPort").placeholder = '8081';

  $("dlgApi").showModal();
});

$("dlgApi").addEventListener("close", () => {
  if ($("dlgApi").returnValue !== "ok") return;

  if (!state.data || typeof state.data !== 'object') state.data = {};

  const enabled = $("apiEnabled").checked;
  const port = Math.max(0, Math.min(65535, Math.floor(num($("apiPort").value, enabled ? 8081 : 0))));
  const user = $("apiUser").value.trim();
  const pass = $("apiPass").value;
  const dark = $("apiDark").checked;

  const metricsEnabled = $("apiMetricsEnabled").checked;
  const scrape = $("apiScrapeInterval").value.trim() || '30s';
  const qll = $("apiQueryLogLimit").value.trim();

  // merge with existing object to not drop unknown keys
  const apiObj = getObjCI(state.data, 'api') || {};

  apiObj.port = enabled ? port : 0;
  if (user) apiObj.username = user; else delKeyCI(apiObj, 'username');
  if (pass) apiObj.password = pass; else delKeyCI(apiObj, 'password');
  apiObj.darkMode = !!dark;

  if (metricsEnabled) {
    apiObj.metrics = (apiObj.metrics && typeof apiObj.metrics==='object') ? apiObj.metrics : {};
    apiObj.metrics.enabled = true;
    apiObj.metrics.scrapeInterval = scrape;
    if (qll) apiObj.metrics.queryLogLimit = Math.max(0, Math.floor(num(qll, 0)));
    else delKeyCI(apiObj.metrics, 'queryLogLimit');
  } else {
    if (apiObj.metrics && typeof apiObj.metrics==='object') apiObj.metrics.enabled = false;
  }

  // clean-up if completely unused
  const hasAny = Object.keys(apiObj).some(k => {
    const v = apiObj[k];
    if (k === 'port') return Number(v) > 0;
    if (k === 'metrics') {
      return v && typeof v==='object' && (v.enabled || v.scrapeInterval || v.queryLogLimit);
    }
    return v !== '' && v !== false && v != null;
  });

  if (!hasAny) delKeyCI(state.data, 'api');
  else state.data.api = apiObj;

  updateYamlFromData("API updated (not saved yet).\nClick Save or Save & Reload.");
});

/* ---------------- logs ---------------- */

function appendLogs(entries) {
  if (!entries || !entries.length) return;
  const el = $("logs");
  const frag = document.createDocumentFragment();
  for (const e of entries) {
    lastLogId = Math.max(lastLogId, e.id);
    const line = document.createElement("div");
    const low = (e.line||"").toLowerCase();
    const isErr = e.stream === "stderr" || /\b(error|failed|fatal|panic)\b/i.test(e.line||"");
    const isWarn = !isErr && /\b(warn|warning)\b/i.test(e.line||"");
    line.className = "logLine " + (isErr ? "err" : isWarn ? "warn" : "out");
    line.textContent = `[${e.ts}] ${e.stream}: ${e.line}`;
    frag.appendChild(line);
  }
  el.appendChild(frag);
  // cap DOM nodes to keep memory low
  while (el.childNodes.length > 1200) el.removeChild(el.firstChild);
  if (!logsPaused) el.scrollTop = el.scrollHeight;
}

async function pollLogs(initial=false) {
  if (logsPaused) return;
  const limit = initial ? 300 : 250;
  const since = initial ? 0 : lastLogId;
  const r = await api(`/api/v1/logs?since=${since}&limit=${limit}`);
  appendLogs(r.entries || []);
}

$("btnLogsClear").addEventListener("click", () => {
  $("logs").textContent = "";
  setMsg("Log view cleared (in-memory log buffer remains).", "");
});

$("btnLogsCopy").addEventListener("click", async () => {
  const text = $("logs").innerText || "";
  try {
    await navigator.clipboard.writeText(text);
    setMsg("Logs copied to clipboard.", "ok");
  } catch {
    setMsg("Copy failed (browser permission).", "bad");
  }
});

$("btnLogsPause").addEventListener("click", () => {
  logsPaused = !logsPaused;
  $("btnLogsPause").textContent = logsPaused ? "Resume" : "Pause";
  if (!logsPaused) pollLogs(false).catch(()=>{});
});

// initial load
load().catch(e => {
  badge("bad", "Error");
  setMsg(e.message, "bad");
});

// start log polling
pollLogs(true).catch(()=>{});
// Use an async loop instead of setInterval; some embedded UIs / browsers throttle timers
// aggressively, which can make logs appear to update only on manual refresh.
(function startLogLoop(){
  const sleep = (ms) => new Promise(r => setTimeout(r, ms));
  (async () => {
    while (true) {
      if (!logsPaused) {
        try { await pollLogs(false); } catch { /* ignore */ }
      }
      await sleep(1200);
    }
  })();
})();


$("yaml").addEventListener("input", () => {
  updateChangesDebounced();
});
