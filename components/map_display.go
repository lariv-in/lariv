package components

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/lariv-in/lago/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Pinned MapLibre for [MapDisplay]; matches seer map plugins.
const mapDisplayLibreCDNVersion = "4.7.1"
const mapDisplayCBORXCDNVersion = "1.6.0"

var mapDisplayIDSanitize = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// MapDisplayLibreHead loads MapLibre CSS and JS from unpkg. Include once per page
// in the shell ExtraHead when using [MapDisplay].
type MapDisplayLibreHead struct {
	Page
}

func (e *MapDisplayLibreHead) GetKey() string     { return e.Key }
func (e *MapDisplayLibreHead) GetRoles() []string { return e.Roles }

func (e *MapDisplayLibreHead) Build(ctx context.Context) Node {
	baseMapLibre := "https://unpkg.com/maplibre-gl@" + mapDisplayLibreCDNVersion + "/dist/"
	baseCBORX := "https://unpkg.com/cbor-x@" + mapDisplayCBORXCDNVersion + "/dist/"
	return Group([]Node{
		Link(Href(baseMapLibre+"maplibre-gl.css"), Rel("stylesheet"), CrossOrigin("anonymous")),
		Script(Src(baseMapLibre+"maplibre-gl.js"), CrossOrigin("anonymous")),
		Script(Src(baseCBORX+"index.js"), CrossOrigin("anonymous")),
	})
}

func mapDisplayIDSuffix(pageKey string) string {
	s := strings.TrimSpace(pageKey)
	if s == "" {
		return "default"
	}
	s = mapDisplayIDSanitize.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "default"
	}
	if len(s) > 48 {
		s = s[:48]
	}
	return s
}

// MapDisplay renders a MapLibre map that opens a WebSocket at DataURL and expects each
// message body to be a JSON array (UTF-8 text frames).
//
// DataURL must be a WebSocket URL (ws: or wss:) or a path beginning with "/" (scheme and host
// are taken from the current page: wss on https, ws on http).
//
// Outbound (client → server): whenever the map viewport changes (pan, zoom, rotate, pitch,
// or container resize), the client sends a CBOR object:
//
//	{"type":"mapDisplayViewport","bounds":{"west":number,"south":number,"east":number,"north":number},"zoom":number}
//
// Longitude/latitude are in degrees from MapLibre’s current map.getBounds().
// Sends only while the WebSocket is OPEN; debounced ~150ms across rapid events.
//
// Inbound (server → client) message payload: CBOR array of objects:
//   - position (required): { "lat": number, "lng": number }
//   - direction (optional): { "x", "y" } unit vector; if set, marker uses arrow icon,
//     rotation follows (x,y) as east/north components, and that marker is not clustered.
//   - velocity (optional): { "x", "y" } in degrees per second (x = d(lng)/dt, y = d(lat)/dt);
//     omitted means zero velocity.
//   - time (optional): Unix timestamp in seconds for position; if omitted, the message
//     receive time is used as the reference time.
//   - link (optional): if non-empty, clicking the marker navigates the page to this URL.
//
// Extrapolated coordinates (seconds): responseTime = wall clock when a message was parsed;
// tRef = time ?? responseTime; lng = position.lng + max(0, now - tRef) * (velocity?.x ?? 0);
// lat = position.lat + max(0, now - tRef) * (velocity?.y ?? 0).
//
// Use [MapDisplayLibreHead] on the same page. Register with a stable Page.Key; use a pointer
// when embedding in parents that need patchability.
type MapDisplay struct {
	Page
	// DataURL is the WebSocket URL (ws/wss or same-origin path like "/app/live/map").
	DataURL getters.Getter[string]
	// RefreshMS if non-nil: milliseconds to wait before reconnecting after the socket closes
	// or errors. Zero uses a 2000ms default; negative disables auto-reconnect.
	RefreshMS getters.Getter[int64]
	// Classes for the map container div (width/height). Empty uses a default tall map box.
	Classes string
}

func (e *MapDisplay) GetKey() string     { return e.Key }
func (e *MapDisplay) GetRoles() []string { return e.Roles }

