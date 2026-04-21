package p_seer_gdelt

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MapLibre GL JS (pinned) — CSS + JS for map page only via ShellScaffold ExtraHead.
const mapLibreCDNVersion = "4.7.1"

type gdeltMapLibreExtraHead struct {
	components.Page
}

func (e *gdeltMapLibreExtraHead) GetKey() string     { return e.Key }
func (e *gdeltMapLibreExtraHead) GetRoles() []string { return e.Roles }

func (e *gdeltMapLibreExtraHead) Build(ctx context.Context) Node {
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

// Sidebar entry: full navigation so inline map script runs (hx-boost disabled on this link).
type gdeltMapSidebarLink struct {
	components.Page
}

func (e *gdeltMapSidebarLink) GetKey() string     { return e.Key }
func (e *gdeltMapSidebarLink) GetRoles() []string { return e.Roles }

func (e *gdeltMapSidebarLink) Build(ctx context.Context) Node {
	href, err := lago.RoutePath("seer_gdelt.MapRoute", nil)(ctx)
	if err != nil || href == "" {
		slog.Error("p_seer_gdelt: MapRoute path failed", "error", err)
		href = AppUrl + "map/"
	}
	return Li(
		A(Href(href), Attr("data-hx-boost", "false"),
			components.Render(components.Icon{Name: "map-pin", Classes: "heroicon-sm"}, ctx),
			Text(" Map"),
		),
	)
}

type gdeltMapLibreMount struct {
	components.Page
}

func (e *gdeltMapLibreMount) GetKey() string     { return e.Key }
func (e *gdeltMapLibreMount) GetRoles() []string { return e.Roles }

func (e *gdeltMapLibreMount) Build(ctx context.Context) Node {
	var markers []gdeltMapMarker
	if v := ctx.Value(gdeltMapMarkersKey); v != nil {
		if m, ok := v.([]gdeltMapMarker); ok {
			markers = m
		}
	}
	jsonBytes, err := json.Marshal(markers)
	if err != nil {
		jsonBytes = []byte("[]")
	}
	// Light: official MapLibre demotiles (maplibre.org only publishes this global style; no dark.json there).
	// Dark: OpenFreeMap dark style (MapLibre Style JSON, no API key; see madewithmaplibre.com basemaps). Theme: [components.ShellBase] :data-theme / localStorage "theme".
	initJS := `(function(){
  var je = document.getElementById("gdelt-map-markers-json");
  var markers = [];
  try { markers = JSON.parse(je ? je.textContent || "[]" : "[]"); } catch (e0) {}
  var mapEl = document.getElementById("gdelt-maplibre-map");
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
  var map = new maplibregl.Map({
    container: mapEl,
    style: lastDark ? styleDark : styleLight,
    center: [0, 20],
    zoom: 1.5
  });
  map.addControl(new maplibregl.NavigationControl(), "top-right");
  var popupOpen = null;
  function closePopup() {
    if (popupOpen) {
      popupOpen.remove();
      popupOpen = null;
    }
  }
  function buildGeoJSONFeatureCollection(ms) {
    return {
      type: "FeatureCollection",
      features: ms.map(function (m) {
        return {
          type: "Feature",
          geometry: { type: "Point", coordinates: [m.lng, m.lat] },
          properties: {
            kind: m.kind,
            title: m.title,
            detailPath: m.detailPath,
            eventId: m.eventId
          }
        };
      })
    };
  }
  function clusterRadiusForDisplay() {
    var dpr = window.devicePixelRatio || 1;
    return Math.round(36 * Math.min(1.85, Math.sqrt(dpr)));
  }
  function removeGDeltaMarkerLayers() {
    ["gdelt-unclustered", "gdelt-clusters"].forEach(function (id) {
      if (map.getLayer(id)) {
        map.removeLayer(id);
      }
    });
    if (map.getSource("gdelt-markers")) {
      map.removeSource("gdelt-markers");
    }
  }
  var clusterClickBound = false;
  var didFitMarkers = false;
  function installMarkers() {
    closePopup();
    removeGDeltaMarkerLayers();
    if (markers.length === 0) {
      return;
    }
    var b = new maplibregl.LngLatBounds();
    markers.forEach(function (m) {
      b.extend([m.lng, m.lat]);
    });
    var geojson = buildGeoJSONFeatureCollection(markers);
    var clusterRadius = clusterRadiusForDisplay();
    var clusterMaxZoom = 14;
    map.addSource("gdelt-markers", {
      type: "geojson",
      data: geojson,
      cluster: true,
      clusterMaxZoom: clusterMaxZoom,
      clusterRadius: clusterRadius,
      clusterMinPoints: 2
    });
    map.addLayer({
      id: "gdelt-clusters",
      type: "circle",
      source: "gdelt-markers",
      filter: ["has", "point_count"],
      paint: {
        "circle-color": "#818cf8",
        "circle-radius": [
          "step",
          ["get", "point_count"],
          16,
          10,
          20,
          50,
          24,
          200,
          30
        ],
        "circle-opacity": 0.92,
        "circle-stroke-width": 2,
        "circle-stroke-color": "#e0e7ff"
      }
    });
    map.addLayer({
      id: "gdelt-unclustered",
      type: "circle",
      source: "gdelt-markers",
      filter: ["!", ["has", "point_count"]],
      paint: {
        "circle-color": [
          "match",
          ["get", "kind"],
          "actor1",
          "#60a5fa",
          "actor2",
          "#4ade80",
          "action",
          "#f87171",
          "#a3a3a3"
        ],
        "circle-radius": 10,
        "circle-stroke-width": 2,
        "circle-stroke-color": "#ffffff"
      }
    });
    if (!clusterClickBound) {
      clusterClickBound = true;
      map.on("click", "gdelt-clusters", function (e) {
        closePopup();
        var feats = map.queryRenderedFeatures(e.point, { layers: ["gdelt-clusters"] });
        if (!feats.length) {
          return;
        }
        var src = map.getSource("gdelt-markers");
        if (!src || typeof src.getClusterLeaves !== "function") {
          return;
        }
        var clusterFeat = feats[0];
        var cid = +clusterFeat.properties.cluster_id;
        var n = +clusterFeat.properties.point_count || 0;
        var center = clusterFeat.geometry.coordinates.slice();
        var limit = Math.max(n, 1);
        var leavesPromise = src.getClusterLeaves(cid, limit, 0);
        function showLeaves(leaves) {
          if (!leaves || !leaves.length) {
            return;
          }
          leaves = leaves.slice().sort(function (a, b) {
            var ta = (a.properties && a.properties.title) || "";
            var tb = (b.properties && b.properties.title) || "";
            return ta.localeCompare(tb);
          });
          var wrap = document.createElement("div");
          wrap.className = "flex flex-col gap-1 min-w-[14rem] max-w-sm max-h-72 overflow-y-auto py-1";
          var head = document.createElement("div");
          head.className = "text-sm font-semibold opacity-90 mb-1 sticky top-0 bg-base-100 pb-1 z-10";
          head.textContent = leaves.length + " locations";
          wrap.appendChild(head);
          leaves.forEach(function (leaf) {
            var p = leaf.properties || {};
            var row = document.createElement("div");
            var a = document.createElement("a");
            a.href = p.detailPath || "#";
            a.textContent = p.title || "Event";
            a.className = "link link-primary text-sm block truncate";
            row.appendChild(a);
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
      });
      map.on("click", "gdelt-unclustered", function (e) {
        closePopup();
        var feats = map.queryRenderedFeatures(e.point, { layers: ["gdelt-unclustered"] });
        if (!feats.length) {
          return;
        }
        var f = feats[0];
        var coords = f.geometry.coordinates.slice();
        var p = f.properties;
        var wrap = document.createElement("div");
        var a = document.createElement("a");
        a.href = p.detailPath;
        a.textContent = p.title;
        a.className = "link link-primary";
        wrap.appendChild(a);
        popupOpen = new maplibregl.Popup({ offset: 12, closeOnClick: true })
          .setLngLat(coords)
          .setDOMContent(wrap)
          .addTo(map);
      });
      map.on("mouseenter", "gdelt-clusters", function () {
        map.getCanvas().style.cursor = "pointer";
      });
      map.on("mouseleave", "gdelt-clusters", function () {
        map.getCanvas().style.cursor = "";
      });
      map.on("mouseenter", "gdelt-unclustered", function () {
        map.getCanvas().style.cursor = "pointer";
      });
      map.on("mouseleave", "gdelt-unclustered", function () {
        map.getCanvas().style.cursor = "";
      });
    }
    if (!didFitMarkers) {
      didFitMarkers = true;
      try {
        map.fitBounds(b, { padding: 60, maxZoom: 12 });
      } catch (e1) {}
    }
  }
  map.on("load", function () {
    installMarkers();
  });
  function syncStyleToTheme() {
    var d = themeIsDark();
    if (d === lastDark) { return; }
    lastDark = d;
    map.setStyle(d ? styleDark : styleLight);
    map.once("idle", function () {
      installMarkers();
    });
  }
  if (document.body) {
    new MutationObserver(syncStyleToTheme).observe(document.body, { attributes: true, attributeFilter: ["data-theme"] });
  }
  window.addEventListener("storage", function(ev) {
    if (ev.key !== "theme") { return; }
    syncStyleToTheme();
  });
})();`
	return Group([]Node{
		Div(
			ID("gdelt-maplibre-map"),
			Class("w-full h-[70vh] min-h-96 rounded-box border border-base-300 z-0"),
		),
		Script(Type("application/json"), ID("gdelt-map-markers-json"), Raw(string(jsonBytes))),
		Script(Raw(initJS)),
	})
}

func registerGDELTMapPages() {
	lago.RegistryPage.Register("seer_gdelt.MapPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_gdelt.MapPageShell"},
		ExtraHead: []components.PageInterface{
			&gdeltMapLibreExtraHead{Page: components.Page{Key: "seer_gdelt.MapLibreHead"}},
		},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_gdelt.MapPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("GDELT events map")},
					&components.FieldText{
						Page:    components.Page{Key: "seer_gdelt.MapBlurb"},
						Getter:  getters.Static("Actor1 / Actor2 / Action locations from stored events (up to 5000 most recent rows). Points cluster when zoomed out (radius scales with screen pixel density); click a cluster for a scrollable list of locations, or a single point for one event link."),
						Classes: "text-sm text-base-content/80 max-w-3xl",
					},
					&gdeltMapLibreMount{Page: components.Page{Key: "seer_gdelt.MapLibreMount"}},
				},
			},
		},
	})
}
