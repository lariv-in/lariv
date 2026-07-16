package components

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"regexp"
	"strings"

	"github.com/lariv-in/lariv/getters"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Pinned MapLibre for [MapDisplay]; matches seer map plugins.
const (
	mapDisplayLibreCDNVersion = "4.7.1"
	mapDisplayCBORXCDNVersion = "1.6.0"
)

var mapDisplayIDSanitize = regexp.MustCompile(`[^a-zA-Z0-9-]+`)

// MapDisplayLibreHead loads MapLibre CSS and JS script dependencies from unpkg.
// It should be registered/included once per page in the shell ExtraHead metadata when using [MapDisplay].
type MapDisplayLibreHead struct {
	// Page embeds common component properties like Key and Roles.
	Page
}

// GetKey returns the unique key identifier for this MapDisplayLibreHead.
func (e *MapDisplayLibreHead) GetKey() string { return e.Key }

// GetRoles returns the authorized roles required to view this MapDisplayLibreHead.
func (e *MapDisplayLibreHead) GetRoles() []string { return e.Roles }

// Build compiles the MapDisplayLibreHead component into CDN stylesheet and script tags.
func (e *MapDisplayLibreHead) Build(ctx context.Context) Node {
	baseMapLibre := "https://unpkg.com/maplibre-gl@" + mapDisplayLibreCDNVersion + "/dist/"
	baseCBORX := "https://unpkg.com/cbor-x@" + mapDisplayCBORXCDNVersion + "/dist/"
	return Group([]Node{
		Link(Href(baseMapLibre+"maplibre-gl.css"), Rel("stylesheet"), CrossOrigin("anonymous")),
		Script(Src(baseMapLibre+"maplibre-gl.js"), CrossOrigin("anonymous")),
		Script(Src(baseCBORX+"index.js"), CrossOrigin("anonymous")),
	})
}

// mapDisplayIDSuffix sanitizes pageKey strings to generate safe, valid DOM element ID suffixes.
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

// MapDisplay renders an interactive MapLibre map that opens a WebSocket to stream real-time marker payloads.
// Payloads streamed over the WebSocket are gzip-compressed CBOR arrays representing marker objects.
//
// Use Cases:
//   - Plotting live vehicle telemetry coordinates, tracking physical deliveries, showing location hotspots, or rendering cluster maps.
//
// Example:
//
//	&components.MapDisplay{
//	    DataURL: lariv.RoutePath("vehicles.LiveCoordinates", nil),
//	    Classes: "w-full h-[600px]",
//	}
type MapDisplay struct {
	// Page embeds common component properties like Key and Roles.
	Page
	// DataURL is the WebSocket endpoint URL (ws/wss or same-origin path like "/app/live/map") where marker updates are fetched.
	DataURL getters.Getter[string]
	// RefreshMS represents the milliseconds to wait before reconnecting after the socket closes or errors (defaults to 2000ms).
	RefreshMS getters.Getter[int64]
	// Classes represents additional CSS classes applied to the output HTML wrapper.
	// (Discouraged: Use layout containers or theme styling instead of custom styling overrides).
	Classes string
	// DeferStart specifies if MapDisplay should postpone starting the WebSocket connection until driven manually via the JS API.
	DeferStart getters.Getter[bool]
	// SkipAutoFitBounds specifies if automatic zoom fit-bounds behavior should be skipped when the first marker payload is received.
	SkipAutoFitBounds getters.Getter[bool]
	// MarkerIconSizeDefault represents the default MapLibre scale multiplier applied to custom raster icons (defaults to 0.32).
	MarkerIconSizeDefault getters.Getter[float64]
}

// GetKey returns the unique key identifier for this MapDisplay.
func (e *MapDisplay) GetKey() string { return e.Key }

// GetRoles returns the authorized roles required to view this MapDisplay.
func (e *MapDisplay) GetRoles() []string { return e.Roles }

