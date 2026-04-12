package p_lacerate

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	html "maragu.dev/gomponents/html"
)

type leafletHeadAssets struct {
	components.Page
}

func (e leafletHeadAssets) GetKey() string {
	return e.Key
}

func (e leafletHeadAssets) GetRoles() []string {
	return e.Roles
}

func (e leafletHeadAssets) Build(context.Context) Node {
	return Group{
		html.Link(
			html.Rel("stylesheet"),
			html.Href("https://unpkg.com/leaflet@1.9.4/dist/leaflet.css"),
			html.Integrity("sha256-p4NxAoJBhIIN+hmNHrzRCf9tD/miZyoHS5obTRR9BMY="),
			html.CrossOrigin(""),
		),
		html.Script(
			html.Src("https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"),
			html.Integrity("sha256-20nQCchB9co0qIjJZRGuk2/Z9VM+kNiyxNV1lvTlZBo="),
			html.CrossOrigin(""),
		),
		html.StyleEl(Text(`
.lacerate-map-popup p {
	margin: 0.25rem 0;
}
.leaflet-control-layers {
	font-size: 0.875rem;
}
`)),
	}
}

type lacerateMapPage struct {
	components.Page
}

func (e lacerateMapPage) GetKey() string {
	return e.Key
}

func (e lacerateMapPage) GetRoles() []string {
	return e.Roles
}

func (e lacerateMapPage) Build(ctx context.Context) Node {
	data, _ := ctx.Value(ctxKeyLacerateMapData).(lacerateMapData)

	mapPayload := struct {
		Layers []mapLayerGroup `json:"layers"`
	}{Layers: data.Layers}
	mapJSON, err := json.Marshal(mapPayload)
	if err != nil {
		slog.Error("lacerate: map: marshal map payload", "error", err)
		mapJSON = []byte(`{"layers":[]}`)
	}

	totalMarkers := 0
	for _, lg := range data.Layers {
		totalMarkers += len(lg.Markers)
	}

	content := Group{
		html.Div(
			html.Class("space-y-2"),
			html.H1(html.Class("text-3xl font-semibold"), Text("Map")),
			html.P(
				html.Class("text-sm text-base-content/70"),
				Text("Events are grouped by nearest target of interest (intel vs TOI embedding similarity); weak matches go under Uncategorized."),
			),
		),
	}

	if data.UnsupportedMessage != "" {
		content = append(content,
			html.Div(
				html.Class("alert"),
				Text(data.UnsupportedMessage),
			),
		)
		return html.Div(html.Class("space-y-6"), content)
	}

	content = append(content,
		html.Div(
			html.Class("stats shadow w-full"),
			html.Div(
				html.Class("stat"),
				html.Div(html.Class("stat-title"), Text("Events")),
				html.Div(html.Class("stat-value text-2xl"), Text(fmt.Sprintf("%d", totalMarkers))),
			),
		),
	)

	if totalMarkers == 0 {
		content = append(content,
			html.Div(
				html.Class("alert"),
				Text("No plotted events yet."),
			),
		)
		return html.Div(html.Class("space-y-6"), content)
	}

	content = append(content,
		html.Div(
			html.Class("card border border-base-300 bg-base-100 shadow-sm"),
			html.Div(
				html.Class("card-body gap-4"),
				html.P(
					html.Class("text-sm text-base-content/70"),
					Text("Use the layer control to show or hide groups. Click a marker for details, target link (when categorized), and intel link."),
				),
				html.Div(
					html.Class("rounded-box border border-base-300 overflow-hidden"),
					Attr("x-data", lacerateMapXData(string(mapJSON))),
					Attr("x-init", "init()"),
					html.Div(
						html.Class("w-full"),
						Attr("x-ref", "map"),
						html.Style("height: 70vh; min-height: 32rem;"),
					),
				),
			),
		),
	)

	return html.Div(html.Class("space-y-6"), content)
}

func registerMapPages() {
	lago.RegistryPage.Patch("lacerate.LacerateMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Map"),
			Url:   lago.RoutePath("lacerate.MapRoute", nil),
		})
		return menu
	})

	lago.RegistryPage.Register("lacerate.MapPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "lacerate.LacerateMenu"},
		},
		ExtraHead: []components.PageInterface{
			&leafletHeadAssets{Page: components.Page{Key: "lacerate.MapHeadAssets"}},
		},
		Children: []components.PageInterface{
			&lacerateMapPage{Page: components.Page{Key: "lacerate.MapContent"}},
		},
	})
}

func lacerateMapXData(mapPayloadJSON string) string {
	return fmt.Sprintf(`{
		mapPayload: %s,
		map: null,
		init() {
			if (!window.L || !this.$refs.map) return;
			this.map = L.map(this.$refs.map);
			L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
				maxZoom: 19,
				attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
			}).addTo(this.map);
			const bounds = [];
			const overlays = {};
			const layers = this.mapPayload.layers || [];
			for (const layerDef of layers) {
				const group = L.layerGroup();
				for (const marker of (layerDef.markers || [])) {
					if (!Number.isFinite(marker.lat) || !Number.isFinite(marker.lng)) continue;
					bounds.push([marker.lat, marker.lng]);
					const popup = [];
					popup.push('<div class="lacerate-map-popup">');
					const head = marker.title || ('Event #' + marker.eventId);
					popup.push('<p><strong>' + this.escapeHtml(head) + '</strong></p>');
					if (marker.datetime) popup.push('<p>' + this.escapeHtml(marker.datetime) + '</p>');
					if (marker.address) popup.push('<p>' + this.escapeHtml(marker.address) + '</p>');
					if (marker.targetOfInterestUrl && marker.targetOfInterestName) {
						popup.push('<p><a class="link link-secondary" href="' + this.escapeHtml(marker.targetOfInterestUrl) + '">Target: ' + this.escapeHtml(marker.targetOfInterestName) + '</a></p>');
					}
					if (marker.intelPreview) popup.push('<p>' + this.escapeHtml(marker.intelPreview) + '</p>');
					if (marker.intelUrl) popup.push('<p><a class="link link-primary" href="' + this.escapeHtml(marker.intelUrl) + '">Open intel</a></p>');
					popup.push('</div>');
					group.addLayer(L.marker([marker.lat, marker.lng]).bindPopup(popup.join('')));
				}
				group.addTo(this.map);
				overlays[layerDef.label] = group;
			}
			if (Object.keys(overlays).length > 0) {
				L.control.layers(null, overlays, { collapsed: false }).addTo(this.map);
			}
			if (bounds.length === 0) {
				this.map.setView([20, 0], 2);
				this.map.invalidateSize();
				return;
			}
			if (bounds.length === 1) {
				this.map.setView(bounds[0], 12);
				this.map.invalidateSize();
				return;
			}
			this.map.fitBounds(bounds, {padding: [24, 24]});
			this.map.invalidateSize();
		},
		escapeHtml(value) {
			return String(value || "").replace(/[&<>"']/g, (ch) => ({
				"&": "&amp;",
				"<": "&lt;",
				">": "&gt;",
				'"': "&quot;",
				"'": "&#39;"
			}[ch]));
		}
	}`, mapPayloadJSON)
}
