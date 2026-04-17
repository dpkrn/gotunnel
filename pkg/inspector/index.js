/**
 * Traffic inspector UI: WebSocket live feed, history, replay, URL ↔ query params sync.
 */
(function () {
  'use strict';

  // --- State ---
  var entries = Object.create(null);
  var selectedId = null;
  var ws;
  var lastRespRaw = '';
  var hasResponseContent = false;
  /** When true, skip URL→params blur sync (avoid fighting params→URL updates). */
  var syncingUrlFromParams = false;

  var el = {
    logList: document.getElementById('logList'),
    sidebar: document.getElementById('sidebar'),
    resizerH: document.getElementById('resizerH'),
    resizerV: document.getElementById('resizerV'),
    paneReq: document.getElementById('paneReq'),
    paneResp: document.getElementById('paneResp'),
    splitVWrap: document.getElementById('splitVWrap'),
    method: document.getElementById('method'),
    url: document.getElementById('url'),
    reqBody: document.getElementById('reqBody'),
    paramsTbody: document.getElementById('params-tbody'),
    headersTbody: document.getElementById('headers-tbody'),
    paramsBadge: document.getElementById('params-badge'),
    headersBadge: document.getElementById('headers-badge'),
    authType: document.getElementById('authType'),
    authToken: document.getElementById('authToken'),
    authBearerBlock: document.getElementById('auth-bearer-block'),
    authBasicBlock: document.getElementById('auth-basic-block'),
    authApikeyBlock: document.getElementById('auth-apikey-block'),
    authBasicUser: document.getElementById('authBasicUser'),
    authBasicPass: document.getElementById('authBasicPass'),
    authApiKeyName: document.getElementById('authApiKeyName'),
    authApiKeyValue: document.getElementById('authApiKeyValue'),
    scriptsPre: document.getElementById('scriptsPre'),
    scriptsPost: document.getElementById('scriptsPost'),
    respBody: document.getElementById('respBody'),
    respMeta: document.getElementById('respMeta'),
    targetBase: document.getElementById('targetBase'),
    wsStatus: document.getElementById('wsStatus'),
    btnReplay: document.getElementById('btnReplay'),
    btnReset: document.getElementById('btnReset'),
    originHint: document.getElementById('origin-hint'),
    themeSelect: document.getElementById('themeSelect'),
    emptyResponse: document.getElementById('emptyResponse'),
    responseStatus: document.getElementById('responseStatus'),
    statusBadge: document.getElementById('statusBadge'),
    respTime: document.getElementById('respTime'),
    respSize: document.getElementById('respSize'),
    respHeadersTable: document.getElementById('respHeadersTable'),
    respCookies: document.getElementById('respCookies'),
    respConsole: document.getElementById('respConsole'),
    fmtPretty: document.getElementById('fmtPretty'),
    fmtRaw: document.getElementById('fmtRaw'),
    btnCopyResp: document.getElementById('btnCopyResp')
  };

  var TARGET_KEY = 'inspectorReplayBase';
  var SIDEBAR_W = 'inspectorSidebarW';
  var SPLIT_RATIO = 'inspectorSplitRatio';
  var THEME_KEY = 'inspectorTheme';

  /** Digits only; set in inspector.html from the tunnel’s forward port when embedded. */
  function localAppPortForDefault() {
    var p =
      typeof window !== 'undefined' && window.__LOCAL_APP_PORT__
        ? String(window.__LOCAL_APP_PORT__).trim()
        : '';
    if (!/^\d+$/.test(p)) p = '8080';
    return p;
  }

  /** Used when “Replay base” is empty; same host the tunnel uses for your local app (not the inspector UI port). */
  var DEFAULT_REPLAY_BASE = 'http://localhost:' + localAppPortForDefault();

  /** Sidebar and URL bar: path + query only, never http://host:port (avoids showing the inspector origin). */
  function requestPathForDisplay(s) {
    if (s == null || s === '') return '/';
    s = String(s).trim();
    if (!s) return '/';
    if (/^https?:\/\//i.test(s)) {
      try {
        var u = new URL(s);
        return u.pathname + u.search + (u.hash || '');
      } catch (e) {
        return s;
      }
    }
    return s.startsWith('/') ? s : '/' + s;
  }

  function replayBaseURL() {
    var b = (el.targetBase.value || '').trim().replace(/\/$/, '');
    return b || DEFAULT_REPLAY_BASE;
  }

  /** Full http(s) URL for replay and for URL(). Relative paths like /name?q=1 become http://host:port/name?q=1 */
  function absolutizeForReplay(urlStr) {
    var s = (urlStr || '').trim();
    if (!s) return s;
    if (/^https?:\/\//i.test(s)) return s;
    var base = replayBaseURL();
    return base + (s.startsWith('/') ? s : '/' + s);
  }

  function parseURLWithReplayBase(urlStr) {
    var s = (urlStr || '').trim();
    if (!s) throw new Error('empty url');
    if (/^https?:\/\//i.test(s)) return new URL(s);
    return new URL(s, replayBaseURL() + '/');
  }

  function inspectorHTTPBase() {
    var q = new URLSearchParams(location.search);
    var api = q.get('api');
    if (api) return api.replace(/\/$/, '');
    var port = q.get('port');
    if (port) return 'http://127.0.0.1:' + port;
    if (location.host) return location.protocol + '//' + location.host;
    return 'http://127.0.0.1:4040';
  }

  function initTheme() {
    if (!el.themeSelect) return;
    var t = localStorage.getItem(THEME_KEY) || 'postman';
    if (t !== 'postman' && t !== 'terminal') t = 'postman';
    document.documentElement.setAttribute('data-theme', t);
    el.themeSelect.value = t;
    el.themeSelect.addEventListener('change', function () {
      var v = el.themeSelect.value;
      document.documentElement.setAttribute('data-theme', v);
      localStorage.setItem(THEME_KEY, v);
    });
  }

  function bytesToUtf8(body) {
    if (body == null || body === '') return '';
    if (typeof body === 'string') {
      try {
        var bin = atob(body);
        var bytes = new Uint8Array(bin.length);
        for (var i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
        return new TextDecoder().decode(bytes);
      } catch (e) {
        return body;
      }
    }
    return String(body);
  }

  function prettyJSON(raw) {
    if (raw == null || raw === '') return '';
    try {
      return JSON.stringify(JSON.parse(raw), null, 2);
    } catch (e) {
      return raw;
    }
  }

  function formatBodySize(bodyField) {
    if (bodyField == null) return '—';
    if (typeof bodyField === 'string') {
      try {
        var bin = atob(bodyField);
        return formatBytes(bin.length);
      } catch (e) {
        return formatBytes(new TextEncoder().encode(bodyField).length);
      }
    }
    return '—';
  }

  function formatBytes(n) {
    if (n < 1024) return n + ' B';
    if (n < 1024 * 1024) return (n / 1024).toFixed(2) + ' KB';
    return (n / (1024 * 1024)).toFixed(2) + ' MB';
  }

  function clearTbody(tb) {
    tb.innerHTML = '';
  }

  function bindTableRow(tr, tbody) {
    var isParams = tbody === el.paramsTbody;
    var del = tr.querySelector('.param-del');
    if (del) {
      del.addEventListener('click', function () {
        tr.remove();
        ensureTrailingEmptyRow(tbody);
        updateBadges();
        if (isParams) syncUrlSearchFromParams();
      });
    }
    tr.querySelectorAll('.param-input, .param-check input').forEach(function (inp) {
      inp.addEventListener('input', function () {
        updateBadges();
        if (isParams) syncUrlSearchFromParams();
      });
      inp.addEventListener('change', function () {
        updateBadges();
        if (isParams) syncUrlSearchFromParams();
      });
    });
  }

  function makeRow(key, value, checked, desc) {
    var tr = document.createElement('tr');
    tr.innerHTML =
      '<td class="param-check"><input type="checkbox"' + (checked ? ' checked' : '') + '></td>' +
      '<td><input class="param-input" type="text" placeholder="Key" value="' + escapeAttr(key) + '"></td>' +
      '<td><input class="param-input" type="text" placeholder="Value" value="' + escapeAttr(value) + '"></td>' +
      '<td><input class="param-input desc" type="text" placeholder="Description" value="' + escapeAttr(desc) + '"></td>' +
      '<td><button type="button" class="param-del" aria-label="Remove">×</button></td>';
    return tr;
  }

  function escapeAttr(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;')
      .replace(/"/g, '&quot;')
      .replace(/</g, '&lt;');
  }

  function ensureTrailingEmptyRow(tbody) {
    var rows = tbody.querySelectorAll('tr');
    var last = rows[rows.length - 1];
    if (!last) {
      tbody.appendChild(makeRow('', '', false, ''));
      bindTableRow(tbody.lastElementChild, tbody);
      return;
    }
    var inputs = last.querySelectorAll('.param-input');
    var key = inputs[0] && inputs[0].value.trim();
    var hasMore = rows.length > 1;
    if (key || !hasMore) {
      tbody.appendChild(makeRow('', '', false, ''));
      bindTableRow(tbody.lastElementChild, tbody);
    }
  }

  function addParamRow(key, value, checked, desc) {
    var tr = makeRow(key, value, checked, desc);
    el.paramsTbody.appendChild(tr);
    bindTableRow(tr, el.paramsTbody);
    updateBadges();
  }

  function addHeaderRow(key, value, checked, desc) {
    var tr = makeRow(key, value, checked, desc);
    el.headersTbody.appendChild(tr);
    bindTableRow(tr, el.headersTbody);
    updateBadges();
  }

  /** Rewrite only the query string from the Params table (path + origin unchanged). */
  function syncUrlSearchFromParams() {
    if (syncingUrlFromParams) return;
    var raw = el.url.value.trim();
    if (!raw) return;
    try {
      var u = parseURLWithReplayBase(raw);
      var rows = readKeyValueRows(el.paramsTbody).filter(function (r) {
        return r.enabled && r.key;
      });
      var q = new URLSearchParams();
      rows.forEach(function (r) {
        q.append(r.key, r.value);
      });
      u.search = q.toString();
      syncingUrlFromParams = true;
      el.url.value = u.pathname + u.search + (u.hash || '');
      syncingUrlFromParams = false;
    } catch (e) {}
  }

  function fillParamsFromURL(urlStr) {
    if (syncingUrlFromParams) return;
    clearTbody(el.paramsTbody);
    try {
      var u = parseURLWithReplayBase(urlStr);
      u.searchParams.forEach(function (val, key) {
        var tr = makeRow(key, val, true, '');
        el.paramsTbody.appendChild(tr);
        bindTableRow(tr, el.paramsTbody);
      });
    } catch (e) {}
    ensureTrailingEmptyRow(el.paramsTbody);
    updateBadges();
  }

  function fillHeadersTable(h) {
    clearTbody(el.headersTbody);
    if (h && typeof h === 'object') {
      Object.keys(h).forEach(function (k) {
        (h[k] || []).forEach(function (v) {
          var tr = makeRow(k, v, true, '');
          el.headersTbody.appendChild(tr);
          bindTableRow(tr, el.headersTbody);
        });
      });
    }
    ensureTrailingEmptyRow(el.headersTbody);
    updateBadges();
  }

  function readKeyValueRows(tbody) {
    var out = [];
    tbody.querySelectorAll('tr').forEach(function (tr) {
      var cb = tr.querySelector('.param-check input');
      var inputs = tr.querySelectorAll('.param-input');
      var key = inputs[0] && inputs[0].value.trim();
      var val = inputs[1] ? inputs[1].value : '';
      out.push({ enabled: cb && cb.checked, key: key, value: val });
    });
    return out;
  }

  function headersFromTable() {
    var h = {};
    readKeyValueRows(el.headersTbody).forEach(function (r) {
      if (!r.enabled || !r.key) return;
      if (!h[r.key]) h[r.key] = [];
      h[r.key].push(r.value);
    });
    return h;
  }

  function mergeAuthHeaders(headers) {
    var t = el.authType.value;
    var copy = JSON.parse(JSON.stringify(headers));
    if (t === 'bearer' && el.authToken.value.trim()) {
      copy['Authorization'] = ['Bearer ' + el.authToken.value.trim()];
    } else if (t === 'basic' && el.authBasicUser.value) {
      var raw = el.authBasicUser.value + ':' + (el.authBasicPass.value || '');
      copy['Authorization'] = ['Basic ' + btoa(unescape(encodeURIComponent(raw)))];
    } else if (t === 'apikey' && el.authApiKeyName.value.trim()) {
      copy[el.authApiKeyName.value.trim()] = [el.authApiKeyValue.value || ''];
    }
    return copy;
  }

  function buildUrlWithQueryParams(urlStr) {
    var rows = readKeyValueRows(el.paramsTbody).filter(function (r) {
      return r.enabled && r.key;
    });
    var u;
    try {
      u = new URL(absolutizeForReplay(urlStr));
    } catch (e) {
      return absolutizeForReplay(urlStr);
    }
    var q = new URLSearchParams();
    rows.forEach(function (r) {
      q.append(r.key, r.value);
    });
    u.search = q.toString();
    return u.href;
  }

  function updateBadges() {
    var p = readKeyValueRows(el.paramsTbody).filter(function (r) {
      return r.enabled && r.key;
    }).length;
    var h = readKeyValueRows(el.headersTbody).filter(function (r) {
      return r.enabled && r.key;
    }).length;
    el.paramsBadge.textContent = String(p);
    el.headersBadge.textContent = String(h);
  }

  function collectReplayPayload() {
    var headers = mergeAuthHeaders(headersFromTable());
    return {
      method: el.method.value,
      url: buildUrlWithQueryParams(el.url.value.trim()),
      headers: headers,
      body: el.reqBody.value
    };
  }

  function durationOf(entry) {
    if (entry.durationMs != null) return entry.durationMs;
    if (entry.response && entry.response.durationMs != null) return entry.response.durationMs;
    return '';
  }

  function methodClass(m) {
    m = (m || 'GET').toUpperCase();
    if (['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].indexOf(m) < 0) return 'GET';
    return m;
  }

  function statusClass(code) {
    var c = parseInt(code, 10);
    if (c >= 200 && c < 300) return 'st-2xx';
    if (c >= 400) return 'st-4xx';
    return '';
  }

  function findLiByDataId(id) {
    var lis = el.logList.getElementsByTagName('li');
    for (var i = 0; i < lis.length; i++) {
      if (lis[i].getAttribute('data-id') === id) return lis[i];
    }
    return null;
  }

  function upsertListItem(entry) {
    if (!entry || !entry.id) return;
    entries[entry.id] = entry;
    var li = findLiByDataId(entry.id);
    var req = entry.request || {};
    var res = entry.response || {};
    var ms = durationOf(entry);
    var st = res.statusCode != null ? res.statusCode : '—';
    var disp = requestPathForDisplay(req.path || '/');
    var html =
      '<span class="m ' + methodClass(req.method) + '">' + escapeHtml((req.method || '—').toUpperCase()) + '</span>' +
      '<span class="path" title="' + escapeHtml(disp) + '">' + escapeHtml(disp) + '</span>' +
      '<span class="ms">' + escapeHtml(ms !== '' ? ms + 'ms' : '—') + '</span>' +
      '<span class="st ' + statusClass(st) + '">' + escapeHtml(st) + '</span>';

    if (li) {
      li.innerHTML = html;
      return;
    }
    li = document.createElement('li');
    li.setAttribute('data-id', entry.id);
    li.innerHTML = html;
    li.addEventListener('click', function () {
      selectEntry(entry.id);
    });
    el.logList.insertBefore(li, el.logList.firstChild);
  }

  function escapeHtml(s) {
    if (s == null || s === undefined) return '';
    return String(s)
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;');
  }

  function selectEntry(id) {
    selectedId = id;
    Array.prototype.forEach.call(el.logList.querySelectorAll('li'), function (li) {
      li.classList.toggle('active', li.getAttribute('data-id') === id);
    });
    loadEntryIntoEditors(id);
  }

  function loadEntryIntoEditors(id) {
    var entry = entries[id];
    if (!entry) {
      fetch(inspectorHTTPBase() + '/log?id=' + encodeURIComponent(id))
        .then(function (r) { return r.json(); })
        .then(function (data) {
          entries[id] = data;
          applyEntryToEditors(data);
        })
        .catch(function () {});
      return;
    }
    applyEntryToEditors(entry);
  }

  /** Copy Authorization into auth panel and return headers without Authorization for the table. */
  function authFromHeadersAndStrip(headers) {
    el.authType.value = 'none';
    el.authToken.value = '';
    el.authBasicUser.value = '';
    el.authBasicPass.value = '';
    el.authApiKeyName.value = '';
    el.authApiKeyValue.value = '';
    var h = headers && typeof headers === 'object' ? JSON.parse(JSON.stringify(headers)) : {};
    var raw = (h.Authorization || h.authorization || [])[0] || '';
    delete h.Authorization;
    delete h.authorization;
    if (raw.indexOf('Bearer ') === 0) {
      el.authType.value = 'bearer';
      el.authToken.value = raw.slice(7).trim();
    } else if (raw.indexOf('Basic ') === 0) {
      try {
        var dec = atob(raw.slice(6).trim());
        var ix = dec.indexOf(':');
        el.authType.value = 'basic';
        el.authBasicUser.value = ix >= 0 ? dec.slice(0, ix) : dec;
        el.authBasicPass.value = ix >= 0 ? dec.slice(ix + 1) : '';
      } catch (e) {}
    }
    showAuthBlocks();
    return h;
  }

  function applyEntryToEditors(entry) {
    var req = entry.request || {};
    var res = entry.response || {};
    el.method.value = (req.method || 'GET').toUpperCase();
    var path = requestPathForDisplay(req.path || '/');
    el.url.value = path;
    fillParamsFromURL(absolutizeForReplay(path));
    fillHeadersTable(authFromHeadersAndStrip(req.headers));
    el.reqBody.value = bytesToUtf8(req.body);

    var capMs = durationOf(entry);
    el.respMeta.innerHTML =
      'Captured: <strong>' + escapeHtml(capMs !== '' ? capMs + ' ms' : '—') + '</strong>' +
      ' · status <strong>' + escapeHtml(res.statusCode != null ? res.statusCode : '—') + '</strong>';

    showResponsePanels(res, capMs, 'captured');
    logConsole('Loaded request ' + escapeHtml(entry.id) + ' from history.');
  }

  function fillRespHeadersTable(headers) {
    el.respHeadersTable.innerHTML = '';
    if (!headers) return;
    Object.keys(headers).forEach(function (k) {
      (headers[k] || []).forEach(function (v) {
        var tr = document.createElement('tr');
        tr.innerHTML =
          '<td class="kv-key">' + escapeHtml(k) + '</td>' +
          '<td>' + escapeHtml(v) + '</td>';
        el.respHeadersTable.appendChild(tr);
      });
    });
  }

  function parseSetCookies(headers) {
    if (!headers) return [];
    var raw = headers['Set-Cookie'] || headers['set-cookie'];
    if (!raw) return [];
    return raw;
  }

  function showResponsePanels(res, durationMs, kind) {
    el.emptyResponse.setAttribute('hidden', '');
    el.responseStatus.hidden = false;
    var code = res.statusCode != null ? res.statusCode : '—';
    el.statusBadge.textContent = String(code);
    el.statusBadge.className = 'status-badge';
    var c = parseInt(code, 10);
    if (!isNaN(c)) {
      if (c >= 200 && c < 300) el.statusBadge.classList.add('st-2xx');
      else if (c >= 400) el.statusBadge.classList.add('st-err');
    }
    el.respTime.textContent = durationMs !== '' && durationMs != null ? durationMs + ' ms' : '—';
    el.respSize.textContent = formatBodySize(res.body);

    fillRespHeadersTable(res.headers);
    var cookies = parseSetCookies(res.headers);
    if (cookies && cookies.length) {
      el.respCookies.innerHTML = cookies.map(function (c) {
        return '<div class="cookie-line">' + escapeHtml(c) + '</div>';
      }).join('');
    } else {
      el.respCookies.textContent = 'No Set-Cookie headers on this response.';
    }

    lastRespRaw = bytesToUtf8(res.body);
    setBodyFormatPretty();
    hasResponseContent = true;
    switchRespTab('body');
  }

  function setBodyFormatPretty() {
    el.fmtPretty.classList.add('active');
    el.fmtRaw.classList.remove('active');
    el.respBody.value = prettyJSON(lastRespRaw);
  }

  function setBodyFormatRaw() {
    el.fmtRaw.classList.add('active');
    el.fmtPretty.classList.remove('active');
    el.respBody.value = lastRespRaw;
  }

  function showEmptyResponseState() {
    hasResponseContent = false;
    el.emptyResponse.removeAttribute('hidden');
    el.responseStatus.hidden = true;
    el.respMeta.textContent = '';
    lastRespRaw = '';
    el.respBody.value = '';
    el.respHeadersTable.innerHTML = '';
    el.respCookies.textContent = 'No cookies parsed (Set-Cookie appears in Headers).';
    document.getElementById('resp-panel-body').hidden = true;
    document.getElementById('resp-panel-headers').hidden = true;
    document.getElementById('resp-panel-cookies').hidden = true;
    document.getElementById('resp-panel-console').hidden = true;
  }

  function reloadSelectionFromCapture() {
    if (!selectedId) return;
    delete entries[selectedId];
    loadEntryIntoEditors(selectedId);
    logConsole('Reset: reloaded captured request from server.');
  }

  function logConsole(line) {
    var t = new Date().toISOString();
    el.respConsole.textContent += '[' + t + '] ' + line + '\n';
  }

  function switchConfigPanel(panel) {
    document.querySelectorAll('.config-tab').forEach(function (tab) {
      tab.classList.toggle('active', tab.getAttribute('data-panel') === panel);
    });
    document.querySelectorAll('.config-panel').forEach(function (p) {
      p.hidden = p.id !== 'panel-' + panel;
    });
  }

  function switchRespTab(name) {
    document.querySelectorAll('.resp-tab').forEach(function (tab) {
      tab.classList.toggle('active', tab.getAttribute('data-resp') === name);
    });
    if (!hasResponseContent) return;
    document.getElementById('resp-panel-body').hidden = name !== 'body';
    document.getElementById('resp-panel-headers').hidden = name !== 'headers';
    document.getElementById('resp-panel-cookies').hidden = name !== 'cookies';
    document.getElementById('resp-panel-console').hidden = name !== 'console';
  }

  function showAuthBlocks() {
    var t = el.authType.value;
    el.authBearerBlock.hidden = t !== 'bearer';
    el.authBasicBlock.hidden = t !== 'basic';
    el.authApikeyBlock.hidden = t !== 'apikey';
  }

  function connectWS() {
    var base = inspectorHTTPBase();
    var wsProto = base.indexOf('https:') === 0 ? 'wss:' : 'ws:';
    var host = base.replace(/^https?:\/\//, '');
    ws = new WebSocket(wsProto + '//' + host + '/ws');

    ws.onopen = function () {
      el.wsStatus.classList.add('live');
      fetch(base + '/logs')
        .then(function (r) { return r.json(); })
        .then(function (logs) {
          el.logList.innerHTML = '';
          entries = Object.create(null);
          for (var i = logs.length - 1; i >= 0; i--) upsertListItem(logs[i]);
          if (logs.length) {
            selectEntry(logs[logs.length - 1].id);
          } else {
            showEmptyResponseState();
          }
        });
    };

    ws.onmessage = function (ev) {
      var msg = JSON.parse(ev.data);
      if (msg.eventType === 'request' && msg.payload && msg.payload.id) {
        upsertListItem(msg.payload);
        selectEntry(msg.payload.id);
      }
    };

    ws.onclose = function () {
      el.wsStatus.classList.remove('live');
      setTimeout(connectWS, 1000);
    };
    ws.onerror = function () {};
  }

  function setupResizerH() {
    if (!el.resizerH || !el.sidebar) return;
    var w = localStorage.getItem(SIDEBAR_W);
    if (w) el.sidebar.style.width = w + 'px';
    el.resizerH.addEventListener('mousedown', function (e) {
      e.preventDefault();
      el.resizerH.classList.add('is-dragging');
      var prevCursor = document.body.style.cursor;
      document.body.style.cursor = 'ew-resize';
      var startX = e.clientX;
      var startW = el.sidebar.offsetWidth;
      function move(ev) {
        ev.preventDefault();
        var nw = Math.max(180, Math.min(startW + (ev.clientX - startX), window.innerWidth * 0.55));
        el.sidebar.style.width = nw + 'px';
      }
      function up() {
        el.resizerH.classList.remove('is-dragging');
        document.body.style.cursor = prevCursor;
        localStorage.setItem(SIDEBAR_W, String(el.sidebar.offsetWidth));
        document.removeEventListener('mousemove', move);
        document.removeEventListener('mouseup', up);
      }
      document.addEventListener('mousemove', move);
      document.addEventListener('mouseup', up);
    });
  }

  function setupResizerV() {
    if (!el.resizerV || !el.splitVWrap || !el.paneReq || !el.paneResp) return;

    function getRatio() {
      var ratio = parseFloat(localStorage.getItem(SPLIT_RATIO) || '0.45', 10);
      if (isNaN(ratio) || ratio < 0.15 || ratio > 0.85) return 0.45;
      return ratio;
    }

    function applySplitRatio() {
      var ratio = getRatio();
      var h = el.splitVWrap.clientHeight;
      if (h < 80) return;
      var topH = Math.round(h * ratio);
      topH = Math.max(120, Math.min(topH, h - 160));
      el.paneReq.style.flex = '0 0 ' + topH + 'px';
      el.paneResp.style.flex = '1 1 0%';
      el.paneResp.style.minHeight = '0';
    }

    applySplitRatio();
    window.addEventListener('resize', applySplitRatio);
    window.addEventListener('load', applySplitRatio);
    requestAnimationFrame(function () {
      applySplitRatio();
      requestAnimationFrame(applySplitRatio);
    });

    el.resizerV.addEventListener('mousedown', function (e) {
      e.preventDefault();
      el.resizerV.classList.add('is-dragging');
      var prevCursor = document.body.style.cursor;
      document.body.style.cursor = 'ns-resize';
      var startY = e.clientY;
      var startTop = el.paneReq.getBoundingClientRect().height;
      function move(ev) {
        ev.preventDefault();
        var wrapH = el.splitVWrap.clientHeight;
        if (wrapH < 80) return;
        var dy = ev.clientY - startY;
        var nh = Math.max(120, Math.min(startTop + dy, wrapH - 160));
        el.paneReq.style.flex = '0 0 ' + nh + 'px';
      }
      function up() {
        el.resizerV.classList.remove('is-dragging');
        document.body.style.cursor = prevCursor;
        var wrapH = el.splitVWrap.clientHeight;
        if (wrapH >= 80) {
          var rect = el.paneReq.getBoundingClientRect();
          var ratio = Math.max(0.15, Math.min(0.85, rect.height / wrapH));
          localStorage.setItem(SPLIT_RATIO, String(ratio));
        }
        document.removeEventListener('mousemove', move);
        document.removeEventListener('mouseup', up);
      }
      document.addEventListener('mousemove', move);
      document.addEventListener('mouseup', up);
    });
  }

  document.querySelectorAll('.config-tab').forEach(function (tab) {
    tab.addEventListener('click', function () {
      switchConfigPanel(tab.getAttribute('data-panel'));
    });
  });

  document.querySelectorAll('.resp-tab').forEach(function (tab) {
    tab.addEventListener('click', function () {
      switchRespTab(tab.getAttribute('data-resp'));
    });
  });

  document.querySelectorAll('.script-subtab').forEach(function (btn) {
    btn.addEventListener('click', function () {
      var which = btn.getAttribute('data-script');
      document.querySelectorAll('.script-subtab').forEach(function (b) {
        b.classList.toggle('active', b === btn);
      });
      el.scriptsPre.hidden = which !== 'pre';
      el.scriptsPost.hidden = which !== 'post';
    });
  });

  el.authType.addEventListener('change', showAuthBlocks);

  document.getElementById('add-param-row').addEventListener('click', function () {
    addParamRow('', '', false, '');
  });
  document.getElementById('add-header-row').addEventListener('click', function () {
    addHeaderRow('', '', false, '');
  });

  el.fmtPretty.addEventListener('click', setBodyFormatPretty);
  el.fmtRaw.addEventListener('click', setBodyFormatRaw);

  el.btnCopyResp.addEventListener('click', function () {
    var t = el.respBody.value;
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(t).then(function () {
        logConsole('Response body copied to clipboard.');
      });
    }
  });

  el.btnReplay.addEventListener('click', function () {
    var payload = collectReplayPayload();
    el.respMeta.textContent = 'Replaying…';
    lastRespRaw = '';
    el.respBody.value = '';
    fetch(inspectorHTTPBase() + '/replay', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
      .then(function (r) {
        return r.text().then(function (text) {
          try {
            return { j: JSON.parse(text) };
          } catch (e) {
            return { j: { error: text || 'bad response' } };
          }
        });
      })
      .then(function (_ref) {
        var j = _ref.j;
        if (j.error) {
          hasResponseContent = true;
          el.emptyResponse.setAttribute('hidden', '');
          el.responseStatus.hidden = false;
          el.statusBadge.textContent = 'Err';
          el.statusBadge.className = 'status-badge st-err';
          el.respTime.textContent = j.durationMs != null ? j.durationMs + ' ms' : '—';
          el.respSize.textContent = '—';
          el.respMeta.innerHTML = 'Replay failed: <strong>' + escapeHtml(j.error) + '</strong>';
          el.respBody.value = '';
          fillRespHeadersTable({});
          logConsole('Replay error: ' + j.error);
          switchRespTab('body');
          return;
        }
        el.respMeta.innerHTML =
          'Live replay: <strong>' + (j.durationMs != null ? j.durationMs + ' ms' : '—') + '</strong>' +
          ' · status <strong>' + escapeHtml(j.statusCode) + '</strong>';
        showResponsePanels(
          { statusCode: j.statusCode, headers: j.headers, body: j.body },
          j.durationMs,
          'replay'
        );
        logConsole('Replay finished: HTTP ' + j.statusCode);
        switchRespTab('body');
      })
      .catch(function (err) {
        el.respMeta.textContent = 'Replay error: ' + err;
        logConsole(String(err));
      });
  });

  el.btnReset.addEventListener('click', reloadSelectionFromCapture);

  el.targetBase.placeholder = DEFAULT_REPLAY_BASE;
  var savedBase = localStorage.getItem(TARGET_KEY);
  if (savedBase) el.targetBase.value = savedBase;
  el.targetBase.addEventListener('change', function () {
    localStorage.setItem(TARGET_KEY, el.targetBase.value);
  });

  el.url.addEventListener('blur', function () {
    if (syncingUrlFromParams) return;
    var t = el.url.value.trim();
    if (t) el.url.value = requestPathForDisplay(t);
    fillParamsFromURL(absolutizeForReplay(el.url.value.trim()));
  });

  if (!location.host) {
    el.originHint.style.display = 'block';
    el.originHint.textContent =
      'Open from the Go inspector (http://127.0.0.1:4040/) or use Live Server with ?port=4040 so API/WebSocket hit the inspector.';
  } else {
    var q = new URLSearchParams(location.search);
    if (!q.get('port') && !q.get('api')) {
      var p = location.port;
      if (p && p !== '4040' && p !== '80' && p !== '443') {
        el.originHint.style.display = 'block';
        el.originHint.textContent =
          'Static server on port ' + p + ': add ?port=4040 (or ?api=http://127.0.0.1:4040) to use the Go inspector for traffic.';
      }
    }
  }

  ensureTrailingEmptyRow(el.paramsTbody);
  ensureTrailingEmptyRow(el.headersTbody);
  showAuthBlocks();
  showEmptyResponseState();

  initTheme();
  setupResizerH();
  setupResizerV();
  connectWS();
})();
