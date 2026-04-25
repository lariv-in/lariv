package p_seer_opensky

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MapLibre (pinned) for map page only; matches [p_seer_gdelt/pages_map.go].
const openskyMapLibreCDNVersion = "4.7.1"

type openskyMapLibreExtraHead struct{ components.Page }

func (e *openskyMapLibreExtraHead) GetKey() string     { return e.Key }
func (e *openskyMapLibreExtraHead) GetRoles() []string { return e.Roles }

func (e *openskyMapLibreExtraHead) Build(ctx context.Context) Node {
	base := "https://unpkg.com/maplibre-gl@" + openskyMapLibreCDNVersion + "/dist/"
	return Group([]Node{
		Link(Href(base+"maplibre-gl.css"), Rel("stylesheet"), CrossOrigin("anonymous")),
		Script(Src(base+"maplibre-gl.js"), CrossOrigin("anonymous")),
	})
}

// openskyMapMenuLink uses full document navigation so the inline map script runs.
type openskyMapMenuLink struct{ components.Page }

func (e *openskyMapMenuLink) GetKey() string     { return e.Key }
func (e *openskyMapMenuLink) GetRoles() []string { return e.Roles }

func (e *openskyMapMenuLink) Build(ctx context.Context) Node {
	// Use states/map/ so a root-relative path does not join under .../states/ (which
	// would hit StateDetail with id=map). Also /seer-opensky/map/ remains valid via MapRoute.
	href, err := lago.RoutePath("seer_opensky.MapRouteUnderStates", nil)(ctx)
	if err != nil || href == "" {
		slog.Error("p_seer_opensky: MapRouteUnderStates path failed", "error", err)
		href = AppUrl + "states/map/"
	}
	if !strings.HasPrefix(href, "/") {
		href = "/" + strings.TrimPrefix(href, "/")
	}
	return Li(
		A(Href(href), Attr("data-hx-boost", "false"),
			components.Render(components.Icon{Name: "map-pin", Classes: "heroicon-sm"}, ctx),
			Text(" Map"),
		),
	)
}

type openskyMapLibreMount struct{ components.Page }

func (e *openskyMapLibreMount) GetKey() string     { return e.Key }
func (e *openskyMapLibreMount) GetRoles() []string { return e.Roles }

