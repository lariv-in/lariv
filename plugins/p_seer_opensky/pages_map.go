package p_seer_opensky

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

func registerOpenSkyMapPages() {
	lago.RegistryPage.Register("seer_opensky.MapPage", &components.ShellScaffold{
		Page: components.Page{Key: "seer_opensky.MapPageShell"},
		ExtraHead: []components.PageInterface{
			&components.MapDisplayLibreHead{Page: components.Page{Key: "seer_opensky.MapLibreHead"}},
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
					&components.MapDisplay{
						Page:    components.Page{Key: "seer_opensky.MapDisplay"},
						DataURL: lago.RoutePath("seer_opensky.MapDataRoute", nil),
					},
				},
			},
		},
	})
}
