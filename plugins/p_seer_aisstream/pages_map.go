package p_seer_aisstream

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

const aisStreamMapLibreCDNVersion = "4.7.1"

type aisStreamMapLibreExtraHead struct{ components.Page }

func (e *aisStreamMapLibreExtraHead) GetKey() string     { return e.Key }
func (e *aisStreamMapLibreExtraHead) GetRoles() []string { return e.Roles }

func (e *aisStreamMapLibreExtraHead) Build(ctx context.Context) Node {
	base := "https://unpkg.com/maplibre-gl@" + aisStreamMapLibreCDNVersion + "/dist/"
	return Group([]Node{
		Link(Href(base+"maplibre-gl.css"), Rel("stylesheet"), CrossOrigin("anonymous")),
		Script(Src(base+"maplibre-gl.js"), CrossOrigin("anonymous")),
	})
}

type aisStreamMapMenuLink struct{ components.Page }

func (e *aisStreamMapMenuLink) GetKey() string     { return e.Key }
func (e *aisStreamMapMenuLink) GetRoles() []string { return e.Roles }

func (e *aisStreamMapMenuLink) Build(ctx context.Context) Node {
	href, err := lago.RoutePath("seer_aisstream.MapRoute", nil)(ctx)
	if err != nil || href == "" {
		slog.Error("p_seer_aisstream: MapRoute path failed", "error", err)
		href = AppUrl + "map/"
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

type aisStreamMapLibreMount struct{ components.Page }

func (e *aisStreamMapLibreMount) GetKey() string     { return e.Key }
func (e *aisStreamMapLibreMount) GetRoles() []string { return e.Roles }

func (e *aisStreamMapLibreMount) Build(ctx context.Context) Node {
	var vessels []aisStreamMapVessel
	if v := ctx.Value(aisStreamMapVesselsKey); v != nil {
		if a, ok := v.([]aisStreamMapVessel); ok {
			vessels = a
		}
	}
	jsonBytes, err := json.Marshal(vessels)
	if err != nil {
		jsonBytes = []byte("[]")
	}
	dataPath, err := lago.RoutePath("seer_aisstream.MapDataRoute", nil)(ctx)
	if err != nil || dataPath == "" {
		slog.Error("p_seer_aisstream: MapDataRoute path failed", "error", err)
		dataPath = AppUrl + "map/data/"
	}
	if !strings.HasPrefix(dataPath, "/") {
		dataPath = "/" + strings.TrimPrefix(dataPath, "/")
	}
	dataPathBytes, _ := json.Marshal(dataPath)
	refreshMS := int64(0)
	if p := Config.MapRefreshEvery(); p > 0 {
		refreshMS = p.Milliseconds()
	}
	refreshMSBytes, _ := json.Marshal(refreshMS)
	// Match OpenSky map behavior: unclustered rotating symbol markers with theme-aware basemap refresh.
	initJS := `(function(){
  var je = document.getElementById("aisstream-map-vessels-json");
  var vessels = [];
  try { vessels = JSON.parse(je ? je.textContent || "[]" : "[]"); } catch (e0) {}
  var dataURL = ` + string(dataPathBytes) + `;
  var refreshMS = ` + string(refreshMSBytes) + `;
  var mapEl = document.getElementById("aisstream-maplibre-map");
  if (!mapEl || typeof maplibregl === "undefined") { return; }
  var styleLight = "https://demotiles.maplibre.org/style.json";
  var styleDark = "https://tiles.openfreemap.org/styles/dark";
  function themeIsDark() {
    try { var t = localStorage.getItem("theme"); if (t === "dark" || t === "light") { return t === "dark"; } } catch (e) {}
    return document.body && document.body.getAttribute("data-theme") === "dark";
  }
  var lastDark = themeIsDark();
  var map = new maplibregl.Map({ container: mapEl, style: lastDark ? styleDark : styleLight, center: [0, 20], zoom: 1.5 });
  map.addControl(new maplibregl.NavigationControl(), "top-right");
  var popupOpen = null;
  function closePopup() { if (popupOpen) { popupOpen.remove(); popupOpen = null; } }
  var workFeatures = [];
  var clickBound = false;
  function addVesselImage() {
    if (!map || typeof map.addImage !== "function") { return; }
    var sz = 64, c = document.createElement("canvas");
    c.width = sz; c.height = sz;
    var x = c.getContext("2d");
    if (!x) { return; }
    if (map.hasImage && map.hasImage("aisstream-vessel") && map.removeImage) { try { map.removeImage("aisstream-vessel"); } catch (e) {} }
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
    try { map.addImage("aisstream-vessel", idata, { pixelRatio: 1 }); } catch (e1) { try { map.addImage("aisstream-vessel", idata); } catch (e2) {} }
  }
  function buildFC(vs) {
    return {
      type: "FeatureCollection",
      features: (vs || []).map(function (v) {
        return {
          type: "Feature", id: v.mmsi || v.id,
          geometry: { type: "Point", coordinates: [v.lng, v.lat] },
          properties: {
            heading: v.heading || 0, title: v.title, detailPath: v.detailPath, mmsi: v.mmsi, sog: v.sog || 0, timeUtc: v.timeUtc || 0
          }
        };
      })
    };
  }
  function cloneWorkFeatures(feats) {
    return (feats || []).map(function (f) {
      return {
        type: "Feature", id: f.id, geometry: { type: f.geometry.type, coordinates: f.geometry.coordinates.slice() },
        properties: { heading: f.properties.heading, title: f.properties.title, detailPath: f.properties.detailPath, mmsi: f.properties.mmsi, sog: f.properties.sog, timeUtc: f.properties.timeUtc }
      };
    });
  }
  function mergeNewerFeatures(nextFeatures) {
    var current = {}, out = [], i, f, key, old, nextT, oldT;
    for (i = 0; i < workFeatures.length; i++) {
      f = workFeatures[i];
      if (!f || !f.properties) { continue; }
      key = f.properties.mmsi || f.id;
      if (key) { current[key] = f; }
    }
    for (i = 0; i < nextFeatures.length; i++) {
      f = nextFeatures[i];
      if (!f || !f.properties) { continue; }
      key = f.properties.mmsi || f.id;
      old = key ? current[key] : null;
      nextT = +(f.properties.timeUtc || 0);
      oldT = old && old.properties ? +(old.properties.timeUtc || 0) : -1;
      out.push(old && nextT <= oldT ? old : f);
    }
    return out;
  }
  function removeVessels() {
    closePopup();
    if (map.getLayer("aisstream-vessels")) { map.removeLayer("aisstream-vessels"); }
    if (map.getSource("aisstream-vessels")) { map.removeSource("aisstream-vessels"); }
  }
  function installVessels() {
    removeVessels();
    if (!vessels || vessels.length === 0) { workFeatures = []; return; }
    addVesselImage();
    var g = buildFC(vessels);
    workFeatures = cloneWorkFeatures(g.features);
    var b = new maplibregl.LngLatBounds();
    vessels.forEach(function (v) { b.extend([v.lng, v.lat]); });
    map.addSource("aisstream-vessels", { type: "geojson", data: g, cluster: false });
    if (!map.hasImage || !map.hasImage("aisstream-vessel")) { addVesselImage(); }
    map.addLayer({
      id: "aisstream-vessels", type: "symbol", source: "aisstream-vessels", minzoom: 0,
      layout: {
        "icon-image": "aisstream-vessel", "icon-size": 0.5, "icon-allow-overlap": true, "icon-ignore-placement": true,
        "icon-rotate": ["get", "heading"], "icon-rotation-alignment": "map"
      }
    });
    try { map.fitBounds(b, { padding: 48, maxZoom: 9 }); } catch (e) {}
    if (!clickBound) {
      clickBound = true;
      map.on("click", "aisstream-vessels", function (e) {
        var feats = map.queryRenderedFeatures(e.point, { layers: ["aisstream-vessels"] });
        if (!feats.length) { return; }
        var p = feats[0].properties || {};
        var cc = feats[0].geometry.coordinates.slice();
        var wrap = document.createElement("div");
        if (p.detailPath) { var a = document.createElement("a"); a.href = p.detailPath; a.className = "link link-primary"; a.textContent = p.title || p.mmsi || "Vessel"; wrap.appendChild(a); }
        else { wrap.textContent = p.title || p.mmsi || "Vessel"; }
        closePopup();
        popupOpen = new maplibregl.Popup({ offset: 20 }).setLngLat(cc).setDOMContent(wrap).addTo(map);
      });
      map.on("mouseenter", "aisstream-vessels", function () { map.getCanvas().style.cursor = "pointer"; });
      map.on("mouseleave", "aisstream-vessels", function () { map.getCanvas().style.cursor = ""; });
    }
  }
  function replaceVessels(vs) {
    vessels = Array.isArray(vs) ? vs : [];
    var g = buildFC(vessels);
    workFeatures = mergeNewerFeatures(cloneWorkFeatures(g.features));
    var src = map.getSource("aisstream-vessels");
    if (src && src.setData) {
      src.setData({ type: "FeatureCollection", features: workFeatures });
      return;
    }
    if (workFeatures.length) { installVessels(); }
  }
  function refreshVessels() {
    if (!dataURL) { return; }
    fetch(dataURL, { headers: { "Accept": "application/json" }, credentials: "same-origin", cache: "no-store" })
      .then(function (r) { if (!r.ok) { throw new Error("HTTP " + r.status); } return r.json(); })
      .then(replaceVessels)
      .catch(function () {});
  }
  map.on("load", installVessels);
  if (refreshMS > 0) { window.setInterval(refreshVessels, refreshMS); }
  function syncStyle() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function () { installVessels(); });
  }
  if (document.body) { new MutationObserver(syncStyle).observe(document.body, { attributes: true, attributeFilter: ["data-theme"] }); }
  window.addEventListener("storage", function(ev) { if (ev.key === "theme") { syncStyle(); } });
})();`
	return Group([]Node{
		Div(ID("aisstream-maplibre-map"), Class("w-full h-[min(80vh,720px)] min-h-80 rounded-box border border-base-300 z-0")),
		Script(Type("application/json"), ID("aisstream-map-vessels-json"), Raw(string(jsonBytes))),
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
			lago.DynamicPage{Name: "seer_aisstream.AppMenu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_aisstream.MapPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("AISstream live map")},
					&components.FieldText{
						Page:    components.Page{Key: "seer_aisstream.MapBlurb"},
						Getter:  getters.Static("Latest AIS messages with valid WGS84 positions. First version subscribes to the world bounding box and stores all message types in typed tables plus a common envelope."),
						Classes: "text-sm text-base-content/80 max-w-3xl",
					},
					&aisStreamMapLibreMount{Page: components.Page{Key: "seer_aisstream.MapLibreMount"}},
				},
			},
		},
	})
}