func (e *openskyMapLibreMount) Build(ctx context.Context) Node {
	var aircraft []openSkyMapAircraft
	if v := ctx.Value(openskyMapAircraftKey); v != nil {
		if a, ok := v.([]openSkyMapAircraft); ok {
			aircraft = a
		}
	}
	jsonBytes, err := json.Marshal(aircraft)
	if err != nil {
		jsonBytes = []byte("[]")
	}
	dataPath, err := lago.RoutePath("seer_opensky.MapDataRoute", nil)(ctx)
	if err != nil || dataPath == "" {
		slog.Error("p_seer_opensky: MapDataRoute path failed", "error", err)
		dataPath = AppUrl + "map/data/"
	}
	if !strings.HasPrefix(dataPath, "/") {
		dataPath = "/" + strings.TrimPrefix(dataPath, "/")
	}
	dataPathBytes, err := json.Marshal(dataPath)
	if err != nil {
		dataPathBytes = []byte(`"/seer-opensky/map/data/"`)
	}
	refreshMS := int64(0)
	if p := Config.PollEvery(); p > 0 {
		refreshMS = p.Milliseconds()
	}
	refreshMSBytes, err := json.Marshal(refreshMS)
	if err != nil {
		refreshMSBytes = []byte("0")
	}
	// Positions refresh via JSON polling; velocity (m/s) animates between refreshes.
	initJS := `(function(){
  var je = document.getElementById("opensky-map-aircraft-json");
  var aircraft = [];
  try { aircraft = JSON.parse(je ? je.textContent || "[]" : "[]"); } catch (e0) {}
  var dataURL = ` + string(dataPathBytes) + `;
  var refreshMS = ` + string(refreshMSBytes) + `;
  var mapEl = document.getElementById("opensky-maplibre-map");
  if (!mapEl || typeof maplibregl === "undefined") { return; }
  var styleLight = "https://demotiles.maplibre.org/style.json";
  var styleDark = "https://tiles.openfreemap.org/styles/dark";
  function themeIsDark() {
    try {
      var t = localStorage.getItem("theme");
      if (t === "dark" || t === "light") { return t === "dark"; }
    } catch (e) {}
    return document.body && document.body.getAttribute("data-theme") === "dark";
  }
  var lastDark = themeIsDark();
  var map = new maplibregl.Map({ container: mapEl, style: lastDark ? styleDark : styleLight, center: [0, 20], zoom: 1.5 });
  map.addControl(new maplibregl.NavigationControl(), "top-right");
  var popupOpen = null;
  function closePopup() { if (popupOpen) { popupOpen.remove(); popupOpen = null; } }
  var workFeatures = [];
  var tickTimer = 0;
  var timePrev = 0;
  var clickBound = false;
  var animationTickMS = 200;
  function addPlaneImage() {
    if (!map || typeof map.addImage !== "function") { return; }
    var sz = 64, c = document.createElement("canvas");
    c.width = sz; c.height = sz;
    var x = c.getContext("2d");
    if (!x) { return; }
    if (map.hasImage && map.hasImage("opensky-plane") && map.removeImage) { try { map.removeImage("opensky-plane"); } catch (e) {} }
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
    try { map.addImage("opensky-plane", idata, { pixelRatio: 1 }); } catch (e1) { try { map.addImage("opensky-plane", idata); } catch (e2) {} }
  }
  function extrapolateCoordinate(lng, lat, heading, velocityMps, lastContact) {
    lng = +lng; lat = +lat;
    var vm = +velocityMps, hd = +heading, lc = +lastContact;
    if (!isFinite(lng) || !isFinite(lat)) { return [0, 0]; }
    if (!isFinite(vm) || vm <= 0 || !isFinite(lc) || lc <= 0) { return [lng, lat]; }
    if (!isFinite(hd)) { hd = 0; }
    var nowSec = Date.now() / 1000;
    var dt = Math.max(0, nowSec - lc);
    var brad = hd * Math.PI/180;
    var m = vm * dt;
    var latrad = lat * Math.PI/180;
    var dN = m * Math.cos(brad);
    var dE = m * Math.sin(brad);
    return [lng + dE/(111320*Math.max(0.01, Math.cos(latrad))), lat + dN/111320];
  }
  function buildFC(ac) {
    return {
      type: "FeatureCollection",
      features: (ac || []).map(function (a) {
        var coords = extrapolateCoordinate(a.lng, a.lat, a.heading, a.velocityMps, a.lastContact);
        return {
          type: "Feature", id: a.icao24,
          geometry: { type: "Point", coordinates: coords },
          properties: {
            heading: a.heading, title: a.title, detailPath: a.detailPath, icao: a.icao24, velocityMps: a.velocityMps || 0, lastContact: a.lastContact || 0
          }
        };
      })
    };
  }
  function cloneWorkFeatures(feats) {
    return (feats || []).map(function (f) {
      return {
        type: "Feature", id: f.id, geometry: { type: f.geometry.type, coordinates: f.geometry.coordinates.slice() },
        properties: { heading: f.properties.heading, title: f.properties.title, detailPath: f.properties.detailPath, icao: f.properties.icao, velocityMps: f.properties.velocityMps, lastContact: f.properties.lastContact }
      };
    });
  }
  function mergeNewerFeatures(nextFeatures) {
    var current = {}, out = [], i, f, key, old, nextLC, oldLC;
    for (i = 0; i < workFeatures.length; i++) {
      f = workFeatures[i];
      if (!f || !f.properties) { continue; }
      key = f.properties.icao || f.id;
      if (key) { current[key] = f; }
    }
    for (i = 0; i < nextFeatures.length; i++) {
      f = nextFeatures[i];
      if (!f || !f.properties) { continue; }
      key = f.properties.icao || f.id;
      old = key ? current[key] : null;
      nextLC = +(f.properties.lastContact || 0);
      oldLC = old && old.properties ? +(old.properties.lastContact || 0) : -1;
      out.push(old && nextLC <= oldLC ? old : f);
    }
    return out;
  }
  function removeAircraft() {
    closePopup();
    if (tickTimer) { try { window.clearInterval(tickTimer); } catch (e) {} tickTimer = 0; }
    if (map.getLayer("opensky-planes")) { map.removeLayer("opensky-planes"); }
    if (map.getSource("opensky-aircraft")) { map.removeSource("opensky-aircraft"); }
  }
  function startAnimation() {
    if (tickTimer || !workFeatures || !workFeatures.length) { return; }
    timePrev = 0;
    tickTimer = window.setInterval(tick, animationTickMS);
  }
  function tick() {
    if (!workFeatures || !workFeatures.length) { return; }
    var now = (window.performance && performance.now) ? performance.now() : Date.now();
    if (timePrev <= 0) { timePrev = now; return; }
    var dt = Math.max(0, Math.min(0.25, (now - timePrev) / 1000));
    timePrev = now;
    var i, f, c, latrad, brad, m, dN, dE, moved = false, vm, hd;
    for (i = 0; i < workFeatures.length; i++) {
      f = workFeatures[i];
      if (!f || !f.properties) { continue; }
      vm = +f.properties.velocityMps;
      if (!isFinite(vm) || vm <= 0) { continue; }
      c = f.geometry.coordinates;
      hd = +f.properties.heading;
      if (!isFinite(hd)) { hd = 0; }
      brad = hd * Math.PI/180;
      m = vm * dt;
      latrad = c[1] * Math.PI/180;
      dN = m * Math.cos(brad);
      dE = m * Math.sin(brad);
      c[1] = c[1] + dN/111320;
      c[0] = c[0] + dE/(111320*Math.max(0.01, Math.cos(latrad)));
      moved = true;
    }
    if (moved) {
      var src = map.getSource("opensky-aircraft");
      if (src && src.setData) { src.setData({ type: "FeatureCollection", features: workFeatures }); }
    }
  }
  function installAircraft() {
    removeAircraft();
    if (!aircraft || aircraft.length === 0) { workFeatures = []; return; }
    addPlaneImage();
    var g = buildFC(aircraft);
    workFeatures = cloneWorkFeatures(g.features);
    var b = new maplibregl.LngLatBounds();
    aircraft.forEach(function (a) { b.extend([a.lng, a.lat]); });
    timePrev = 0;
    map.addSource("opensky-aircraft", { type: "geojson", data: g, cluster: false });
    if (!map.hasImage || !map.hasImage("opensky-plane")) { addPlaneImage(); }
    map.addLayer({
      id: "opensky-planes", type: "symbol", source: "opensky-aircraft", minzoom: 0,
      layout: {
        "icon-image": "opensky-plane", "icon-size": 0.5, "icon-allow-overlap": true, "icon-ignore-placement": true,
        "icon-rotate": ["get", "heading"], "icon-rotation-alignment": "map"
      }
    });
    try { map.fitBounds(b, { padding: 48, maxZoom: 9 }); } catch (e) {}
    if (!clickBound) {
      clickBound = true;
      map.on("click", "opensky-planes", function (e) {
        var feats = map.queryRenderedFeatures(e.point, { layers: ["opensky-planes"] });
        if (!feats.length) { return; }
        var p = feats[0].properties || {};
        var cc = feats[0].geometry.coordinates.slice();
        var wrap = document.createElement("div");
        if (p.detailPath) { var a = document.createElement("a"); a.href = p.detailPath; a.className = "link link-primary"; a.textContent = p.title || p.icao || "State"; wrap.appendChild(a); }
        else { wrap.textContent = p.title || p.icao || "Aircraft"; }
        closePopup();
        popupOpen = new maplibregl.Popup({ offset: 20 }).setLngLat(cc).setDOMContent(wrap).addTo(map);
      });
      map.on("mouseenter", "opensky-planes", function () { map.getCanvas().style.cursor = "pointer"; });
      map.on("mouseleave", "opensky-planes", function () { map.getCanvas().style.cursor = ""; });
    }
    startAnimation();
  }
  function replaceAircraft(ac) {
    aircraft = Array.isArray(ac) ? ac : [];
    var g = buildFC(aircraft);
    workFeatures = mergeNewerFeatures(cloneWorkFeatures(g.features));
    timePrev = 0;
    var src = map.getSource("opensky-aircraft");
    if (src && src.setData) {
      src.setData({ type: "FeatureCollection", features: workFeatures });
      if (workFeatures.length) { startAnimation(); }
      return;
    }
    if (workFeatures.length) { installAircraft(); }
  }
  function refreshAircraft() {
    if (!dataURL) { return; }
    fetch(dataURL, { headers: { "Accept": "application/json" }, credentials: "same-origin", cache: "no-store" })
      .then(function (r) { if (!r.ok) { throw new Error("HTTP " + r.status); } return r.json(); })
      .then(replaceAircraft)
      .catch(function () {});
  }
  map.on("load", installAircraft);
  if (refreshMS > 0) { window.setInterval(refreshAircraft, refreshMS); }
  function syncStyle() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function () { installAircraft(); });
  }
  if (document.body) { new MutationObserver(syncStyle).observe(document.body, { attributes: true, attributeFilter: ["data-theme"] }); }
  window.addEventListener("storage", function (ev) { if (ev.key === "theme") { syncStyle(); } });
})();`
	return Group([]Node{
		Div(ID("opensky-maplibre-map"), Class("w-full h-[min(80vh,720px)] min-h-80 rounded-box border border-base-300 z-0")),
		Script(Type("application/json"), ID("opensky-map-aircraft-json"), Raw(string(jsonBytes))),
		Script(Raw(initJS)),
	})
}

func registerOpenSkyMapPages() {
	lago.RegistryPage.Register("seer_opensky.MapPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_opensky.MapPageShell"},
		ExtraHead: []components.PageInterface{
			&openskyMapLibreExtraHead{Page: components.Page{Key: "seer_opensky.MapLibreHead"}},
		},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_opensky.AppMenu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_opensky.MapPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("OpenSky live map")},
					&components.FieldText{
						Page:    components.Page{Key: "seer_opensky.MapBlurb"},
						Getter:  getters.Static("Only aircraft whose latest state is not on ground, has a non-null, non-zero ground velocity (m/s), and last_contact within [Plugins.p_seer_opensky] mapLastContactMaxAge (default 1m; 0s = no time filter). Up to 50k ICAOs. Heading = bearing from the previous point to the latest, or true track. Client motion uses that velocity. PostgreSQL with position column required. Reload to resync with the server."),
						Classes: "text-sm text-base-content/80 max-w-3xl",
					},
					&openskyMapLibreMount{Page: components.Page{Key: "seer_opensky.MapLibreMount"}},
				},
			},
		},
	})
}
