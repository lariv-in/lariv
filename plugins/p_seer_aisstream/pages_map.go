package p_seer_aisstream

import (
	"context"
	"strconv"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const mapLibreCDNVersion = "4.7.1"

type aisStreamMapLibreExtraHead struct {
	components.Page
}

func (e *aisStreamMapLibreExtraHead) GetKey() string     { return e.Key }
func (e *aisStreamMapLibreExtraHead) GetRoles() []string { return e.Roles }

func (e *aisStreamMapLibreExtraHead) Build(ctx context.Context) Node {
	base := "https://unpkg.com/maplibre-gl@" + mapLibreCDNVersion + "/dist/"
	return Group([]Node{
		Link(
			Href(base+"maplibre-gl.css"),
			Rel("stylesheet"),
			CrossOrigin("anonymous"),
		),
		Script(
			Src(base+"maplibre-gl.js"),
			CrossOrigin("anonymous"),
		),
	})
}

type aisStreamMapSidebarLink struct {
	components.Page
}

func (e *aisStreamMapSidebarLink) GetKey() string     { return e.Key }
func (e *aisStreamMapSidebarLink) GetRoles() []string { return e.Roles }

func (e *aisStreamMapSidebarLink) Build(ctx context.Context) Node {
	href, err := lago.RoutePath("seer_aisstream.MapRoute", nil)(ctx)
	if err != nil || href == "" {
		href = AppUrl + "map/"
	}
	return Li(
		A(Href(href), Attr("data-hx-boost", "false"),
			components.Render(components.Icon{Name: "map-pin", Classes: "heroicon-sm"}, ctx),
			Text(" Ship map"),
		),
	)
}

type aisStreamMapLibreMount struct {
	components.Page
}

func (e *aisStreamMapLibreMount) GetKey() string     { return e.Key }
func (e *aisStreamMapLibreMount) GetRoles() []string { return e.Roles }

func (e *aisStreamMapLibreMount) Build(ctx context.Context) Node {
	apiURL := AppUrl + "api/vessels/"
	pollMs := 8000
	initJS := `(function(){
  var vesselsURL = ` + strconv.Quote(apiURL) + `;
  var pollMs = ` + strconv.Itoa(pollMs) + `;
  var mapEl = document.getElementById("aisstream-maplibre-map");
  if (!mapEl || typeof maplibregl === "undefined") { return; }
  var styleLight = "https://demotiles.maplibre.org/style.json";
  var styleDark = "https://tiles.openfreemap.org/styles/dark";
  function themeIsDark() {
    try {
      var b = document.body && document.body.getAttribute("data-theme");
      if (b === "dark") { return true; }
      if (b === "light") { return false; }
    } catch (e0) {}
    try {
      var t = localStorage.getItem("theme");
      if (t === "dark") { return true; }
      if (t === "light") { return false; }
    } catch (e1) {}
    return false;
  }
  var lastDark = themeIsDark();
  var map = new maplibregl.Map({
    container: mapEl,
    style: lastDark ? styleDark : styleLight,
    center: [5, 50],
    zoom: 4
  });
  map.addControl(new maplibregl.NavigationControl(), "top-right");

  var shipReady = false;
  // Same Iconify URL as components.Icon — SVG cannot be used directly as a WebGL texture; rasterize.
  var heroiconName = "map-pin";
  var iconRequestSeq = 0;
  function addShipIcon(cb) {
    if (shipReady && map.hasImage && map.hasImage("aisstream-ship")) { cb(); return; }
    var myReq = ++iconRequestSeq;
    var url = "https://api.iconify.design/heroicons/" + heroiconName + ".svg";
    var img = new Image();
    img.crossOrigin = "anonymous";
    function fail() { shipReady = false; try { cb(); } catch (e) {} }
    function doneBitmap(bmp) {
      if (myReq !== iconRequestSeq) {
        if (bmp && typeof bmp.close === "function") { try { bmp.close(); } catch (e) {} }
        return;
      }
      if (typeof map.isStyleLoaded === "function" && !map.isStyleLoaded()) {
        if (bmp && typeof bmp.close === "function") { try { bmp.close(); } catch (e2) {} }
        return;
      }
      try {
        if (map.hasImage && map.hasImage("aisstream-ship")) {
          map.removeImage("aisstream-ship");
        }
        map.addImage("aisstream-ship", bmp, { pixelRatio: 2 });
        shipReady = true;
        cb();
      } catch (e) {
        fail();
      }
    }
    img.onload = function() {
      if (myReq !== iconRequestSeq) { return; }
      var w = 64, h = 64;
      var c = document.createElement("canvas");
      c.width = w;
      c.height = h;
      var g = c.getContext("2d", { willReadFrequently: true });
      if (!g) { fail(); return; }
      g.clearRect(0, 0, w, h);
      try { g.drawImage(img, 0, 0, w, h); } catch (e) { if (myReq === iconRequestSeq) { fail(); } return; }
      if (myReq !== iconRequestSeq) { return; }
      if (typeof createImageBitmap === "function") {
        createImageBitmap(c, { premultiplyAlpha: "default" })
          .then(function(bmp) { doneBitmap(bmp); }, function() { if (myReq === iconRequestSeq) { fail(); } });
      } else {
        if (myReq !== iconRequestSeq) { return; }
        if (typeof map.isStyleLoaded === "function" && !map.isStyleLoaded()) { return; }
        try {
          if (map.hasImage && map.hasImage("aisstream-ship")) { map.removeImage("aisstream-ship"); }
          map.addImage("aisstream-ship", c, { pixelRatio: 2 });
          shipReady = true;
          cb();
        } catch (e2) { if (myReq === iconRequestSeq) { fail(); } }
      }
    };
    img.onerror = function() { if (myReq === iconRequestSeq) { fail(); } };
    img.src = url;
  }

  var sourceId = "aisstream-vessels";
  var layerId = "aisstream-vessels-sym";

  var prevByMmsi = {};
  var currByMmsi = {};
  var blendStart = 0;
  var blendMs = pollMs;
  var rafId = 0;
  var moveTimer = null;
  var fetchStarted = false;

  function boundsQuery() {
    var b = map.getBounds();
    return "lamin=" + encodeURIComponent(b.getSouth().toFixed(4)) +
      "&lomin=" + encodeURIComponent(b.getWest().toFixed(4)) +
      "&lamax=" + encodeURIComponent(b.getNorth().toFixed(4)) +
      "&lomax=" + encodeURIComponent(b.getEast().toFixed(4));
  }

  function lerp(a, b, t) { return a + (b - a) * t; }

  function buildFeatures(t) {
    if (t > 1) { t = 1; }
    var feats = [];
    Object.keys(currByMmsi).forEach(function(mmsi) {
      var cur = currByMmsi[mmsi];
      var pr = prevByMmsi[mmsi];
      var lat = cur.lat;
      var lng = cur.lng;
      if (pr && pr.lat != null && pr.lng != null) {
        lat = lerp(pr.lat, cur.lat, t);
        lng = lerp(pr.lng, cur.lng, t);
      }
      var cog = (typeof cur.cog === "number" ? cur.cog : 0) || 0;
      var label = (cur.name && String(cur.name).trim()) || mmsi;
      feats.push({
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties: { mmsi: mmsi, cog: cog, title: label }
      });
    });
    return { type: "FeatureCollection", features: feats };
  }

  function refreshSource() {
    var src = map.getSource(sourceId);
    if (!src) { return; }
    var elapsed = performance.now() - blendStart;
    var t = blendMs > 0 ? elapsed / blendMs : 1;
    src.setData(buildFeatures(t));
  }

  function tick() {
    rafId = 0;
    refreshSource();
    var elapsed = performance.now() - blendStart;
    if (elapsed < blendMs + 50) {
      rafId = requestAnimationFrame(tick);
    }
  }

  function scheduleTick() {
    if (rafId) { return; }
    rafId = requestAnimationFrame(tick);
  }

  function applySnapshot(data) {
    prevByMmsi = currByMmsi;
    currByMmsi = {};
    (data.vessels || []).forEach(function(v) {
      if (v == null || !v.mmsi) { return; }
      currByMmsi[String(v.mmsi)] = v;
    });
    blendStart = performance.now();
    scheduleTick();
  }

  function layerSpec() {
    var dark = themeIsDark();
    return {
      id: layerId,
      type: "symbol",
      source: sourceId,
      layout: {
        "icon-image": "aisstream-ship",
        "icon-size": 0.5,
        "icon-rotate": ["get", "cog"],
        "icon-rotation-alignment": "map",
        "icon-allow-overlap": true,
        "icon-ignore-placement": true,
        "text-field": ["get", "title"],
        "text-size": 11,
        "text-max-width": 8,
        "text-offset": [0, 1.3],
        "text-anchor": "top",
        "text-rotation-alignment": "viewport",
        "text-pitch-alignment": "viewport",
        "text-optional": true
      },
      paint: dark
        ? {
            "text-color": "#f1f5f9",
            "text-halo-color": "#020617",
            "text-halo-width": 2,
            "text-halo-blur": 0.5
          }
        : {
            "text-color": "#0f172a",
            "text-halo-color": "#ffffff",
            "text-halo-width": 2,
            "text-halo-blur": 0.5
          }
    };
  }

  function fetchVessels() {
    var u = vesselsURL + "?" + boundsQuery();
    fetch(u, { credentials: "same-origin", headers: { Accept: "application/json" } })
      .then(function(res) { return res.json().then(function(data) { return { ok: res.ok, st: res.status, data: data }; }); })
      .then(function(wrap) {
        if (!wrap || !wrap.data) { return; }
        if (wrap.data.error) { return; }
        if (!wrap.ok) { return; }
        applySnapshot(wrap.data);
      })
      .catch(function() {});
  }

  function installOverlay() {
    addShipIcon(function() {
      try {
        if (map.getLayer(layerId)) { map.removeLayer(layerId); }
      } catch (eL) {}
      try {
        if (map.getSource(sourceId)) { map.removeSource(sourceId); }
      } catch (eS) {}
      map.addSource(sourceId, { type: "geojson", data: { type: "FeatureCollection", features: [] } });
      map.addLayer(layerSpec());
      refreshSource();
      if (!fetchStarted) {
        fetchStarted = true;
        fetchVessels();
        setInterval(fetchVessels, pollMs);
      }
    });
  }

  map.on("load", function() { installOverlay(); });
  map.on("moveend", function() {
    if (moveTimer) { clearTimeout(moveTimer); }
    moveTimer = setTimeout(fetchVessels, 400);
  });

  function syncStyleToTheme() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    shipReady = false;
    iconRequestSeq++;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function() { installOverlay(); });
  }
  if (document.body) {
    new MutationObserver(syncStyleToTheme).observe(document.body, { attributes: true, attributeFilter: ["data-theme"] });
  }
  window.addEventListener("storage", function(ev) {
    if (ev.key !== "theme") { return; }
    syncStyleToTheme();
  });

  map.on("click", layerId, function(e) {
    var f = e.features && e.features[0];
    if (!f) { return; }
    var p = f.properties || {};
    new maplibregl.Popup({ offset: 12 }).setLngLat(e.lngLat)
      .setHTML((p.title || "") + " — MMSI " + (p.mmsi || "")).addTo(map);
  });
  map.on("mouseenter", layerId, function() { map.getCanvas().style.cursor = "pointer"; });
  map.on("mouseleave", layerId, function() { map.getCanvas().style.cursor = ""; });
})();`

	return Group([]Node{
		Div(
			ID("aisstream-maplibre-map"),
			Class("w-full h-[70vh] min-h-96 rounded-box border border-base-300 z-0"),
		),
		Script(Raw(initJS)),
	})
}

func registerAISStreamMapPages() {
	lago.RegistryPage.Register("seer_aisstream.MapPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_aisstream.MapPageShell"},
		ExtraHead: []components.PageInterface{
			&aisStreamMapLibreExtraHead{Page: components.Page{Key: "seer_aisstream.MapLibreHead"}},
		},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_aisstream.Menu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_aisstream.MapPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("AISStream live vessels")},
					&components.FieldText{
						Page:    components.Page{Key: "seer_aisstream.MapBlurb"},
						Getter:  getters.Static("Vessel positions from aisstream.io (BETA, no SLA). API key in [Plugins.p_seer_aisstream] in seer.toml; stream runs server-side only. Not for navigation or safety."),
						Classes: "text-sm text-base-content/80 max-w-3xl",
					},
					&aisStreamMapLibreMount{Page: components.Page{Key: "seer_aisstream.MapLibreMount"}},
				},
			},
		},
	})
}
