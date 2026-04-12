package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultInspectorAddr = ":4040"

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize:  1024,
	CheckOrigin:      func(r *http.Request) bool { return true },
	HandshakeTimeout: 10 * time.Second,
}

// inspectorHub owns WebSocket clients for live log streaming. It is safe for
// concurrent use.
type inspectorHub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]struct{}
}

func newInspectorHub() *inspectorHub {
	return &inspectorHub{clients: make(map[*websocket.Conn]struct{})}
}

func (h *inspectorHub) register(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *inspectorHub) unregister(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	_ = c.Close()
}

func (h *inspectorHub) closeAll() {
	h.mu.Lock()
	for c := range h.clients {
		delete(h.clients, c)
		_ = c.Close()
	}
	h.mu.Unlock()
}

func (h *inspectorHub) broadcast(entry RequestLog) {
	h.mu.Lock()
	list := make([]*websocket.Conn, 0, len(h.clients))
	for c := range h.clients {
		list = append(list, c)
	}
	h.mu.Unlock()

	for _, c := range list {
		if err := c.WriteJSON(entry); err != nil {
			h.unregister(c)
		}
	}
}

var (
	inspectorMu     sync.RWMutex
	inspectorHubPtr *inspectorHub
)

func notifyInspectorSubscribers(entry RequestLog) {
	inspectorMu.RLock()
	h := inspectorHubPtr
	inspectorMu.RUnlock()
	if h == nil {
		return
	}
	go h.broadcast(entry)
}

// startInspector runs the traffic inspector UI on addr (e.g. ":4040"). localPort
// is the tunnel target (your app) and is used for POST /replay. It uses a
// dedicated [http.ServeMux], not [http.DefaultServeMux].
func startInspector(addr, localPort string) func() {
	if addr == "" {
		addr = defaultInspectorAddr
	}

	hub := newInspectorHub()
	inspectorMu.Lock()
	inspectorHubPtr = hub
	inspectorMu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", serveInspectorUI)
	mux.HandleFunc("GET /logs", serveInspectorLogs)
	mux.HandleFunc("POST /replay", handleReplay(localPort))
	mux.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		handleInspectorWS(hub, w, r)
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Fprintf(os.Stderr, "gotunnel: traffic inspector → http://127.0.0.1%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "gotunnel: inspector stopped: %v\n", err)
		}
	}()

	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)

		inspectorMu.Lock()
		inspectorHubPtr = nil
		inspectorMu.Unlock()

		hub.closeAll()
	}
}

func serveInspectorUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	const page = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Tunnel traffic — dev</title>
<style>
:root {
  --bg: #0d1117;
  --panel: #161b22;
  --border: #30363d;
  --text: #e6edf3;
  --muted: #8b949e;
  --accent: #58a6ff;
  --green: #3fb950;
  --danger: #f85149;
  --row-alt: #21262d;
}
* { box-sizing: border-box; }
body {
  font-family: ui-sans-serif, system-ui, -apple-system, sans-serif;
  margin: 0; padding: 1rem 1.25rem 2rem;
  background: var(--bg); color: var(--text); min-height: 100vh;
}
header { max-width: 1200px; margin: 0 auto 1rem; }
header h1 { font-size: 1.25rem; font-weight: 600; margin: 0 0 0.35rem 0; }
header p { margin: 0; color: var(--muted); font-size: 0.875rem; }
.shell {
  display: grid; grid-template-columns: minmax(300px, 38%) minmax(0, 1fr);
  gap: 1rem; align-items: start; max-width: 1200px; margin: 0 auto;
}
@media (max-width: 880px) { .shell { grid-template-columns: 1fr; } }
.col-left {
  background: var(--panel); border: 1px solid var(--border); border-radius: 10px;
  padding: 0.65rem 0.75rem; min-height: 200px;
}
.col-left .left-topbar {
  display: flex; flex-wrap: wrap; gap: 0.5rem; justify-content: space-between; align-items: center;
  margin-bottom: 0.65rem; padding-bottom: 0.5rem; border-bottom: 1px solid var(--border);
}
.col-right {
  background: var(--panel); border: 1px solid var(--border); border-radius: 10px;
  padding: 0.85rem 1rem 1rem; min-height: 280px;
}
.detail-toolbar {
  display: flex; flex-wrap: wrap; justify-content: space-between; align-items: flex-start;
  gap: 0.5rem; margin-bottom: 0.75rem;
}
.detail-toolbar .toolbar-actions { display: flex; gap: 0.45rem; flex-shrink: 0; }
#detail-badge { font-size: 12px; color: var(--muted); max-width: 55%; line-height: 1.35; }
.btn {
  padding: 0.4rem 0.85rem; font-size: 13px; cursor: pointer; border-radius: 6px;
  border: 1px solid var(--border); background: var(--row-alt); color: var(--text); font-weight: 500;
}
.btn:hover:not(:disabled) { background: #262c36; }
.btn-primary { background: #238636; border-color: #2ea043; color: #fff; }
.btn-primary:hover:not(:disabled) { background: #2ea043; }
.btn-sm { padding: 0.25rem 0.55rem; font-size: 12px; }
#log-list { display: flex; flex-direction: column; gap: 0.5rem; }
.log-card {
  background: #1c2128; border: 1px solid var(--border); border-radius: 8px;
  overflow: hidden;
}
.log-card.selected { box-shadow: 0 0 0 2px var(--accent); border-color: #388bfd; }
.log-card-head {
  display: grid;
  grid-template-columns: auto 1fr auto auto auto;
  gap: 0.45rem 0.6rem; align-items: center;
  padding: 0.5rem 0.6rem; font-size: 13px;
}
.log-card-head .method { font-weight: 600; color: var(--accent); }
.log-card-head .path { font-family: ui-monospace, monospace; font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.log-card-head .status { color: var(--green); font-weight: 600; font-size: 12px; }
.log-card-head .ms { color: var(--muted); font-size: 11px; }
.btn-toggle {
  border: 1px solid var(--border); background: var(--row-alt); color: var(--text);
  border-radius: 6px; padding: 0.25rem 0.5rem; font-size: 11px; cursor: pointer;
}
.btn-toggle:hover { border-color: var(--accent); color: var(--accent); }
.section-title {
  font-size: 11px; text-transform: uppercase; letter-spacing: 0.08em;
  color: var(--muted); margin: 0.85rem 0 0.45rem 0; font-weight: 600;
}
.section-title:first-of-type { margin-top: 0; }
#req-slot { min-height: 1rem; }
table.kv { width: 100%; border-collapse: collapse; font-size: 12px; margin: 0; }
table.kv th {
  text-align: left; vertical-align: top; width: 6.5rem;
  padding: 0.4rem 0.55rem; border: 1px solid var(--border);
  background: var(--row-alt); color: var(--muted); font-weight: 600;
}
table.kv td {
  padding: 0.4rem 0.55rem; border: 1px solid var(--border);
  word-break: break-word; background: #1c2128;
}
table.kv td pre, table.kv .mono {
  margin: 0; font-family: ui-monospace, Menlo, monospace; font-size: 11px;
  white-space: pre-wrap; max-height: 180px; overflow: auto;
}
table.kv-edit input[type="text"], table.kv-edit textarea {
  width: 100%; margin: 0; padding: 0.35rem 0.45rem; font-size: 12px;
  font-family: ui-monospace, Menlo, monospace;
  border: 1px solid var(--border); border-radius: 4px; background: #0d1117; color: var(--text);
}
table.kv-edit textarea { resize: vertical; min-height: 3.5rem; display: block; }
table.kv-edit input:focus, table.kv-edit textarea:focus {
  outline: none; border-color: var(--accent); box-shadow: 0 0 0 1px var(--accent);
}
.err-box {
  color: var(--danger); font-family: ui-monospace, monospace; font-size: 12px;
  padding: 0.5rem; border: 1px solid var(--danger); border-radius: 6px; background: #f8514922;
}
</style>
</head>
<body>
<header>
  <h1>Tunnel traffic</h1>
  <p>Left: request list. Right: request/response for the <strong>latest</strong> capture until you click <strong>Show</strong> on a row. Replay updates the Response panel.</p>
</header>
<div class="shell">
  <aside class="col-left">
    <div class="left-topbar">
      <button type="button" class="btn btn-sm" id="btn-latest" title="Show most recent in the right panel">Latest</button>
      <button type="button" class="btn btn-sm" id="clear-all">Clear all</button>
    </div>
    <div id="log-list"></div>
  </aside>
  <section class="col-right">
    <div class="detail-toolbar">
      <span id="detail-badge">No captures yet.</span>
      <div class="toolbar-actions">
        <button type="button" class="btn btn-sm btn-primary" id="btn-mod-replay" disabled>Modify</button>
        <button type="button" class="btn btn-sm" id="btn-reset-req" disabled>Reset</button>
      </div>
    </div>
    <div class="section-title">Request</div>
    <div id="req-slot"><p style="color:var(--muted);font-size:13px;margin:0">Waiting for traffic…</p></div>
    <div class="section-title"><span id="resp-label">Response</span></div>
    <div id="resp-slot"></div>
  </section>
</div>
<script>
(function () {
  var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  var ws = new WebSocket(proto + '//' + location.host + '/ws');
  var logList = document.getElementById('log-list');
  var clearAllBtn = document.getElementById('clear-all');
  var btnLatest = document.getElementById('btn-latest');
  var detailBadge = document.getElementById('detail-badge');
  var reqSlot = document.getElementById('req-slot');
  var respSlot = document.getElementById('resp-slot');
  var respLabel = document.getElementById('resp-label');
  var btnModReplay = document.getElementById('btn-mod-replay');
  var btnResetReq = document.getElementById('btn-reset-req');
  var originals = {};
  var latestId = null;
  var focusedId = null;
  var editing = false;
  var respIsReplay = false;

  function effectiveId() {
    return focusedId || latestId;
  }
  function updateCardSelection() {
    var nodes = logList.querySelectorAll('.log-card');
    var i;
    for (i = 0; i < nodes.length; i++) {
      nodes[i].classList.toggle('selected', focusedId && nodes[i].getAttribute('data-id') === focusedId);
    }
  }
  function updateBadge() {
    var id = effectiveId();
    if (!id || !originals[id]) {
      detailBadge.textContent = 'No captures yet.';
      return;
    }
    var log = originals[id];
    var verb = focusedId ? 'Selected' : 'Latest';
    detailBadge.textContent = verb + ' · ' + log.method + ' ' + log.path + ' · ' + String(log.status);
  }

  function escapeHtml(s) {
    if (s == null) return '';
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }
  function escapeAttr(s) {
    return String(s == null ? '' : s).replace(/&/g,'&amp;').replace(/"/g,'&quot;').replace(/</g,'&lt;');
  }
  function headerRows(h) {
    if (!h || typeof h !== 'object') return '';
    var keys = Object.keys(h).sort();
    var i, k, parts = [];
    for (i = 0; i < keys.length; i++) {
      k = keys[i];
      parts.push('<tr><th>' + escapeHtml(k) + '</th><td class="mono">' + escapeHtml(Array.isArray(h[k]) ? h[k].join(', ') : String(h[k])) + '</td></tr>');
    }
    return parts.join('');
  }
  function renderReqView(log) {
    return '<table class="kv"><tbody>' +
      '<tr><th>Method</th><td class="mono">' + escapeHtml(log.method) + '</td></tr>' +
      '<tr><th>Path</th><td class="mono">' + escapeHtml(log.path) + '</td></tr>' +
      headerRows(log.headers) +
      '<tr><th>Body</th><td><pre>' + escapeHtml(log.body != null ? String(log.body) : '') + '</pre></td></tr>' +
      '</tbody></table>';
  }
  function renderReqEdit(log) {
    var h = log.headers || {};
    return '<table class="kv kv-edit"><tbody>' +
      '<tr><th>Method</th><td><input type="text" class="f-method" value="' + escapeAttr(log.method) + '"/></td></tr>' +
      '<tr><th>Path</th><td><input type="text" class="f-path" value="' + escapeAttr(log.path) + '"/></td></tr>' +
      '<tr><th>Headers</th><td><textarea class="f-headers" rows="5">' + escapeHtml(JSON.stringify(h, null, 2)) + '</textarea></td></tr>' +
      '<tr><th>Body</th><td><textarea class="f-body" rows="6">' + escapeHtml(log.body != null ? String(log.body) : '') + '</textarea></td></tr>' +
      '</tbody></table>';
  }
  function renderRespTable(log) {
    return '<table class="kv"><tbody>' +
      '<tr><th>Status</th><td class="mono">' + escapeHtml(String(log.status)) + '</td></tr>' +
      headerRows(log.resp_headers) +
      '<tr><th>Body</th><td><pre>' + escapeHtml(log.resp_body != null ? String(log.resp_body) : '') + '</pre></td></tr>' +
      '</tbody></table>';
  }
  function renderReplayKv(data) {
    if (data.error) {
      return '<div class="err-box">' + escapeHtml(String(data.error)) + '</div>';
    }
    return '<table class="kv"><tbody>' +
      '<tr><th>Status</th><td class="mono">' + escapeHtml(String(data.status)) + '</td></tr>' +
      headerRows(data.headers) +
      '<tr><th>Body</th><td><pre>' + escapeHtml(data.body != null ? String(data.body) : '') + '</pre></td></tr>' +
      '</tbody></table>';
  }
  function snapshotLog(log) {
    try {
      return JSON.parse(JSON.stringify(log));
    } catch (e) {
      return log;
    }
  }
  function syncModReplayBtn() {
    var ed = !!reqSlot.querySelector('.kv-edit');
    btnModReplay.textContent = ed ? 'Replay' : 'Modify';
  }
  function getPayloadFromPanel() {
    var id = effectiveId();
    if (!id || !originals[id]) throw new Error('no selection');
    var edit = reqSlot.querySelector('.kv-edit');
    if (edit) {
      var method = reqSlot.querySelector('.f-method').value;
      var path = reqSlot.querySelector('.f-path').value;
      var body = reqSlot.querySelector('.f-body').value;
      var headersRaw = reqSlot.querySelector('.f-headers').value.trim();
      var headers = {};
      if (headersRaw) headers = JSON.parse(headersRaw);
      return { method: method, path: path, headers: headers, body: body };
    }
    var o = originals[id];
    return {
      method: o.method,
      path: o.path,
      headers: o.headers || {},
      body: o.body != null ? String(o.body) : ''
    };
  }
  function paint() {
    var id = effectiveId();
    updateBadge();
    updateCardSelection();
    if (!id || !originals[id]) {
      reqSlot.innerHTML = '<p style="color:var(--muted);font-size:13px;margin:0">Waiting for traffic…</p>';
      respSlot.innerHTML = '';
      respLabel.textContent = 'Response';
      btnModReplay.disabled = true;
      btnResetReq.disabled = true;
      return;
    }
    btnModReplay.disabled = false;
    btnResetReq.disabled = false;
    var log = originals[id];
    if (!editing) {
      reqSlot.innerHTML = renderReqView(log);
    } else {
      reqSlot.innerHTML = renderReqEdit(log);
    }
    if (!respIsReplay) {
      respSlot.innerHTML = renderRespTable(log);
      respLabel.textContent = 'Response';
    }
    syncModReplayBtn();
  }
  function wireCardHead(card, id) {
    card.querySelector('.btn-toggle').addEventListener('click', function (e) {
      e.stopPropagation();
      focusedId = id;
      editing = false;
      respIsReplay = false;
      paint();
    });
  }
  function appendLogCard(log) {
    var id = log.id || ('tmp-' + Date.now() + '-' + Math.random());
    log.id = id;
    originals[id] = snapshotLog(log);
    latestId = id;
    var card = document.createElement('article');
    card.className = 'log-card';
    card.setAttribute('data-id', id);
    card.innerHTML =
      '<div class="log-card-head">' +
        '<span class="method">' + escapeHtml(log.method) + '</span>' +
        '<span class="path" title="' + escapeAttr(log.path) + '">' + escapeHtml(log.path) + '</span>' +
        '<span class="status">' + escapeHtml(String(log.status)) + '</span>' +
        '<span class="ms">' + escapeHtml(String(log.duration_ms != null ? log.duration_ms : '')) + ' ms</span>' +
        '<button type="button" class="btn-toggle">Show</button>' +
      '</div>';
    wireCardHead(card, id);
    logList.insertBefore(card, logList.firstChild);
    if (focusedId === null) {
      editing = false;
      respIsReplay = false;
      paint();
    }
    return card;
  }
  function clearAll() {
    logList.innerHTML = '';
    originals = {};
    latestId = null;
    focusedId = null;
    editing = false;
    respIsReplay = false;
    paint();
  }
  function loadHistory() {
    fetch('/logs').then(function (res) { return res.json(); }).then(function (logs) {
      if (!Array.isArray(logs) || !logs.length) return;
      var i;
      for (i = 0; i < logs.length; i++) appendLogCard(logs[i]);
    }).catch(function () {});
  }

  btnLatest.onclick = function () {
    focusedId = null;
    editing = false;
    respIsReplay = false;
    paint();
  };
  btnModReplay.onclick = function () {
    var id = effectiveId();
    if (!id || !originals[id]) return;
    if (btnModReplay.textContent === 'Modify') {
      editing = true;
      paint();
      return;
    }
    var payload;
    try {
      payload = getPayloadFromPanel();
    } catch (err) {
      respIsReplay = true;
      respLabel.textContent = 'Response';
      respSlot.innerHTML = '<div class="err-box">' + escapeHtml(err.message) + '</div>';
      return;
    }
    if (typeof payload.method !== 'string' || !payload.method.trim()) {
      respIsReplay = true;
      respSlot.innerHTML = '<div class="err-box">method required</div>';
      return;
    }
    if (typeof payload.path !== 'string') {
      respIsReplay = true;
      respSlot.innerHTML = '<div class="err-box">path required</div>';
      return;
    }
    btnModReplay.disabled = true;
    respIsReplay = true;
    respLabel.textContent = 'Response (replay)';
    respSlot.innerHTML = '<p style="color:var(--muted);margin:0">…</p>';
    fetch('/replay', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        method: payload.method,
        path: payload.path,
        headers: payload.headers || {},
        body: payload.body != null ? payload.body : ''
      })
    }).then(function (res) { return res.json(); })
      .then(function (data) {
        respSlot.innerHTML = renderReplayKv(data);
      })
      .catch(function (err) {
        respSlot.innerHTML = '<div class="err-box">' + escapeHtml(String(err)) + '</div>';
      })
      .finally(function () {
        btnModReplay.disabled = false;
      });
  };
  btnResetReq.onclick = function () {
    var id = effectiveId();
    if (!id || !originals[id]) return;
    editing = false;
    respIsReplay = false;
    paint();
  };
  reqSlot.addEventListener('input', function () {
    syncModReplayBtn();
  });

  loadHistory();
  paint();

  ws.onmessage = function (event) {
    appendLogCard(JSON.parse(event.data));
  };

  clearAllBtn.onclick = clearAll;
})();
</script>
</body>
</html>`
	w.Write([]byte(page))
}

func serveInspectorLogs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(GetLogs())
}

func handleInspectorWS(hub *inspectorHub, w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	hub.register(conn)

	// Drain pings / detect disconnect; broadcast is server → client only.
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			hub.unregister(conn)
			return
		}
	}
}
