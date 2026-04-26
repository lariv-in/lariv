package p_seer_aisstream

import (
	"context"
	"log/slog"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

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

func registerAISStreamMapPages() {
	lago.RegistryPage.Register("seer_aisstream.MapPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_aisstream.MapPageShell"},
		ExtraHead: []components.PageInterface{
			&components.MapDisplayLibreHead{Page: components.Page{Key: "seer_aisstream.MapLibreHead"}},
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
					&components.MapDisplay{
						Page:    components.Page{Key: "seer_aisstream.MapDisplay"},
						DataURL: lago.RoutePath("seer_aisstream.MapDataRoute", nil),
					},
				},
			},
		},
	})
}
