package p_seer_opensky

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerStateFormPages() {
	createName := getters.Static("seer_opensky.StateCreateForm")
	updateName := getters.Static("seer_opensky.StateUpdateForm")
	deleteName := getters.Static("seer_opensky.StateDeleteForm")

	lago.RegistryPage.Register("seer_opensky.StateCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_opensky.AppMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createName,
				ActionURL: lago.RoutePath("seer_opensky.StateCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[OpenSkyState]{
						Getter:   getters.Static(OpenSkyState{}),
						Attr:     getters.FormBubbling(createName),
						Title:    "Create state row",
						Subtitle: "Manual row; poller also inserts (deduped on icao24 + last_contact).",
						Classes:  "@container max-w-4xl",
						ChildrenInput: []components.PageInterface{
							&components.ContainerColumn{
								Page:     components.Page{Key: "seer_opensky.StateCreateFields"},
								Classes:  "gap-2",
								Children: stateFormInputs(),
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_opensky.StateUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_opensky.StateDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      updateName,
				ActionURL: lago.RoutePath("seer_opensky.StateUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("openskyState.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[OpenSkyState]{
						Getter:   getters.Key[OpenSkyState]("openskyState"),
						Attr:     getters.FormBubbling(updateName),
						Title:    "Edit state",
						Classes:  "@container max-w-4xl",
						ChildrenInput: []components.PageInterface{
							&components.ContainerColumn{
								Page:     components.Page{Key: "seer_opensky.StateUpdateFields"},
								Classes:  "gap-2",
								Children: stateFormInputs(),
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 w-full",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
									&components.ButtonModalForm{
										Label:  "Delete",
										Icon:   "trash",
										Name:   deleteName,
										Url:    lago.RoutePath("seer_opensky.StateDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("openskyState.ID"))}),
										FormPostURL: lago.RoutePath("seer_opensky.StateDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("openskyState.ID"))}),
										ModalUID:    "seer-opensky-state-delete",
										Classes:     "btn-outline btn-error btn-sm",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_opensky.StateDeleteFormModal", &components.Modal{
		UID: "seer-opensky-state-delete",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete this state row?",
				Message: "Removes this row from the local database only.",
				Attr:    getters.FormBubbling(deleteName),
			},
		},
	})
}

func stateFormInputs() []components.PageInterface {
	return []components.PageInterface{
		numSec("Core",
			&components.InputText{Label: "ICAO24 (hex)", Name: "Icao24", Required: true, Getter: getters.Key[string]("$in.Icao24"), Classes: "w-full max-w-md"},
			&components.InputNumber[int64]{Label: "Last contact (unix s)", Name: "LastContact", Required: true, Getter: getters.Key[int64]("$in.LastContact"), Classes: "w-full max-w-md"},
			&components.InputNumber[int64]{Label: "Response snapshot time (unix s)", Name: "SnapshotTime", Required: true, Getter: getters.Key[int64]("$in.SnapshotTime"), Classes: "w-full max-w-md"},
		),
		wrap("Callsign, origin, squawk (optional, strings)",
			&components.InputText{Label: "Callsign", Name: "Callsign", Getter: getters.Deref(getters.Key[*string]("$in.Callsign")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "Origin country", Name: "OriginCountry", Getter: getters.Deref(getters.Key[*string]("$in.OriginCountry")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "Squawk", Name: "Squawk", Getter: getters.Deref(getters.Key[*string]("$in.Squawk")), Classes: "w-full max-w-md"},
		),
		wrap("Time & position (optional: empty = unknown; lon/lat stored as PostgreSQL point)",
			&components.InputText{Label: "Time position (unix, optional)", Name: "TimePosition", Getter: int64PtrForm(getters.Key[*int64]("$in.TimePosition")), Classes: "w-full max-w-md"},
			&components.InputNumber[float64]{Label: "Longitude (WGS84, optional)", Name: "Longitude", Getter: getters.Key[float64]("$in.Longitude"), Classes: "w-full max-w-md"},
			&components.InputNumber[float64]{Label: "Latitude (WGS84, optional)", Name: "Latitude", Getter: getters.Key[float64]("$in.Latitude"), Classes: "w-full max-w-md"},
		),
		wrap("Motion / altitude (optional, numbers)",
			&components.InputText{Label: "Baro altitude (m)", Name: "BaroAltitude", Getter: floatPtrForm(getters.Key[*float64]("$in.BaroAltitude")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "Geometric altitude (m)", Name: "GeoAltitude", Getter: floatPtrForm(getters.Key[*float64]("$in.GeoAltitude")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "Velocity (m/s)", Name: "Velocity", Getter: floatPtrForm(getters.Key[*float64]("$in.Velocity")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "True track (deg)", Name: "TrueTrack", Getter: floatPtrForm(getters.Key[*float64]("$in.TrueTrack")), Classes: "w-full max-w-md"},
			&components.InputText{Label: "Vertical rate (m/s)", Name: "VerticalRate", Getter: floatPtrForm(getters.Key[*float64]("$in.VerticalRate")), Classes: "w-full max-w-md"},
		),
		wrap("Flags (optional: empty = unknown, or true / false / 1 / 0)",
			&components.InputText{Label: "On ground", Name: "OnGround", Getter: boolPtrForm(getters.Key[*bool]("$in.OnGround")), Classes: "w-full max-w-sm"},
			&components.InputText{Label: "SPI", Name: "SPI", Getter: boolPtrForm(getters.Key[*bool]("$in.SPI")), Classes: "w-full max-w-sm"},
		),
		wrap("Integers (optional, empty = unknown)",
			&components.InputText{Label: "Position source (0–3+)", Name: "PositionSource", Getter: intPtrForm(getters.Key[*int]("$in.PositionSource")), Classes: "w-full max-w-sm"},
			&components.InputText{Label: "Category", Name: "Category", Getter: intPtrForm(getters.Key[*int]("$in.Category")), Classes: "w-full max-w-sm"},
		),
		wrap("Sensors (JSON int array, e.g. [1,2,3] or [] )",
			&components.InputText{Label: "Sensors (JSON array)", Name: "SensorsText", Getter: getters.Key[string]("$in.SensorsText"), Classes: "w-full max-w-2xl font-mono text-sm"},
		),
	}
}

func numSec(title string, ch ...components.PageInterface) *components.LabelInline {
	return &components.LabelInline{Title: title, Children: ch}
}

func wrap(title string, ch ...components.PageInterface) *components.LabelInline {
	return &components.LabelInline{Title: title, Children: ch}
}