func (e *MapDisplay) Build(ctx context.Context) Node {
	dataURL := ""
	if e.DataURL != nil {
		u, err := e.DataURL(ctx)
		if err != nil {
			slog.Error("MapDisplay DataURL getter failed", "error", err, "key", e.Key)
			return ContainerError{
				Page:  Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		dataURL = strings.TrimSpace(u)
	}
	if dataURL == "" {
		err := errors.New("MapDisplay: empty DataURL")
		slog.Error("MapDisplay missing DataURL", "key", e.Key)
		return ContainerError{
			Page:  Page{Key: e.Key + ".err"},
			Error: getters.Static(err),
		}.Build(ctx)
	}

	refreshMS := int64(0)
	if e.RefreshMS != nil {
		v, err := e.RefreshMS(ctx)
		if err != nil {
			slog.Error("MapDisplay RefreshMS getter failed", "error", err, "key", e.Key)
			return ContainerError{
				Page:  Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		refreshMS = v
	}

	suffix := mapDisplayIDSuffix(e.Key)
	mapElID := "mapdisplay-" + suffix + "-map"
	dataURLBytes, _ := json.Marshal(dataURL)
	refreshMSBytes, _ := json.Marshal(refreshMS)
	suffixBytes, _ := json.Marshal(suffix)

	classes := strings.TrimSpace(e.Classes)
	if classes == "" {
		classes = "w-full h-[min(80vh,720px)] min-h-80 rounded-box border border-base-300 z-0"
	}

	initJS := `(function(){
  var suffix = ` + string(suffixBytes) + `;
  var dataURL = ` + string(dataURLBytes) + `;
  var refreshMS = ` + string(refreshMSBytes) + `;
  var mapElId = "mapdisplay-" + suffix + "-map";
  var mapEl = document.getElementById(mapElId);
  if (!mapEl || typeof maplibregl === "undefined") { return; }
  var styleLight = "https://demotiles.maplibre.org/style.json";
  var styleDark = "https://tiles.openfreemap.org/styles/dark";
  function themeIsDark() {
    try {
      var t = localStorage.getItem("theme");
      if (t === "dark" || t === "light") { return t === "dark"; }
    } catch (e0) {}
    return document.body && document.body.getAttribute("data-theme") === "dark";
  }
  var lastDark = themeIsDark();
  var map = new maplibregl.Map({
    container: mapEl,
    style: lastDark ? styleDark : styleLight,
    center: [0, 20],
    zoom: 1.5
  });
  map.addControl(new maplibregl.NavigationControl(), "top-right");

  var srcC = "md-" + suffix + "-c-src";
  var srcD = "md-" + suffix + "-d-src";
  var layCC = "md-" + suffix + "-c-clusters";
  var layCP = "md-" + suffix + "-c-points";
  var layDS = "md-" + suffix + "-d-sym";
  var imgArrow = "md-" + suffix + "-arrow";

  var popupOpen = null;
  function closePopup() {
    if (popupOpen) { popupOpen.remove(); popupOpen = null; }
  }

  var rawItems = [];
  var lastResponseTime = 0;
  var tickTimer = 0;
  var animationTickMS = 200;
  var didFit = false;

  function pointerCursor() {
    map.getCanvas().style.cursor = "pointer";
  }
  function defaultCursor() {
    map.getCanvas().style.cursor = "";
  }

  function clearLayerEvents() {
    try {
      if (map.getLayer(layCP)) {
        map.off("click", layCP, onUndirectedPointClick);
        map.off("mouseenter", layCP, pointerCursor);
        map.off("mouseleave", layCP, defaultCursor);
      }
      if (map.getLayer(layDS)) {
        map.off("click", layDS, onDirectedClick);
        map.off("mouseenter", layDS, pointerCursor);
        map.off("mouseleave", layDS, defaultCursor);
      }
      if (map.getLayer(layCC)) {
        map.off("click", layCC, onClusterClick);
        map.off("mouseenter", layCC, pointerCursor);
        map.off("mouseleave", layCC, defaultCursor);
      }
    } catch (e8) {}
  }

  function bearingFromDirection(dx, dy) {
    if (!isFinite(dx) || !isFinite(dy)) { return 0; }
    return Math.atan2(dx, dy) * 180 / Math.PI;
  }

  function addArrowImage() {
    if (!map || typeof map.addImage !== "function") { return; }
    var sz = 64, c = document.createElement("canvas");
    c.width = sz; c.height = sz;
    var x = c.getContext("2d");
    if (!x) { return; }
    if (map.hasImage && map.hasImage(imgArrow) && map.removeImage) {
      try { map.removeImage(imgArrow); } catch (e1) {}
    }
    x.clearRect(0, 0, sz, sz);
    x.save();
    x.translate(sz/2, sz/2);
    x.beginPath();
    x.moveTo(0, -20);
    x.lineTo(14, 12);
    x.lineTo(0, 2);
    x.lineTo(-14, 12);
    x.closePath();
    x.fillStyle = "rgba(59, 130, 246, 0.95)";
    x.fill();
    x.lineWidth = 2;
    x.strokeStyle = "rgba(255, 255, 255, 0.95)";
    x.stroke();
    x.restore();
    var idata = x.getImageData(0, 0, sz, sz);
    try { map.addImage(imgArrow, idata, { pixelRatio: 1 }); } catch (e2) {
      try { map.addImage(imgArrow, idata); } catch (e3) {}
    }
  }

  function rowTRef(row, responseTime) {
    var t = row.time;
    if (typeof t === "number" && isFinite(t)) { return t; }
    return responseTime;
  }

  function rowVelocity(row) {
    var v = row.velocity;
    if (!v || typeof v !== "object") { return { x: 0, y: 0 }; }
    var vx = +v.x, vy = +v.y;
    if (!isFinite(vx)) { vx = 0; }
    if (!isFinite(vy)) { vy = 0; }
    return { x: vx, y: vy };
  }

  function hasDirection(row) {
    var d = row.direction;
    if (!d || typeof d !== "object") { return false; }
    var dx = +d.x, dy = +d.y;
    return isFinite(dx) && isFinite(dy) && (dx !== 0 || dy !== 0);
  }

  function positionOf(row, responseTime, nowSec) {
    var p = row.position;
    if (!p || typeof p !== "object") { return null; }
    var lat = +p.lat, lng = +p.lng;
    if (!isFinite(lat) || !isFinite(lng)) { return null; }
    var tRef = rowTRef(row, responseTime);
    var dt = Math.max(0, nowSec - tRef);
    var vel = rowVelocity(row);
    return { lng: lng + dt * vel.x, lat: lat + dt * vel.y };
  }

  function buildSplit(nowSec) {
    var clusteredFeatures = [];
    var directedFeatures = [];
    var responseTime = lastResponseTime;
    (rawItems || []).forEach(function (row, idx) {
      var pos = positionOf(row, responseTime, nowSec);
      if (!pos) { return; }
      var link = (typeof row.link === "string") ? row.link : "";
      var props = { link: link, idx: idx };
      if (hasDirection(row)) {
        var d = row.direction;
        var bearing = bearingFromDirection(+d.x, +d.y);
        props.bearing = bearing;
        directedFeatures.push({
          type: "Feature",
          id: "d-" + idx,
          geometry: { type: "Point", coordinates: [pos.lng, pos.lat] },
          properties: props
        });
      } else {
        clusteredFeatures.push({
          type: "Feature",
          id: "c-" + idx,
          geometry: { type: "Point", coordinates: [pos.lng, pos.lat] },
          properties: props
        });
      }
    });
    return {
      clustered: { type: "FeatureCollection", features: clusteredFeatures },
      directed: { type: "FeatureCollection", features: directedFeatures }
    };
  }

  function removeDynamicLayers() {
    closePopup();
    clearLayerEvents();
    [layDS, layCP, layCC].forEach(function (id) {
      if (map.getLayer(id)) { map.removeLayer(id); }
    });
    if (map.getSource(srcD)) { map.removeSource(srcD); }
    if (map.getSource(srcC)) { map.removeSource(srcC); }
  }

  function stopTick() {
    if (tickTimer) {
      try { window.clearInterval(tickTimer); } catch (e4) {}
      tickTimer = 0;
    }
  }

  function startTick() {
    stopTick();
    if (!rawItems || !rawItems.length) { return; }
    tickTimer = window.setInterval(tick, animationTickMS);
  }

  function tick() {
    var nowSec = Date.now() / 1000;
    var both = buildSplit(nowSec);
    var srcDc = map.getSource(srcC);
    if (srcDc && srcDc.setData) {
      srcDc.setData(both.clustered);
    }
    var srcDd = map.getSource(srcD);
    if (srcDd && srcDd.setData) {
      srcDd.setData(both.directed);
    }
  }

  function clusterRadiusForDisplay() {
    var dpr = window.devicePixelRatio || 1;
    return Math.round(36 * Math.min(1.85, Math.sqrt(dpr)));
  }

  function fitBoundsBoth(both) {
    var b = new maplibregl.LngLatBounds();
    var any = false;
    both.clustered.features.forEach(function (f) {
      if (f.geometry && f.geometry.coordinates) { b.extend(f.geometry.coordinates); any = true; }
    });
    both.directed.features.forEach(function (f) {
      if (f.geometry && f.geometry.coordinates) { b.extend(f.geometry.coordinates); any = true; }
    });
    if (!any) { return; }
    try {
      map.fitBounds(b, { padding: 48, maxZoom: 12 });
    } catch (e6) {}
  }

  function installFromState() {
    removeDynamicLayers();
    stopTick();
    var nowSec = Date.now() / 1000;
    var both = buildSplit(nowSec);
    if (!both.clustered.features.length && !both.directed.features.length) {
      return;
    }
    if (both.clustered.features.length) {
      map.addSource(srcC, {
        type: "geojson",
        data: both.clustered,
        cluster: true,
        clusterMaxZoom: 14,
        clusterRadius: clusterRadiusForDisplay(),
        clusterMinPoints: 2
      });
      map.addLayer({
        id: layCC,
        type: "circle",
        source: srcC,
        filter: ["has", "point_count"],
        paint: {
          "circle-color": "#818cf8",
          "circle-radius": [
            "step", ["get", "point_count"],
            16, 10, 20, 50, 24, 200, 30
          ],
          "circle-opacity": 0.92,
          "circle-stroke-width": 2,
          "circle-stroke-color": "#e0e7ff"
        }
      });
      map.addLayer({
        id: layCP,
        type: "circle",
        source: srcC,
        filter: ["!", ["has", "point_count"]],
        paint: {
          "circle-color": "#60a5fa",
          "circle-radius": 10,
          "circle-stroke-width": 2,
          "circle-stroke-color": "#ffffff"
        }
      });
    }
    if (both.directed.features.length) {
      addArrowImage();
      map.addSource(srcD, { type: "geojson", data: both.directed, cluster: false });
      if (!map.hasImage || !map.hasImage(imgArrow)) { addArrowImage(); }
      map.addLayer({
        id: layDS,
        type: "symbol",
        source: srcD,
        layout: {
          "icon-image": imgArrow,
          "icon-size": 0.5,
          "icon-allow-overlap": true,
          "icon-ignore-placement": true,
          "icon-rotate": ["get", "bearing"],
          "icon-rotation-alignment": "map"
        }
      });
    }
    if (!didFit) {
      fitBoundsBoth(both);
      didFit = true;
    }
    wireLayerEvents();
    startTick();
  }

  function onMarkerClick(props) {
    var link = (props && props.link) || "";
    if (link) {
      window.location.assign(link);
    }
  }

  function onUndirectedPointClick(e) {
    closePopup();
    var feats = map.queryRenderedFeatures(e.point, { layers: [layCP] });
    if (!feats.length) { return; }
    onMarkerClick(feats[0].properties || {});
  }
  function onDirectedClick(e) {
    closePopup();
    var feats = map.queryRenderedFeatures(e.point, { layers: [layDS] });
    if (!feats.length) { return; }
    onMarkerClick(feats[0].properties || {});
  }
  function onClusterClick(e) {
    closePopup();
    var feats = map.queryRenderedFeatures(e.point, { layers: [layCC] });
    if (!feats.length) { return; }
    var src = map.getSource(srcC);
    if (!src || typeof src.getClusterLeaves !== "function") { return; }
    var clusterFeat = feats[0];
    var cid = +clusterFeat.properties.cluster_id;
    var n = +clusterFeat.properties.point_count || 0;
    var center = clusterFeat.geometry.coordinates.slice();
    var limit = Math.max(n, 1);
    var leavesPromise = src.getClusterLeaves(cid, limit, 0);
    function showLeaves(leaves) {
      if (!leaves || !leaves.length) { return; }
      var wrap = document.createElement("div");
      wrap.className = "flex flex-col gap-1 min-w-[14rem] max-w-sm max-h-72 overflow-y-auto py-1";
      var head = document.createElement("div");
      head.className = "text-sm font-semibold opacity-90 mb-1 sticky top-0 bg-base-100 pb-1 z-10";
      head.textContent = leaves.length + " locations";
      wrap.appendChild(head);
      leaves.forEach(function (leaf) {
        var p = leaf.properties || {};
        var row = document.createElement("div");
        var href = p.link || "";
        if (href) {
          var a = document.createElement("a");
          a.href = href;
          a.className = "link link-primary text-sm block truncate";
          a.textContent = href.length > 64 ? href.slice(0, 61) + "…" : href;
          row.appendChild(a);
        } else {
          row.textContent = "Location";
          row.className = "text-sm opacity-80";
        }
        wrap.appendChild(row);
      });
      popupOpen = new maplibregl.Popup({ offset: 12, closeOnClick: true, maxWidth: "360px" })
        .setLngLat(center)
        .setDOMContent(wrap)
        .addTo(map);
    }
    if (leavesPromise && typeof leavesPromise.then === "function") {
      leavesPromise.then(showLeaves).catch(function () {});
    }
  }
  function wireLayerEvents() {
    if (map.getLayer(layCP)) {
      map.on("click", layCP, onUndirectedPointClick);
      map.on("mouseenter", layCP, pointerCursor);
      map.on("mouseleave", layCP, defaultCursor);
    }
    if (map.getLayer(layDS)) {
      map.on("click", layDS, onDirectedClick);
      map.on("mouseenter", layDS, pointerCursor);
      map.on("mouseleave", layDS, defaultCursor);
    }
    if (map.getLayer(layCC)) {
      map.on("click", layCC, onClusterClick);
      map.on("mouseenter", layCC, pointerCursor);
      map.on("mouseleave", layCC, defaultCursor);
    }
  }

  function applyPayload(arr) {
    rawItems = Array.isArray(arr) ? arr : [];
    lastResponseTime = Date.now() / 1000;
    var nowSec = Date.now() / 1000;
    var both = buildSplit(nowSec);
    var hasC = both.clustered.features.length > 0;
    var hasD = both.directed.features.length > 0;
    var srcHasC = !!map.getSource(srcC);
    var srcHasD = !!map.getSource(srcD);
    var needsRebuild = (hasC !== srcHasC) || (hasD !== srcHasD);

    if (needsRebuild) {
      installFromState();
      return;
    }

    if (srcHasC) {
      var srcClustered = map.getSource(srcC);
      if (srcClustered && srcClustered.setData) {
        srcClustered.setData(both.clustered);
      }
    }
    if (srcHasD) {
      var srcDirected = map.getSource(srcD);
      if (srcDirected && srcDirected.setData) {
        srcDirected.setData(both.directed);
      }
    }
    if (!didFit && (hasC || hasD)) {
      fitBoundsBoth(both);
      didFit = true;
    }
    startTick();
  }

  function resolveWebSocketURL(u) {
    if (!u) { return ""; }
    u = String(u).trim();
    if (/^wss?:\/\//i.test(u)) { return u; }
    var loc = window.location;
    var scheme = loc.protocol === "https:" ? "wss:" : "ws:";
    if (u.charAt(0) === "/") {
      return scheme + "//" + loc.host + u;
    }
    return scheme + "//" + loc.host + "/" + u.replace(/^\/+/, "");
  }

  var ws = null;
  var reconnectTimer = 0;
  var shuttingDown = false;

  function reconnectDelayMs() {
    if (refreshMS < 0) { return -1; }
    if (refreshMS > 0) { return refreshMS; }
    return 2000;
  }

  function clearReconnectTimer() {
    if (reconnectTimer) {
      try { window.clearTimeout(reconnectTimer); } catch (eR0) {}
      reconnectTimer = 0;
    }
  }

  function scheduleReconnect() {
    clearReconnectTimer();
    if (shuttingDown) { return; }
    var d = reconnectDelayMs();
    if (d < 0) { return; }
    reconnectTimer = window.setTimeout(connectWebSocket, d);
  }

  function connectWebSocket() {
    clearReconnectTimer();
    if (shuttingDown || !dataURL) { return; }
    var url = resolveWebSocketURL(dataURL);
    if (!url) {
      try { console.error("MapDisplay: invalid WebSocket URL"); } catch (eR1) {}
      return;
    }
    try {
      if (ws) {
        ws.onopen = ws.onmessage = ws.onerror = ws.onclose = null;
        try { ws.close(); } catch (eR2) {}
        ws = null;
      }
    } catch (eR3) {}
    try {
      ws = new WebSocket(url);
    } catch (eR4) {
      try { console.error("MapDisplay WebSocket construct failed", eR4); } catch (eR5) {}
      scheduleReconnect();
      return;
    }
    ws.onmessage = async function (ev) {
      var data;
      try {
        if (ev.data instanceof ArrayBuffer) {
          if (typeof CBOR === "undefined" || typeof CBOR.decode !== "function") { return; }
          data = CBOR.decode(new Uint8Array(ev.data));
        } else if (typeof Blob !== "undefined" && ev.data instanceof Blob) {
          if (typeof CBOR === "undefined" || typeof CBOR.decode !== "function") { return; }
          var ab = await ev.data.arrayBuffer();
          data = CBOR.decode(new Uint8Array(ab));
        } else if (typeof ev.data === "string") {
          data = JSON.parse(ev.data);
        } else {
          return;
        }
      } catch (eR6) {
        try { console.error("MapDisplay WebSocket decode failed", eR6); } catch (eR7) {}
        return;
      }
      if (Array.isArray(data)) {
        applyPayload(data);
      }
    };
    ws.onopen = function () {
      sendViewportBoundsNow();
    };
    ws.onerror = function () {
      try { console.error("MapDisplay WebSocket error"); } catch (eR9) {}
    };
    ws.onclose = function () {
      ws = null;
      if (!shuttingDown) { scheduleReconnect(); }
    };
  }

  var boundsDebounceTimer = 0;
  var boundsDebounceMs = 150;

  function sendViewportBoundsNow() {
    if (!map || typeof map.getBounds !== "function") { return; }
    if (!ws || ws.readyState !== WebSocket.OPEN) { return; }
    try {
      var b = map.getBounds();
      var sw = b.getSouthWest();
      var ne = b.getNorthEast();
      var msgObj = {
        type: "mapDisplayViewport",
        bounds: { west: sw.lng, south: sw.lat, east: ne.lng, north: ne.lat },
        zoom: map.getZoom()
      };
      if (typeof CBOR !== "undefined" && typeof CBOR.encode === "function") {
        ws.send(CBOR.encode(msgObj));
      } else {
        ws.send(JSON.stringify(msgObj));
      }
    } catch (eB0) {
      try { console.error("MapDisplay send viewport bounds failed", eB0); } catch (eB1) {}
    }
  }

  function scheduleSendViewportBounds() {
    if (boundsDebounceTimer) {
      try { window.clearTimeout(boundsDebounceTimer); } catch (eB2) {}
      boundsDebounceTimer = 0;
    }
    boundsDebounceTimer = window.setTimeout(function () {
      boundsDebounceTimer = 0;
      sendViewportBoundsNow();
    }, boundsDebounceMs);
  }

  map.on("moveend", scheduleSendViewportBounds);
  map.on("zoomend", scheduleSendViewportBounds);
  map.on("rotateend", scheduleSendViewportBounds);
  map.on("pitchend", scheduleSendViewportBounds);
  map.on("resize", scheduleSendViewportBounds);

  map.on("load", function () {
    connectWebSocket();
  });

  window.addEventListener("beforeunload", function () {
    shuttingDown = true;
    if (boundsDebounceTimer) {
      try { window.clearTimeout(boundsDebounceTimer); } catch (eB3) {}
      boundsDebounceTimer = 0;
    }
    clearReconnectTimer();
    try {
      if (ws) {
        ws.onopen = ws.onmessage = ws.onerror = ws.onclose = null;
        ws.close();
      }
    } catch (eR10) {}
    ws = null;
  });

  function syncStyle() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function () {
      didFit = false;
      installFromState();
    });
  }
  if (document.body) {
    new MutationObserver(syncStyle).observe(document.body, { attributes: true, attributeFilter: ["data-theme"] });
  }
  window.addEventListener("storage", function (ev) {
    if (ev.key !== "theme") { return; }
    syncStyle();
  });
})();`

	return Group([]Node{
		Div(ID(mapElID), Class(classes)),
		Script(Raw(initJS)),
	})
}
