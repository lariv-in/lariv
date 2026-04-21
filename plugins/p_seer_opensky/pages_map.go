package p_seer_opensky

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

type openskyMapLibreExtraHead struct {
	components.Page
}

func (e *openskyMapLibreExtraHead) GetKey() string     { return e.Key }
func (e *openskyMapLibreExtraHead) GetRoles() []string { return e.Roles }

func (e *openskyMapLibreExtraHead) Build(ctx context.Context) Node {
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

type openskyMapSidebarLink struct {
	components.Page
}

func (e *openskyMapSidebarLink) GetKey() string     { return e.Key }
func (e *openskyMapSidebarLink) GetRoles() []string { return e.Roles }

func (e *openskyMapSidebarLink) Build(ctx context.Context) Node {
	href, err := lago.RoutePath("seer_opensky.MapRoute", nil)(ctx)
	if err != nil || href == "" {
		href = AppUrl + "map/"
	}
	return Li(
		A(Href(href), Attr("data-hx-boost", "false"),
			components.Render(components.Icon{Name: "map-pin", Classes: "heroicon-sm"}, ctx),
			Text(" Flight map"),
		),
	)
}

type openskyMapLibreMount struct {
	components.Page
}

func (e *openskyMapLibreMount) GetKey() string     { return e.Key }
func (e *openskyMapLibreMount) GetRoles() []string { return e.Roles }

func (e *openskyMapLibreMount) Build(ctx context.Context) Node {
	apiURL := AppUrl + "api/states/"
	pollMs := 10000
	initJS := `(function(){
  var statesURL = ` + strconv.Quote(apiURL) + `;
  var pollMs = ` + strconv.Itoa(pollMs) + `;
  var mapEl = document.getElementById("opensky-maplibre-map");
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
    center: [10, 45],
    zoom: 4
  });
  map.addControl(new maplibregl.NavigationControl(), "top-right");

  var planeReady = false;
  function addPlaneIcon(cb) {
    // setStyle() drops images; hasImage is the real check (planeReady can be stale).
    if (planeReady && map.hasImage && map.hasImage("opensky-plane")) { cb(); return; }
    var c = document.createElement("canvas");
    c.width = 64; c.height = 64;
    var g = c.getContext("2d");
    g.translate(32, 32);
    g.rotate(-Math.PI / 2);
    g.fillStyle = "#38bdf8";
    g.strokeStyle = "#0f172a";
    g.lineWidth = 2;
    g.beginPath();
    g.moveTo(24, 0);
    g.lineTo(-20, 16);
    g.lineTo(-8, 16);
    g.lineTo(-24, 28);
    g.lineTo(-24, -28);
    g.lineTo(-8, -16);
    g.lineTo(-20, -16);
    g.closePath();
    g.fill();
    g.stroke();
    createImageBitmap(c).then(function(bmp) {
      try {
        if (map.hasImage && map.hasImage("opensky-plane")) {
          map.removeImage("opensky-plane");
        }
      } catch (eRm) {}
      map.addImage("opensky-plane", bmp, { pixelRatio: 2 });
      planeReady = true;
      cb();
    }).catch(function() { cb(); });
  }

  var sourceId = "opensky-aircraft";
  var layerId = "opensky-aircraft-sym";

  var prevByIcao = {};
  var currByIcao = {};
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

  function headingFor(a, prev) {
    if (a.heading != null && typeof a.heading === "number") { return a.heading; }
    if (prev && prev.lat != null && prev.lng != null) {
      var φ1 = prev.lat * Math.PI / 180, φ2 = a.lat * Math.PI / 180;
      var Δλ = (a.lng - prev.lng) * Math.PI / 180;
      var y = Math.sin(Δλ) * Math.cos(φ2);
      var x = Math.cos(φ1) * Math.sin(φ2) - Math.sin(φ1) * Math.cos(φ2) * Math.cos(Δλ);
      var θ = Math.atan2(y, x);
      return (θ * 180 / Math.PI + 360) % 360;
    }
    return 0;
  }

  function buildFeatures(t) {
    if (t > 1) { t = 1; }
    var feats = [];
    Object.keys(currByIcao).forEach(function(icao) {
      var cur = currByIcao[icao];
      var pr = prevByIcao[icao];
      var lat = cur.lat;
      var lng = cur.lng;
      if (pr && pr.lat != null && pr.lng != null) {
        lat = lerp(pr.lat, cur.lat, t);
        lng = lerp(pr.lng, cur.lng, t);
      }
      var h = headingFor(cur, pr);
      var title = cur.callsign || icao;
      feats.push({
        type: "Feature",
        geometry: { type: "Point", coordinates: [lng, lat] },
        properties: {
          icao24: icao,
          heading: h,
          title: title,
          onGround: !!cur.onGround
        }
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
    prevByIcao = currByIcao;
    currByIcao = {};
    (data.aircraft || []).forEach(function(a) {
      if (!a.icao24) { return; }
      currByIcao[a.icao24] = a;
    });
    blendStart = performance.now();
    scheduleTick();
  }

  function aircraftLayerSpec() {
    var dark = themeIsDark();
    return {
      id: layerId,
      type: "symbol",
      source: sourceId,
      layout: {
        "icon-image": "opensky-plane",
        "icon-size": 0.55,
        "icon-rotate": ["get", "heading"],
        "icon-rotation-alignment": "map",
        "icon-allow-overlap": true,
        "icon-ignore-placement": true,
        "text-field": ["get", "title"],
        "text-size": 11,
        "text-max-width": 8,
        "text-offset": [0, 1.35],
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

  function fetchStates() {
    var u = statesURL + "?" + boundsQuery();
    fetch(u, { credentials: "same-origin", headers: { Accept: "application/json" } })
      .then(function(res) {
        if (res.status === 429) {
          var ra = res.headers.get("Retry-After") || res.headers.get("X-Rate-Limit-Retry-After-Seconds");
          var wait = ra ? parseInt(ra, 10) * 1000 : pollMs * 2;
          setTimeout(fetchStates, isNaN(wait) ? pollMs * 2 : wait);
          return null;
        }
        if (!res.ok) { throw new Error("HTTP " + res.status); }
        return res.json();
      })
      .then(function(data) {
        if (!data) { return; }
        applySnapshot(data);
      })
      .catch(function() {});
  }

  function installAircraftOverlay() {
    addPlaneIcon(function() {
      try {
        if (map.getLayer(layerId)) { map.removeLayer(layerId); }
      } catch (eL) {}
      try {
        if (map.getSource(sourceId)) { map.removeSource(sourceId); }
      } catch (eS) {}
      map.addSource(sourceId, { type: "geojson", data: { type: "FeatureCollection", features: [] } });
      map.addLayer(aircraftLayerSpec());
      refreshSource();
      if (!fetchStarted) {
        fetchStarted = true;
        fetchStates();
        setInterval(fetchStates, pollMs);
      }
    });
  }

  map.on("load", function() {
    installAircraftOverlay();
  });

  map.on("moveend", function() {
    if (moveTimer) { clearTimeout(moveTimer); }
    moveTimer = setTimeout(fetchStates, 400);
  });

  function syncStyleToTheme() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    planeReady = false;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function() {
      installAircraftOverlay();
    });
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
    var html = (p.title || "") + " — " + (p.icao24 || "");
    new maplibregl.Popup({ offset: 12 }).setLngLat(e.lngLat).setHTML(html).addTo(map);
  });
  map.on("mouseenter", layerId, function() { map.getCanvas().style.cursor = "pointer"; });
  map.on("mouseleave", layerId, function() { map.getCanvas().style.cursor = ""; });
})();`

	return Group([]Node{
		Div(
			ID("opensky-maplibre-map"),
			Class("w-full h-[70vh] min-h-96 rounded-box border border-base-300 z-0"),
		),
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
			lago.DynamicPage{Name: "seer_opensky.Menu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_opensky.MapPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("OpenSky live flights")},
					&components.FieldText{
						Page:    components.Page{Key: "seer_opensky.MapBlurb"},
						Getter:  getters.Static("Aircraft positions from OpenSky Network (viewport bounding box; polls every 10s). Optional OAuth: [Plugins.p_seer_opensky] clientID and clientSecret, or credentialsFile pointing at JSON with client_id and client_secret. TOML overrides file when both set. Animation blends between snapshots."),
						Classes: "text-sm text-base-content/80 max-w-3xl",
					},
					&openskyMapLibreMount{Page: components.Page{Key: "seer_opensky.MapLibreMount"}},
				},
			},
		},
	})
}