// Build compiles the MapDisplay component into MapLibre container Divs and dynamic script setup tags.
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

	deferStart := false
	if e.DeferStart != nil {
		v, err := e.DeferStart(ctx)
		if err != nil {
			slog.Error("MapDisplay DeferStart getter failed", "error", err, "key", e.Key)
			return ContainerError{
				Page:  Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		deferStart = v
	}

	skipAutoFitBounds := false
	if e.SkipAutoFitBounds != nil {
		v, err := e.SkipAutoFitBounds(ctx)
		if err != nil {
			slog.Error("MapDisplay SkipAutoFitBounds getter failed", "error", err, "key", e.Key)
			return ContainerError{
				Page:  Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		skipAutoFitBounds = v
	}

	markerIconSizeDefault := 0.32
	if e.MarkerIconSizeDefault != nil {
		v, err := e.MarkerIconSizeDefault(ctx)
		if err != nil {
			slog.Error("MapDisplay MarkerIconSizeDefault getter failed", "error", err, "key", e.Key)
			return ContainerError{
				Page:  Page{Key: e.Key + ".err"},
				Error: getters.Static(err),
			}.Build(ctx)
		}
		markerIconSizeDefault = v
	}
	if markerIconSizeDefault < 0.02 {
		markerIconSizeDefault = 0.02
	}
	if markerIconSizeDefault > 2 {
		markerIconSizeDefault = 2
	}

	suffix := mapDisplayIDSuffix(e.Key)
	mapElID := "mapdisplay-" + suffix + "-map"
	dataURLBytes, _ := json.Marshal(dataURL)
	refreshMSBytes, _ := json.Marshal(refreshMS)
	suffixBytes, _ := json.Marshal(suffix)
	deferStartBytes, _ := json.Marshal(deferStart)
	skipAutoFitBoundsBytes, _ := json.Marshal(skipAutoFitBounds)
	markerIconSizeDefaultBytes, _ := json.Marshal(markerIconSizeDefault)

	classes := strings.TrimSpace(e.Classes)
	if classes == "" {
		classes = "w-full h-[min(80vh,720px)] min-h-80 rounded-box border border-base-300 relative z-[1]"
	}

	mapCtrlCSS := "#" + mapElID + `.maplibregl-map .maplibregl-control-container {
  z-index: 11 !important;
  pointer-events: none !important;
}
#` + mapElID + `.maplibregl-map .maplibregl-ctrl-top-left,
#` + mapElID + `.maplibregl-map .maplibregl-ctrl-top-right {
  z-index: 12 !important;
  pointer-events: auto !important;
}
#` + mapElID + `.maplibregl-map .maplibregl-ctrl,
#` + mapElID + `.maplibregl-map .maplibregl-ctrl-group,
#` + mapElID + `.maplibregl-map .maplibregl-ctrl-group button {
  pointer-events: auto !important;
}
#` + mapElID + `.maplibregl-map .maplibregl-ctrl-group button {
  min-width: 29px !important;
  min-height: 29px !important;
  box-sizing: border-box !important;
}
#` + mapElID + `.maplibregl-map .maplibregl-ctrl span {
  max-width: none !important;
}
#` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px;
  max-height: min(50vh, 320px);
  overflow-y: auto;
  overflow-x: hidden;
  background: rgba(255, 255, 255, 0.96);
  border: 1px solid rgba(15, 23, 42, 0.12);
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.12);
}
#` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar button.mapdisplay-layer-toggle-btn {
  min-width: 0 !important;
  min-height: 0 !important;
  width: 100%;
  height: auto !important;
  padding: 6px 10px !important;
  font-size: 12px !important;
  line-height: 1.25 !important;
  font-weight: 500;
  white-space: normal !important;
  word-break: break-word !important;
  border-radius: 6px !important;
  text-align: center;
  color: #0f172a;
  background: rgba(241, 245, 249, 0.95);
  border: 1px solid rgba(15, 23, 42, 0.08) !important;
}
#` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar button.mapdisplay-layer-toggle-btn.maplibregl-ctrl-active {
  background: rgba(59, 130, 246, 0.18);
  border-color: rgba(59, 130, 246, 0.45) !important;
  color: #0f172a;
}
body[data-theme="dark"] #` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar {
  background: rgba(30, 41, 59, 0.96);
  border-color: rgba(148, 163, 184, 0.22);
}
body[data-theme="dark"] #` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar button.mapdisplay-layer-toggle-btn {
  color: #e2e8f0;
  background: rgba(51, 65, 85, 0.6);
  border-color: rgba(148, 163, 184, 0.2) !important;
}
body[data-theme="dark"] #` + mapElID + `.maplibregl-map .mapdisplay-layer-toolbar button.mapdisplay-layer-toggle-btn.maplibregl-ctrl-active {
  background: rgba(59, 130, 246, 0.35);
  border-color: rgba(96, 165, 250, 0.55) !important;
  color: #f8fafc;
}
`

	initJS := `(function(){
  var suffix = ` + string(suffixBytes) + `;
  var dataURL = ` + string(dataURLBytes) + `;
  var refreshMS = ` + string(refreshMSBytes) + `;
  var deferStart = ` + string(deferStartBytes) + `;
  var skipAutoFitBounds = ` + string(skipAutoFitBoundsBytes) + `;
  var mapMarkerIconSizeDefault = ` + string(markerIconSizeDefaultBytes) + `;
  var mapElId = "mapdisplay-" + suffix + "-map";
  var _currentScript = document.currentScript;
  var _scriptSiblingEl = _currentScript ? _currentScript.previousElementSibling : null;

  function mapDisplayRunInit() {
  var mapEl = _scriptSiblingEl || document.getElementById(mapElId);
  if (!mapEl) { return; }
  if (mapEl.classList.contains("maplibregl-map")) { return; }
  if (typeof maplibregl === "undefined") {
    mapDisplayRunInit._n = (mapDisplayRunInit._n || 0) + 1;
    if (mapDisplayRunInit._n > 120) { return; }
    setTimeout(mapDisplayRunInit, 50);
    return;
  }
  var styleLight = "https://tiles.openfreemap.org/styles/liberty";
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

  var srcC = "md-" + suffix + "-c-src";
  var srcD = "md-" + suffix + "-d-src";
  var layCC = "md-" + suffix + "-c-clusters";
  var layCP = "md-" + suffix + "-c-points";
  var layCPI = "md-" + suffix + "-c-icons";
  var layDS = "md-" + suffix + "-d-sym";
  var imgArrow = "md-" + suffix + "-arrow";

  var currentLayerMode = false;
  var dynamicIds = [];
  var lastLayerSig = "";
  var layerVisibility = {};
  var layerToggleControlInstance = null;

  var popupOpen = null;
  function closePopup() {
    if (popupOpen) { popupOpen.remove(); popupOpen = null; }
  }

  var rawItems = [];
  var lastResponseTime = 0;
  var tickTimer = 0;
  var animationTickMS = 200;
  var didFit = !!skipAutoFitBounds;
  var mapIconRasterGen = 0;
  function bumpMapIconRasterGen() {
    mapIconRasterGen++;
  }

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
      if (map.getLayer(layCPI)) {
        map.off("click", layCPI, onUndirectedPointClick);
        map.off("mouseenter", layCPI, pointerCursor);
        map.off("mouseleave", layCPI, defaultCursor);
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
    dynamicIds.forEach(function (x) {
      try {
        if (x.layCP && map.getLayer(x.layCP) && x.hUP) {
          map.off("click", x.layCP, x.hUP);
          map.off("mouseenter", x.layCP, pointerCursor);
          map.off("mouseleave", x.layCP, defaultCursor);
        }
        if (x.layCPI && map.getLayer(x.layCPI) && x.hUP) {
          map.off("click", x.layCPI, x.hUP);
          map.off("mouseenter", x.layCPI, pointerCursor);
          map.off("mouseleave", x.layCPI, defaultCursor);
        }
        if (x.layDS && map.getLayer(x.layDS) && x.hDS) {
          map.off("click", x.layDS, x.hDS);
          map.off("mouseenter", x.layDS, pointerCursor);
          map.off("mouseleave", x.layDS, defaultCursor);
        }
        if (x.layCC && map.getLayer(x.layCC) && x.hCC) {
          map.off("click", x.layCC, x.hCC);
          map.off("mouseenter", x.layCC, pointerCursor);
          map.off("mouseleave", x.layCC, defaultCursor);
        }
      } catch (e9) {}
    });
  }

  function bearingFromDirection(dx, dy) {
    if (!isFinite(dx) || !isFinite(dy)) { return 0; }
    return Math.atan2(dx, dy) * 180 / Math.PI;
  }

  function ensureColoredArrowImage(arrowImageId, fillCss) {
    if (!map || typeof map.addImage !== "function" || !arrowImageId) { return; }
    if (map.hasImage && map.hasImage(arrowImageId)) { return; }
    var sz = 64, c = document.createElement("canvas");
    c.width = sz; c.height = sz;
    var x = c.getContext("2d");
    if (!x) { return; }
    var fill = (typeof fillCss === "string" && fillCss.trim()) ? fillCss.trim() : "rgba(59, 130, 246, 0.95)";
    x.clearRect(0, 0, sz, sz);
    x.save();
    x.translate(sz/2, sz/2);
    x.beginPath();
    x.moveTo(0, -20);
    x.lineTo(14, 12);
    x.lineTo(0, 2);
    x.lineTo(-14, 12);
    x.closePath();
    x.fillStyle = fill;
    x.fill();
    x.lineWidth = 2;
    x.strokeStyle = "rgba(255, 255, 255, 0.95)";
    x.stroke();
    x.restore();
    var idata = x.getImageData(0, 0, sz, sz);
    try { map.addImage(arrowImageId, idata, { pixelRatio: 1 }); } catch (e2) {
      try { map.addImage(arrowImageId, idata); } catch (e3) {}
    }
  }

  function addArrowImage() {
    ensureColoredArrowImage(imgArrow, "rgba(59, 130, 246, 0.95)");
  }

  function cborMapToObject(v) {
    if (!v || typeof v !== "object") { return v; }
    if (typeof Map !== "undefined" && v instanceof Map) {
      var o = {};
      v.forEach(function (val, key) {
        var ks = (typeof key === "string") ? key : String(key);
        o[ks] = val;
      });
      return o;
    }
    return v;
  }

  function normalizeDecodedRows(arr) {
    if (!Array.isArray(arr)) { return []; }
    return arr.map(function (row) {
      row = cborMapToObject(row);
      if (!row || typeof row !== "object") { return row; }
      row.position = cborMapToObject(row.position);
      row.direction = cborMapToObject(row.direction);
      row.velocity = cborMapToObject(row.velocity);
      return row;
    });
  }

  function rowTRef(row, responseTime) {
    var t = row.time;
    if (typeof t !== "number" || !isFinite(t)) {
      t = row.Time;
    }
    if (typeof t === "number" && isFinite(t)) { return t; }
    return responseTime;
  }

  function rowVelocity(row) {
    var v = row.velocity || row.Velocity;
    if (!v || typeof v !== "object") { return { x: 0, y: 0 }; }
    v = cborMapToObject(v);
    var vx = v.x !== undefined ? +v.x : +v.X;
    var vy = v.y !== undefined ? +v.y : +v.Y;
    if (!isFinite(vx)) { vx = 0; }
    if (!isFinite(vy)) { vy = 0; }
    return { x: vx, y: vy };
  }

  function hasDirection(row) {
    var d = row.direction || row.Direction;
    if (!d || typeof d !== "object") { return false; }
    d = cborMapToObject(d);
    var dx = d.x !== undefined ? +d.x : +d.X;
    var dy = d.y !== undefined ? +d.y : +d.Y;
    return isFinite(dx) && isFinite(dy) && (dx !== 0 || dy !== 0);
  }

  function rowColorString(row) {
    if (!row || typeof row !== "object") { return ""; }
    var c = row.color;
    if (typeof c === "string" && c.trim()) { return c.trim(); }
    c = row.Color;
    if (typeof c === "string" && c.trim()) { return c.trim(); }
    return "";
  }

  function rowIconString(row) {
    if (!row || typeof row !== "object") { return ""; }
    var ic = row.icon;
    if (typeof ic === "string" && ic.trim()) { return ic.trim(); }
    ic = row.Icon;
    if (typeof ic === "string" && ic.trim()) { return ic.trim(); }
    return "";
  }

  function isIconLoadableURL(s) {
    if (!s || typeof s !== "string") { return false; }
    if (/^https?:\/\//i.test(s)) { return true; }
    if (/^\/\//.test(s)) { return true; }
    if (/^\//.test(s)) { return true; }
    if (/^data:image\//i.test(s)) { return true; }
    return false;
  }

  function resolveIconURL(s) {
    if (!s) { return ""; }
    if (/^\/\//.test(s)) {
      try {
        var p = window.location && window.location.protocol;
        return (p || "https:") + s;
      } catch (eRes) { return "https:" + s; }
    }
    if (/^\//.test(s) && window.location && window.location.origin) {
      return window.location.origin + s;
    }
    return s;
  }

  function hashStr(s) {
    var h = 0;
    if (!s) { return "0"; }
    for (var i = 0; i < s.length; i++) {
      h = ((h << 5) - h + s.charCodeAt(i)) | 0;
    }
    return (h >>> 0).toString(16) + "-" + String(s.length);
  }

  function markerIconImageId(resolvedURL) {
    return "md-" + suffix + "-mi-" + hashStr(resolvedURL);
  }

  function clampMarkerIconSize(n) {
    if (!isFinite(n)) { return mapMarkerIconSizeDefault; }
    if (n < 0.02) { return 0.02; }
    if (n > 2) { return 2; }
    return n;
  }

  function rowIconSize(row) {
    if (!row || typeof row !== "object") { return NaN; }
    var keys = ["iconSize", "IconSize", "icon_size", "Icon_Size"];
    for (var ki = 0; ki < keys.length; ki++) {
      var v = row[keys[ki]];
      if (typeof v === "number" && isFinite(v)) { return clampMarkerIconSize(v); }
      if (typeof v === "string" && v.trim()) {
        var n = parseFloat(v.trim());
        if (isFinite(n)) { return clampMarkerIconSize(n); }
      }
    }
    return NaN;
  }

  function applyMarkerStyleToProps(row, props) {
    var cStr = rowColorString(row);
    if (cStr) { props.mdColor = cStr; }
    var icRaw = rowIconString(row);
    if (icRaw && isIconLoadableURL(icRaw)) {
      var abs = resolveIconURL(icRaw);
      if (abs) { props.mdIconImgId = markerIconImageId(abs); }
    }
    var isz = rowIconSize(row);
    if (isFinite(isz)) { props.mdIconSize = isz; }
    if (hasDirection(row) && !props.mdIconImgId && cStr) {
      props.arrowImg = imgArrow + "-c-" + hashStr(cStr);
    }
  }

  function collectMarkerIconURLs(items) {
    var out = [];
    var seen = {};
    (items || []).forEach(function (row) {
      var u = rowIconString(row);
      if (!isIconLoadableURL(u)) { return; }
      var abs = resolveIconURL(u);
      if (!abs || seen[abs]) { return; }
      seen[abs] = true;
      out.push(abs);
    });
    return out;
  }

  function cloneImageData(src) {
    if (!src || !src.data) { return null; }
    try {
      var copy = new Uint8ClampedArray(src.data);
      return new ImageData(copy, src.width, src.height);
    } catch (eCl) {
      return null;
    }
  }

  function rasterizeImageForMapIcon(image) {
    if (!image) { return null; }
    if (image.data && typeof image.width === "number" && typeof image.height === "number") {
      return cloneImageData(image) || image;
    }
    var ow = image.width;
    var oh = image.height;
    if (!ow || !oh || !isFinite(ow) || !isFinite(oh)) { return null; }
    var maxPx = 256;
    var w = ow;
    var h = oh;
    if (w > maxPx || h > maxPx) {
      var scale = Math.min(maxPx / w, maxPx / h, 1);
      w = Math.max(1, Math.round(w * scale));
      h = Math.max(1, Math.round(h * scale));
    }
    var c = document.createElement("canvas");
    c.width = w;
    c.height = h;
    var x = c.getContext("2d");
    if (!x) { return null; }
    try {
      x.drawImage(image, 0, 0, w, h);
    } catch (eDraw) {
      return null;
    }
    try {
      var raw = x.getImageData(0, 0, w, h);
      return cloneImageData(raw) || raw;
    } catch (eGid) {
      return null;
    }
  }

  function iconURLIsCrossOrigin(url) {
    try {
      var u = new URL(url, window.location.href);
      return u.origin !== window.location.origin;
    } catch (e0) {
      return false;
    }
  }

  function loadIconRasterImage(url, cb) {
    var img = new Image();
    if (iconURLIsCrossOrigin(url)) {
      img.crossOrigin = "anonymous";
    }
    var done = false;
    function finish(err, im) {
      if (done) { return; }
      done = true;
      try {
        cb(err, im);
      } catch (eCb) {}
    }
    img.onload = function () {
      finish(null, img);
    };
    img.onerror = function () {
      finish(new Error("icon load failed"), null);
    };
    try {
      img.src = url;
    } catch (eSrc) {
      finish(eSrc, null);
    }
  }

  function preloadMarkerIcons(urls, done) {
    if (!map || !urls || !urls.length) {
      if (typeof done === "function") { done(); }
      return;
    }
    var snap = mapIconRasterGen;
    var left = urls.length;
    function step() {
      left--;
      if (left <= 0 && typeof done === "function") { done(); }
    }
    urls.forEach(function (url) {
      var id = markerIconImageId(url);
      if (map.hasImage && map.hasImage(id)) { step(); return; }
      loadIconRasterImage(url, function (err, image) {
        if (!map || snap !== mapIconRasterGen) { step(); return; }
        if (map.hasImage && map.hasImage(id)) { step(); return; }
        if (err || !image) { step(); return; }
        var idata = null;
        try {
          idata = rasterizeImageForMapIcon(image);
        } catch (eRas) {
          idata = null;
        }
        if (!idata) { step(); return; }
        try {
          if (map.hasImage && map.hasImage(id) && map.removeImage) {
            try { map.removeImage(id); } catch (eRmI) {}
          }
        } catch (eRmI2) {}
        try {
          map.addImage(id, idata, { pixelRatio: 1 });
        } catch (eAdI) {
          try { map.addImage(id, idata); } catch (eAdI2) {}
        }
        step();
      });
    });
  }

  function ensureDirectedArrowsFromItems() {
    if (!rawItems || !rawItems.length) { return; }
    rawItems.forEach(function (row) {
      if (!hasDirection(row)) { return; }
      if (isIconLoadableURL(rowIconString(row))) { return; }
      var cStr = rowColorString(row);
      if (!cStr) { return; }
      var aid = imgArrow + "-c-" + hashStr(cStr);
      ensureColoredArrowImage(aid, cStr);
    });
  }

  function clusteredHasIconFeatures(fc) {
    if (!fc || !fc.features) { return false; }
    for (var i = 0; i < fc.features.length; i++) {
      var p = fc.features[i].properties || {};
      if (p.mdIconImgId) { return true; }
    }
    return false;
  }

  function positionOf(row, responseTime, nowSec) {
    var p = row.position || row.Position;
    if (!p || typeof p !== "object") { return null; }
    p = cborMapToObject(p);
    var lat = p.lat !== undefined ? +p.lat : +p.Lat;
    var lng = p.lng !== undefined ? +p.lng : +p.Lng;
    if (!isFinite(lat) || !isFinite(lng)) { return null; }
    var tRef = rowTRef(row, responseTime);
    var dt = Math.max(0, nowSec - tRef);
    var vel = rowVelocity(row);
    return { lng: lng + dt * vel.x, lat: lat + dt * vel.y };
  }

  function sanitizeLayerId(raw) {
    var s = String(raw || "").trim();
    if (!s) { return "_"; }
    s = s.replace(/[^a-zA-Z0-9-]+/g, "-").replace(/^-+|-+$/g, "");
    if (!s) { return "_"; }
    if (s.length > 48) { s = s.substring(0, 48); }
    return s;
  }

  function rowLayerString(row) {
    if (!row || typeof row !== "object") { return ""; }
    var L = row.layer;
    if (typeof L === "string") {
      L = L.trim();
      if (L !== "") { return L; }
    }
    L = row.Layer;
    if (typeof L === "string") {
      L = L.trim();
      if (L !== "") { return L; }
    }
    return "";
  }

  function rowTitleString(row) {
    if (!row || typeof row !== "object") { return ""; }
    var t = row.title;
    if (typeof t === "string") {
      t = t.trim();
      if (t !== "") { return t; }
    }
    t = row.Title;
    if (typeof t === "string") {
      t = t.trim();
      if (t !== "") { return t; }
    }
    return "";
  }

  function anchorLabelFromProps(p, href) {
    var t = (p && typeof p.title === "string") ? p.title.trim() : "";
    if (t === "" && p && typeof p.Title === "string") { t = p.Title.trim(); }
    if (t !== "") { return t.length > 64 ? t.slice(0, 61) + "…" : t; }
    if (!href) { return ""; }
    return href.length > 64 ? href.slice(0, 61) + "…" : href;
  }

  function itemsUseLayers(items) {
    if (!items || !items.length) { return false; }
    for (var i = 0; i < items.length; i++) {
      if (rowLayerString(items[i]) !== "") { return true; }
    }
    return false;
  }

  function layerKeyForRow(row) {
    var L = rowLayerString(row);
    if (L.trim() !== "") { return sanitizeLayerId(L); }
    return "_";
  }

  function buildSplit(nowSec) {
    var clusteredFeatures = [];
    var directedFeatures = [];
    var responseTime = lastResponseTime;
    (rawItems || []).forEach(function (row, idx) {
      var pos = positionOf(row, responseTime, nowSec);
      if (!pos) { return; }
      var link = (typeof row.link === "string") ? row.link : ((typeof row.Link === "string") ? row.Link : "");
      var title = rowTitleString(row);
      var props = { link: link, title: title, idx: idx };
      applyMarkerStyleToProps(row, props);
      if (hasDirection(row)) {
        var d = row.direction || row.Direction;
        if (d && typeof d === "object") { d = cborMapToObject(d); }
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

  function buildLayerBuckets(nowSec) {
    var buckets = {};
    var responseTime = lastResponseTime;
    (rawItems || []).forEach(function (row, idx) {
      var pos = positionOf(row, responseTime, nowSec);
      if (!pos) { return; }
      var lid = layerKeyForRow(row);
      if (!buckets[lid]) {
        buckets[lid] = {
          clustered: { type: "FeatureCollection", features: [] },
          directed: { type: "FeatureCollection", features: [] }
        };
      }
      var link = (typeof row.link === "string") ? row.link : ((typeof row.Link === "string") ? row.Link : "");
      var title = rowTitleString(row);
      var props = { link: link, title: title, idx: idx };
      applyMarkerStyleToProps(row, props);
      if (hasDirection(row)) {
        var d = row.direction || row.Direction;
        if (d && typeof d === "object") { d = cborMapToObject(d); }
        props.bearing = bearingFromDirection(+d.x, +d.y);
        buckets[lid].directed.features.push({
          type: "Feature",
          id: "d-" + lid + "-" + idx,
          geometry: { type: "Point", coordinates: [pos.lng, pos.lat] },
          properties: props
        });
      } else {
        buckets[lid].clustered.features.push({
          type: "Feature",
          id: "c-" + lid + "-" + idx,
          geometry: { type: "Point", coordinates: [pos.lng, pos.lat] },
          properties: props
        });
      }
    });
    return buckets;
  }

  function removeDynamicLayers() {
    closePopup();
    clearLayerEvents();
    dynamicIds.forEach(function (x) {
      [x.layDS, x.layCPI, x.layCP, x.layCC].forEach(function (id) {
        if (id && map.getLayer(id)) { map.removeLayer(id); }
      });
      if (x.srcD && map.getSource(x.srcD)) { map.removeSource(x.srcD); }
      if (x.srcC && map.getSource(x.srcC)) { map.removeSource(x.srcC); }
    });
    dynamicIds = [];
    [layDS, layCPI, layCP, layCC].forEach(function (id) {
      if (map.getLayer(id)) { map.removeLayer(id); }
    });
    if (map.getSource(srcD)) { map.removeSource(srcD); }
    if (map.getSource(srcC)) { map.removeSource(srcC); }
    removeLayerToggleControl();
  }

  function removeLayerToggleControl() {
    if (layerToggleControlInstance && map) {
      try { map.removeControl(layerToggleControlInstance); } catch (eRm) {}
      layerToggleControlInstance = null;
    }
  }

  function createLayerToggleMapControl() {
    var self = {
      _map: null,
      _container: null,
      onAdd: function (m) {
        this._map = m;
        this._container = document.createElement("div");
        this._container.className = "maplibregl-ctrl maplibregl-ctrl-group mapdisplay-layer-toolbar";
        this._container.setAttribute("aria-label", "Map layers");
        return this._container;
      },
      onRemove: function () {
        if (this._container && this._container.parentNode) {
          this._container.parentNode.removeChild(this._container);
        }
        this._map = null;
        this._container = null;
      },
      getDefaultPosition: function () { return "top-left"; }
    };
    return self;
  }

  function syncLayerToggleButtons(bucketKeys) {
    if (!layerToggleControlInstance || !layerToggleControlInstance._container) { return; }
    var wrap = layerToggleControlInstance._container;
    wrap.innerHTML = "";
    bucketKeys.forEach(function (lid) {
      if (layerVisibility[lid] === undefined) { layerVisibility[lid] = true; }
      var btn = document.createElement("button");
      btn.type = "button";
      btn.className = "mapdisplay-layer-toggle-btn";
      btn.setAttribute("aria-pressed", layerVisibility[lid] !== false ? "true" : "false");
      btn.title = lid === "_" ? "Other" : lid;
      var label = lid === "_" ? "Other" : (lid.length > 24 ? lid.slice(0, 23) + "…" : lid);
      btn.textContent = label;
      if (layerVisibility[lid] !== false) {
        btn.classList.add("maplibregl-ctrl-active");
      }
      btn.addEventListener("click", function (ev) {
        if (ev && ev.stopPropagation) { ev.stopPropagation(); }
        if (ev && ev.preventDefault) { ev.preventDefault(); }
        var wasOn = layerVisibility[lid] !== false;
        var on = !wasOn;
        layerVisibility[lid] = on;
        btn.setAttribute("aria-pressed", on ? "true" : "false");
        btn.classList.toggle("maplibregl-ctrl-active", on);
        setLayerGeomVisibility(lid, on);
      });
      wrap.appendChild(btn);
    });
  }

  function syncLayerToolbar(bucketKeys) {
    if (!currentLayerMode || !bucketKeys.length) {
      removeLayerToggleControl();
      return;
    }
    function mountOrRefreshLayerControl() {
      if (!layerToggleControlInstance) {
        layerToggleControlInstance = createLayerToggleMapControl();
        try {
          map.addControl(layerToggleControlInstance, "top-left");
        } catch (eAdd) {
          layerToggleControlInstance = null;
          return;
        }
      }
      syncLayerToggleButtons(bucketKeys);
    }
    mountOrRefreshLayerControl();
    if (!layerToggleControlInstance || !layerToggleControlInstance._container) {
      window.setTimeout(function () {
        if (!currentLayerMode || !bucketKeys.length) { return; }
        mountOrRefreshLayerControl();
      }, 0);
    }
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
    if (currentLayerMode) {
      var buckets = buildLayerBuckets(nowSec);
      dynamicIds.forEach(function (x) {
        var both = buckets[x.lid] || {
          clustered: { type: "FeatureCollection", features: [] },
          directed: { type: "FeatureCollection", features: [] }
        };
        var srcDc = map.getSource(x.srcC);
        if (srcDc && srcDc.setData) { srcDc.setData(both.clustered); }
        var srcDd = map.getSource(x.srcD);
        if (srcDd && srcDd.setData) { srcDd.setData(both.directed); }
      });
    } else {
      var both = buildSplit(nowSec);
      var srcDc = map.getSource(srcC);
      if (srcDc && srcDc.setData) { srcDc.setData(both.clustered); }
      var srcDd = map.getSource(srcD);
      if (srcDd && srcDd.setData) { srcDd.setData(both.directed); }
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

  function fitBoundsLayered(buckets) {
    var b = new maplibregl.LngLatBounds();
    var any = false;
    Object.keys(buckets).forEach(function (lid) {
      if (layerVisibility[lid] === false) { return; }
      var both = buckets[lid];
      both.clustered.features.forEach(function (f) {
        if (f.geometry && f.geometry.coordinates) { b.extend(f.geometry.coordinates); any = true; }
      });
      both.directed.features.forEach(function (f) {
        if (f.geometry && f.geometry.coordinates) { b.extend(f.geometry.coordinates); any = true; }
      });
    });
    if (!any) { return; }
    try {
      map.fitBounds(b, { padding: 48, maxZoom: 12 });
    } catch (eLb) {}
  }

  function layerSignatureFromBuckets(buckets) {
    var keys = Object.keys(buckets).sort();
    return keys.map(function (k) {
      var both = buckets[k];
      var hc = both.clustered.features.length > 0 ? 1 : 0;
      var hd = both.directed.features.length > 0 ? 1 : 0;
      var hi = (hc && clusteredHasIconFeatures(both.clustered)) ? 1 : 0;
      return k + ":" + hc + hd + hi;
    }).join("|");
  }

  function setLayerGeomVisibility(lid, vis) {
    var v = vis ? "visible" : "none";
    dynamicIds.forEach(function (x) {
      if (x.lid !== lid) { return; }
      [x.layCC, x.layCPI, x.layCP, x.layDS].forEach(function (id) {
        if (id && map.getLayer(id)) {
          try { map.setLayoutProperty(id, "visibility", v); } catch (eV) {}
        }
      });
    });
  }

  function makeClusterClick(srcCId, layCCId) {
    return function (e) {
      closePopup();
      var feats = map.queryRenderedFeatures(e.point, { layers: [layCCId] });
      if (!feats.length) { return; }
      var src = map.getSource(srcCId);
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
            a.textContent = anchorLabelFromProps(p, href);
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
    };
  }

  function installFromState() {
    removeDynamicLayers();
    stopTick();
    if (!rawItems || !rawItems.length) {
      lastLayerSig = "";
      currentLayerMode = false;
      return;
    }
    var urls = collectMarkerIconURLs(rawItems);
    preloadMarkerIcons(urls, function () {
      ensureDirectedArrowsFromItems();
      installFromStateImpl();
    });
  }

  function installFromStateImpl() {
    if (!rawItems || !rawItems.length) {
      lastLayerSig = "";
      return;
    }
    var nowSec = Date.now() / 1000;
    var useLayers = itemsUseLayers(rawItems);
    currentLayerMode = useLayers;
    if (useLayers) {
      var buckets = buildLayerBuckets(nowSec);
      var bucketKeys = Object.keys(buckets).filter(function (k) {
        var both = buckets[k];
        return both.clustered.features.length > 0 || both.directed.features.length > 0;
      }).sort();
      if (!bucketKeys.length) {
        lastLayerSig = "";
        return;
      }
      lastLayerSig = layerSignatureFromBuckets(buckets);
      var anyDirected = false;
      bucketKeys.forEach(function (lid) {
        var both = buckets[lid];
        var srcCt = "md-" + suffix + "-L-" + lid + "-c-src";
        var srcDt = "md-" + suffix + "-L-" + lid + "-d-src";
        var layCCt = "md-" + suffix + "-L-" + lid + "-c-clusters";
        var layCPt = "md-" + suffix + "-L-" + lid + "-c-points";
        var layCPIt = "md-" + suffix + "-L-" + lid + "-c-icons";
        var layDSt = "md-" + suffix + "-L-" + lid + "-d-sym";
        var entry = { lid: lid, srcC: srcCt, srcD: srcDt, layCC: layCCt, layCP: layCPt, layCPI: null, layDS: layDSt };
        if (both.clustered.features.length) {
          map.addSource(srcCt, {
            type: "geojson",
            data: both.clustered,
            cluster: true,
            clusterMaxZoom: 14,
            clusterRadius: clusterRadiusForDisplay(),
            clusterMinPoints: 2
          });
          map.addLayer({
            id: layCCt,
            type: "circle",
            source: srcCt,
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
            id: layCPt,
            type: "circle",
            source: srcCt,
            filter: ["all", ["!", ["has", "point_count"]], ["!", ["has", "mdIconImgId"]]],
            paint: {
              "circle-color": ["coalesce", ["get", "mdColor"], "#60a5fa"],
              "circle-radius": 10,
              "circle-stroke-width": 2,
              "circle-stroke-color": "#ffffff"
            }
          });
          if (clusteredHasIconFeatures(both.clustered)) {
            map.addLayer({
              id: layCPIt,
              type: "symbol",
              source: srcCt,
              filter: ["all", ["!", ["has", "point_count"]], ["has", "mdIconImgId"]],
              layout: {
                "icon-image": ["get", "mdIconImgId"],
                "icon-size": ["coalesce", ["get", "mdIconSize"], mapMarkerIconSizeDefault],
                "icon-allow-overlap": true,
                "icon-ignore-placement": true
              }
            });
            entry.layCPI = layCPIt;
          }
        }
        if (both.directed.features.length) {
          anyDirected = true;
          addArrowImage();
          map.addSource(srcDt, { type: "geojson", data: both.directed, cluster: false });
          if (!map.hasImage || !map.hasImage(imgArrow)) { addArrowImage(); }
          map.addLayer({
            id: layDSt,
            type: "symbol",
            source: srcDt,
            layout: {
              "icon-image": [
                "case",
                ["all", ["has", "mdIconImgId"], ["!=", ["get", "mdIconImgId"], ""]],
                ["get", "mdIconImgId"],
                ["all", ["has", "arrowImg"], ["!=", ["get", "arrowImg"], ""]],
                ["get", "arrowImg"],
                imgArrow
              ],
              "icon-size": ["coalesce", ["get", "mdIconSize"], mapMarkerIconSizeDefault],
              "icon-allow-overlap": true,
              "icon-ignore-placement": true,
              "icon-rotate": ["get", "bearing"],
              "icon-rotation-alignment": "map"
            }
          });
        }
        if (both.clustered.features.length) {
          entry.hUP = function (e) {
            closePopup();
            var ls = entry.layCPI ? [layCPt, entry.layCPI] : [layCPt];
            var feats = map.queryRenderedFeatures(e.point, { layers: ls });
            if (!feats.length) { return; }
            onMarkerClick(feats[0].properties || {});
          };
          entry.hCC = makeClusterClick(srcCt, layCCt);
          map.on("click", layCPt, entry.hUP);
          map.on("mouseenter", layCPt, pointerCursor);
          map.on("mouseleave", layCPt, defaultCursor);
          if (entry.layCPI) {
            map.on("click", entry.layCPI, entry.hUP);
            map.on("mouseenter", entry.layCPI, pointerCursor);
            map.on("mouseleave", entry.layCPI, defaultCursor);
          }
          map.on("click", layCCt, entry.hCC);
          map.on("mouseenter", layCCt, pointerCursor);
          map.on("mouseleave", layCCt, defaultCursor);
        }
        if (both.directed.features.length) {
          entry.hDS = function (e) {
            closePopup();
            var feats = map.queryRenderedFeatures(e.point, { layers: [layDSt] });
            if (!feats.length) { return; }
            onMarkerClick(feats[0].properties || {});
          };
          map.on("click", layDSt, entry.hDS);
          map.on("mouseenter", layDSt, pointerCursor);
          map.on("mouseleave", layDSt, defaultCursor);
        }
        dynamicIds.push(entry);
      });
      if (anyDirected && (!map.hasImage || !map.hasImage(imgArrow))) { addArrowImage(); }
      dynamicIds.forEach(function (x) {
        setLayerGeomVisibility(x.lid, layerVisibility[x.lid] !== false);
      });
      syncLayerToolbar(bucketKeys);
      if (!didFit) {
        fitBoundsLayered(buckets);
        didFit = true;
      }
      startTick();
      return;
    }
    lastLayerSig = "";
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
        filter: ["all", ["!", ["has", "point_count"]], ["!", ["has", "mdIconImgId"]]],
        paint: {
          "circle-color": ["coalesce", ["get", "mdColor"], "#60a5fa"],
          "circle-radius": 10,
          "circle-stroke-width": 2,
          "circle-stroke-color": "#ffffff"
        }
      });
      if (clusteredHasIconFeatures(both.clustered)) {
        map.addLayer({
          id: layCPI,
          type: "symbol",
          source: srcC,
          filter: ["all", ["!", ["has", "point_count"]], ["has", "mdIconImgId"]],
          layout: {
            "icon-image": ["get", "mdIconImgId"],
            "icon-size": ["coalesce", ["get", "mdIconSize"], mapMarkerIconSizeDefault],
            "icon-allow-overlap": true,
            "icon-ignore-placement": true
          }
        });
      }
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
          "icon-image": [
            "case",
            ["all", ["has", "mdIconImgId"], ["!=", ["get", "mdIconImgId"], ""]],
            ["get", "mdIconImgId"],
            ["all", ["has", "arrowImg"], ["!=", ["get", "arrowImg"], ""]],
            ["get", "arrowImg"],
            imgArrow
          ],
          "icon-size": ["coalesce", ["get", "mdIconSize"], mapMarkerIconSizeDefault],
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
    var ls = map.getLayer(layCPI) ? [layCP, layCPI] : [layCP];
    var feats = map.queryRenderedFeatures(e.point, { layers: ls });
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
          a.textContent = anchorLabelFromProps(p, href);
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
    if (map.getLayer(layCPI)) {
      map.on("click", layCPI, onUndirectedPointClick);
      map.on("mouseenter", layCPI, pointerCursor);
      map.on("mouseleave", layCPI, defaultCursor);
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
    rawItems = normalizeDecodedRows(Array.isArray(arr) ? arr : []);
    lastResponseTime = Date.now() / 1000;
    var nextLM = itemsUseLayers(rawItems);
    if (nextLM !== currentLayerMode) {
      installFromState();
      return;
    }
    currentLayerMode = nextLM;
    preloadMarkerIcons(collectMarkerIconURLs(rawItems), function () {
      applyPayloadAfterIcons();
    });
  }

  function applyPayloadAfterIcons() {
    ensureDirectedArrowsFromItems();
    var nowSec = Date.now() / 1000;
    if (!currentLayerMode) {
      var both = buildSplit(nowSec);
      var hasC = both.clustered.features.length > 0;
      var hasD = both.directed.features.length > 0;
      var srcHasC = !!map.getSource(srcC);
      var srcHasD = !!map.getSource(srcD);
      var wantIconLayer = hasC && clusteredHasIconFeatures(both.clustered);
      var mapHasIconLayer = !!map.getLayer(layCPI);
      var needsRebuild = (hasC !== srcHasC) || (hasD !== srcHasD) || (wantIconLayer !== mapHasIconLayer);
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
      return;
    }
    var buckets = buildLayerBuckets(nowSec);
    var sig = layerSignatureFromBuckets(buckets);
    if (sig !== lastLayerSig) {
      installFromState();
      return;
    }
    dynamicIds.forEach(function (x) {
      var both = buckets[x.lid] || {
        clustered: { type: "FeatureCollection", features: [] },
        directed: { type: "FeatureCollection", features: [] }
      };
      var srcClustered = map.getSource(x.srcC);
      if (srcClustered && srcClustered.setData) {
        srcClustered.setData(both.clustered);
      }
      var srcDirected = map.getSource(x.srcD);
      if (srcDirected && srcDirected.setData) {
        srcDirected.setData(both.directed);
      }
    });
    var bk = Object.keys(buckets).filter(function (k) {
      var b = buckets[k];
      return b.clustered.features.length > 0 || b.directed.features.length > 0;
    }).sort();
    syncLayerToolbar(bk);
    if (!didFit && bk.length) {
      fitBoundsLayered(buckets);
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

  function mapDisplayGzipSupported() {
    return typeof CompressionStream !== "undefined" && typeof DecompressionStream !== "undefined";
  }

  async function mapDisplayGzipEncode(bytes) {
    var u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
    var cs = new CompressionStream("gzip");
    var stream = new Blob([u8]).stream().pipeThrough(cs);
    var ab = await new Response(stream).arrayBuffer();
    return new Uint8Array(ab);
  }

  async function mapDisplayGzipDecode(bytes) {
    var u8 = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
    var ds = new DecompressionStream("gzip");
    var stream = new Blob([u8]).stream().pipeThrough(ds);
    var ab = await new Response(stream).arrayBuffer();
    return new Uint8Array(ab);
  }

  async function mapDisplayDecodeInbound(raw) {
    if (!mapDisplayGzipSupported()) {
      try { console.error("MapDisplay: CompressionStream unsupported; gzip wire format required"); } catch (eGz0) {}
      return null;
    }
    if (typeof CBOR === "undefined" || typeof CBOR.decode !== "function") { return null; }
    var body = await mapDisplayGzipDecode(raw);
    return CBOR.decode(body);
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
          data = await mapDisplayDecodeInbound(new Uint8Array(ev.data));
        } else if (typeof Blob !== "undefined" && ev.data instanceof Blob) {
          data = await mapDisplayDecodeInbound(new Uint8Array(await ev.data.arrayBuffer()));
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

  async function sendViewportBoundsNow() {
    if (!map || typeof map.getBounds !== "function") { return; }
    if (!ws || ws.readyState !== WebSocket.OPEN) { return; }
    if (!mapDisplayGzipSupported()) {
      try { console.error("MapDisplay: CompressionStream unsupported; cannot send viewport"); } catch (eVp0) {}
      return;
    }
    try {
      var b = map.getBounds();
      var sw = b.getSouthWest();
      var ne = b.getNorthEast();
      var msgObj = {
        type: "mapDisplayViewport",
        bounds: { west: sw.lng, south: sw.lat, east: ne.lng, north: ne.lat },
        zoom: map.getZoom()
      };
      if (typeof CBOR === "undefined" || typeof CBOR.encode !== "function") { return; }
      var encoded = CBOR.encode(msgObj);
      var gz = await mapDisplayGzipEncode(encoded);
      var buf = gz.buffer.slice(gz.byteOffset, gz.byteOffset + gz.byteLength);
      ws.send(buf);
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

  var mapLoaded = false;

  var beforeUnloadListener = function () {
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
  };
  window.addEventListener("beforeunload", beforeUnloadListener);

  function syncStyle() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    bumpMapIconRasterGen();
    lastDark = d;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function () {
      didFit = !!skipAutoFitBounds;
      installFromState();
    });
  }
  var observer = null;
  if (document.body) {
    observer = new MutationObserver(syncStyle);
    observer.observe(document.body, { attributes: true, attributeFilter: ["data-theme"] });
  }
  var storageListener = function (ev) {
    if (ev.key !== "theme") { return; }
    syncStyle();
  };
  window.addEventListener("storage", storageListener);

  var _instance = {
    start: function () {
      shuttingDown = false;
      connectWebSocket();
    },
    flyTo: function (lng, lat, zoom) {
      if (!map || typeof map.flyTo !== "function") { return; }
      try {
        map.flyTo({ center: [lng, lat], zoom: zoom, essential: true });
      } catch (eApi0) {}
    },
    unproject: function (x, y) {
      if (!map || typeof map.unproject !== "function") { return null; }
      try {
        var ll = map.unproject([x, y]);
        return { lng: ll.lng, lat: ll.lat };
      } catch (eApi1) { return null; }
    },
    isReady: function () { return mapLoaded; },
    destroy: function () {
      shuttingDown = true;
      stopTick();
      clearReconnectTimer();
      if (boundsDebounceTimer) {
        try { clearTimeout(boundsDebounceTimer); } catch (e) {}
        boundsDebounceTimer = 0;
      }
      if (ws) {
        ws.onopen = ws.onmessage = ws.onerror = ws.onclose = null;
        try { ws.close(); } catch (e) {}
        ws = null;
      }
      window.removeEventListener("beforeunload", beforeUnloadListener);
      window.removeEventListener("storage", storageListener);
      if (observer) {
        try { observer.disconnect(); } catch (e) {}
      }
      if (map) {
        try { removeDynamicLayers(); } catch (e) {}
        try {
          map.off("moveend", scheduleSendViewportBounds);
          map.off("zoomend", scheduleSendViewportBounds);
          map.off("rotateend", scheduleSendViewportBounds);
          map.off("pitchend", scheduleSendViewportBounds);
          map.off("resize", scheduleSendViewportBounds);
        } catch (e) {}
        try { map.remove(); } catch (e) {}
        map = null;
      }
      if (window["mapDisplay_" + suffix] === _instance) {
        window["mapDisplay_" + suffix] = null;
      }
    }
  };
  window["mapDisplay_" + suffix] = _instance;
  mapEl.mapDisplayInstance = _instance;

  map.on("load", function () {
    try {
      map.addControl(new maplibregl.NavigationControl(), "top-right");
    } catch (eNav0) {}
    mapLoaded = true;
    try {
      document.dispatchEvent(new CustomEvent("mapDisplayReady", { detail: { suffix: suffix } }));
    } catch (eRdy0) {}
    if (!deferStart) {
      connectWebSocket();
    }
  });

  }

  function mapDisplayScheduleInit() {
    if (typeof requestAnimationFrame === "function") {
      requestAnimationFrame(function () {
        requestAnimationFrame(mapDisplayRunInit);
      });
    } else {
      setTimeout(mapDisplayRunInit, 0);
    }
  }
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", mapDisplayScheduleInit);
  } else {
    mapDisplayScheduleInit();
  }
})();`

	return Group([]Node{
		StyleEl(Raw(mapCtrlCSS)),
		Div(
			ID(mapElID),
			Class(classes),
			Attr("x-data", "{}"),
			Attr("x-on:destroy", "if ($el.mapDisplayInstance) { $el.mapDisplayInstance.destroy(); }"),
		),
		Script(Raw(initJS)),
	})
}
