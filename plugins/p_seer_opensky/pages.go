package p_seer_opensky

import (
	"context"
	"strconv"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerOpenSkyMenuPages()
	registerOpenSkyMapPages()
	registerStateTableAndDetail()
	registerStateFormPages()
}

func registerOpenSkyMenuPages() {
	lago.RegistryPage.Register("seer_opensky.AppMenu", &components.SidebarMenu{
		Title: getters.Static("OpenSky"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("States"),
				Url:   lago.RoutePath("seer_opensky.DefaultRoute", nil),
			},
			&openskyMapMenuLink{Page: components.Page{Key: "seer_opensky.AppMenuMapLink"}},
		},
	})
}

func registerStateTableAndDetail() {
	lago.RegistryPage.Register("seer_opensky.StateTablePage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_opensky.AppMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[OpenSkyState]{
				Page:    components.Page{Key: "seer_opensky.StateTableBody"},
				UID:     "seer-opensky-states",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[OpenSkyState]]("openskyStates"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_opensky.StateCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{Label: "ICAO24", Name: "Icao24", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Icao24")},
					}},
					{Label: "Last contact (unix)", Name: "LastContact", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int64]("$row.LastContact")))},
					}},
					{Label: "Snapshot (unix)", Name: "SnapshotTime", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int64]("$row.SnapshotTime")))},
					}},
					{Label: "Callsign", Name: "Callsign", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Deref(getters.Key[*string]("$row.Callsign"))},
					}},
					{Label: "Country", Name: "OriginCountry", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Deref(getters.Key[*string]("$row.OriginCountry"))},
					}},
					{Label: "On ground", Name: "OnGround", Children: []components.PageInterface{
						&components.FieldText{Getter: boolPtrString(getters.Key[*bool]("$row.OnGround"))},
					}},
					{Label: "Lon", Name: "Lon", Children: []components.PageInterface{
						&components.FieldText{Getter: pointAxisString(getters.Key[lago.PGPoint]("$row.Position"), true)},
					}},
					{Label: "Lat", Name: "Lat", Children: []components.PageInterface{
						&components.FieldText{Getter: pointAxisString(getters.Key[lago.PGPoint]("$row.Position"), false)},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_opensky.StateDetailMenu", &components.SidebarMenu{
		Title: getters.Static("State"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All states"),
			Url:   lago.RoutePath("seer_opensky.StateListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("View"),
				Url: lago.RoutePath("seer_opensky.StateDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("openskyState.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("seer_opensky.StateUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("openskyState.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("seer_opensky.StateDetailPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_opensky.StateDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[OpenSkyState]{
				Getter: getters.Key[OpenSkyState]("openskyState"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page:    components.Page{Key: "seer_opensky.StateDetailBody"},
						Classes: "gap-3 w-full max-w-4xl",
						Children: append(
							[]components.PageInterface{
								&components.ContainerRow{
									Page:    components.Page{Key: "seer_opensky.StateDetailHeader"},
									Classes: "flex flex-wrap justify-between gap-2 w-full",
									Children: []components.PageInterface{
										&components.FieldTitle{Getter: getters.Format("Aircraft %s", getters.Any(getters.Key[string]("$in.Icao24")))},
										&components.ButtonLink{
											Label:   "Edit",
											Link:    lago.RoutePath("seer_opensky.StateUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
											Classes: "btn-primary btn-sm",
										},
									},
								},
							},
							stateDetailRows()...,
						),
					},
				},
			},
		},
	})
}

func boolPtrString(g getters.Getter[*bool]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		if *p {
			return "yes", nil
		}
		return "no", nil
	}
}

func stateDetailRows() []components.PageInterface {
	return []components.PageInterface{
		lbl("ID", getters.Format("%d", getters.Any(getters.Key[uint]("$in.ID")))),
		lbl("Snapshot (unix)", getters.Format("%d", getters.Any(getters.Key[int64]("$in.SnapshotTime")))),
		lbl("ICAO24", getters.Key[string]("$in.Icao24")),
		lbl("Last contact (unix)", getters.Format("%d", getters.Any(getters.Key[int64]("$in.LastContact")))),
		lbl("Callsign", getters.Deref(getters.Key[*string]("$in.Callsign"))),
		lbl("Origin country", getters.Deref(getters.Key[*string]("$in.OriginCountry"))),
		lbl("Time position (unix)", int64PtrString(getters.Key[*int64]("$in.TimePosition"))),
		lbl("Longitude (WGS84)", pointAxisString(getters.Key[lago.PGPoint]("$in.Position"), true)),
		lbl("Latitude (WGS84)", pointAxisString(getters.Key[lago.PGPoint]("$in.Position"), false)),
		lbl("Baro altitude (m)", floatPtrString(getters.Key[*float64]("$in.BaroAltitude"))),
		lbl("On ground", boolPtrString(getters.Key[*bool]("$in.OnGround"))),
		lbl("Velocity (m/s)", floatPtrString(getters.Key[*float64]("$in.Velocity"))),
		lbl("True track (°)", floatPtrString(getters.Key[*float64]("$in.TrueTrack"))),
		lbl("Vertical rate (m/s)", floatPtrString(getters.Key[*float64]("$in.VerticalRate"))),
		lbl("Sensors (JSON)", getters.Key[string]("$in.SensorsText")),
		lbl("Geometric altitude (m)", floatPtrString(getters.Key[*float64]("$in.GeoAltitude"))),
		lbl("Squawk", getters.Deref(getters.Key[*string]("$in.Squawk"))),
		lbl("SPI", boolPtrString(getters.Key[*bool]("$in.SPI"))),
		lbl("Position source", intPtrString(getters.Key[*int]("$in.PositionSource"))),
		lbl("Category", intPtrString(getters.Key[*int]("$in.Category"))),
		lbl("Created (UTC)", timeString(getters.Key[time.Time]("$in.CreatedAt"))),
		lbl("Updated (UTC)", timeString(getters.Key[time.Time]("$in.UpdatedAt"))),
	}
}

func lbl(title string, g getters.Getter[string]) *components.LabelInline {
	return &components.LabelInline{Title: title, Children: []components.PageInterface{
		&components.FieldText{Getter: g, Classes: "min-w-0 break-all"},
	}}
}

func int64PtrString(g getters.Getter[*int64]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.FormatInt(*p, 10), nil
	}
}

func intPtrString(g getters.Getter[*int]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.Itoa(*p), nil
	}
}

func pointAxisString(g getters.Getter[lago.PGPoint], longitude bool) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if !p.Valid {
			return "", nil
		}
		v := p.P.Y
		if longitude {
			v = p.P.X
		}
		return strconv.FormatFloat(v, 'f', 6, 64), nil
	}
}

func floatPtrString(g getters.Getter[*float64]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		p, err := g(ctx)
		if err != nil {
			return "", err
		}
		if p == nil {
			return "", nil
		}
		return strconv.FormatFloat(*p, 'f', 5, 64), nil
	}
}

func timeString(g getters.Getter[time.Time]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		t, err := g(ctx)
		if err != nil {
			return "", err
		}
		if t.IsZero() {
			return "", nil
		}
		return t.UTC().Format(time.RFC3339), nil
	}
}
