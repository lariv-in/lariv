package p_seer_gdelt

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

// Anti-pattern (allowed only in this file): hand-built gdeltForm* wrappers plus long section lists
// that duplicate [Event] field-by-field. Elsewhere prefer shared list/form builders, registries, or
// narrower forms so UI does not fork from the model. Here it is intentional: one full local mirror
// of the external GDELT events schema, explicit wiring, grep-friendly—do not copy this shape for
// normal app models.
//
// Form inputs use $in.*; order matches models.Event.

func gdeltFormStr(name, label string) components.PageInterface {
	key := "$in." + name
	return &components.ContainerError{
		Children: []components.PageInterface{
			&components.InputText{
				Label:   label,
				Name:    name,
				Getter:  getters.Key[string](key),
				Classes: "w-full max-w-4xl",
			},
		},
	}
}

func gdeltFormInt(name, label string) components.PageInterface {
	key := "$in." + name
	return &components.ContainerError{
		Children: []components.PageInterface{
			&components.InputNumber[int]{
				Label:   label,
				Name:    name,
				Getter:  getters.Key[int](key),
				Classes: "w-full max-w-md",
			},
		},
	}
}

func gdeltFormInt64(name, label string) components.PageInterface {
	key := "$in." + name
	return &components.ContainerError{
		Children: []components.PageInterface{
			&components.InputNumber[int64]{
				Label:   label,
				Name:    name,
				Getter:  getters.Key[int64](key),
				Classes: "w-full max-w-md",
			},
		},
	}
}

func gdeltFormUint64(name, label string, required bool) components.PageInterface {
	key := "$in." + name
	return &components.ContainerError{
		Children: []components.PageInterface{
			&components.InputNumber[uint64]{
				Label:    label,
				Name:     name,
				Required: required,
				Getter:   getters.Key[uint64](key),
				Classes:  "w-full max-w-md",
			},
		},
	}
}

func gdeltFormFloat64(name, label string) components.PageInterface {
	key := "$in." + name
	return &components.ContainerError{
		Children: []components.PageInterface{
			&components.InputNumber[float64]{
				Label:   label,
				Name:    name,
				Getter:  getters.Key[float64](key),
				Classes: "w-full max-w-md",
			},
		},
	}
}

func gdeltSection(title string, fields ...components.PageInterface) components.PageInterface {
	return &components.LabelInline{Title: title, Children: fields}
}

func gdeltEventFormInputs() []components.PageInterface {
	return []components.PageInterface{
		gdeltSection("Identifiers",
			gdeltFormUint64("GlobalEventID", "Global event ID", true),
			gdeltFormInt("SQLDate", "SQL date (YYYYMMDD)"),
			gdeltFormStr("MonthYear", "Month-year"),
			gdeltFormStr("Year", "Year"),
			gdeltFormFloat64("FractionDate", "Fraction date"),
		),
		gdeltSection("Actor 1",
			gdeltFormStr("Actor1Code", "Actor1 code"),
			gdeltFormStr("Actor1Name", "Actor1 name"),
			gdeltFormStr("Actor1CountryCode", "Actor1 country code"),
			gdeltFormStr("Actor1KnownGroupCode", "Actor1 known group code"),
			gdeltFormStr("Actor1EthnicCode", "Actor1 ethnic code"),
			gdeltFormStr("Actor1Religion1Code", "Actor1 religion1 code"),
			gdeltFormStr("Actor1Religion2Code", "Actor1 religion2 code"),
			gdeltFormStr("Actor1Type1Code", "Actor1 type1 code"),
			gdeltFormStr("Actor1Type2Code", "Actor1 type2 code"),
			gdeltFormStr("Actor1Type3Code", "Actor1 type3 code"),
		),
		gdeltSection("Actor 2",
			gdeltFormStr("Actor2Code", "Actor2 code"),
			gdeltFormStr("Actor2Name", "Actor2 name"),
			gdeltFormStr("Actor2CountryCode", "Actor2 country code"),
			gdeltFormStr("Actor2KnownGroupCode", "Actor2 known group code"),
			gdeltFormStr("Actor2EthnicCode", "Actor2 ethnic code"),
			gdeltFormStr("Actor2Religion1Code", "Actor2 religion1 code"),
			gdeltFormStr("Actor2Religion2Code", "Actor2 religion2 code"),
			gdeltFormStr("Actor2Type1Code", "Actor2 type1 code"),
			gdeltFormStr("Actor2Type2Code", "Actor2 type2 code"),
			gdeltFormStr("Actor2Type3Code", "Actor2 type3 code"),
		),
		gdeltSection("Event",
			gdeltFormInt("IsRootEvent", "Is root event"),
			gdeltFormStr("EventCode", "Event code"),
			gdeltFormStr("EventBaseCode", "Event base code"),
			gdeltFormStr("EventRootCode", "Event root code"),
			gdeltFormInt("QuadClass", "Quad class"),
			gdeltFormFloat64("GoldsteinScale", "Goldstein scale"),
			gdeltFormInt("NumMentions", "Num mentions"),
			gdeltFormInt("NumSources", "Num sources"),
			gdeltFormInt("NumArticles", "Num articles"),
			gdeltFormFloat64("AvgTone", "Avg tone"),
		),
		gdeltSection("Actor 1 geo",
			gdeltFormInt("Actor1GeoType", "Actor1 geo type"),
			gdeltFormStr("Actor1GeoFullName", "Actor1 geo full name"),
			gdeltFormStr("Actor1GeoCountryCode", "Actor1 geo country code"),
			gdeltFormStr("Actor1GeoADM1Code", "Actor1 geo ADM1"),
			gdeltFormStr("Actor1GeoADM2Code", "Actor1 geo ADM2"),
			gdeltFormFloat64("Actor1GeoLat", "Actor1 geo lat"),
			gdeltFormFloat64("Actor1GeoLong", "Actor1 geo long"),
			gdeltFormStr("Actor1GeoFeatureID", "Actor1 geo feature ID"),
		),
		gdeltSection("Actor 2 geo",
			gdeltFormInt("Actor2GeoType", "Actor2 geo type"),
			gdeltFormStr("Actor2GeoFullName", "Actor2 geo full name"),
			gdeltFormStr("Actor2GeoCountryCode", "Actor2 geo country code"),
			gdeltFormStr("Actor2GeoADM1Code", "Actor2 geo ADM1"),
			gdeltFormStr("Actor2GeoADM2Code", "Actor2 geo ADM2"),
			gdeltFormFloat64("Actor2GeoLat", "Actor2 geo lat"),
			gdeltFormFloat64("Actor2GeoLong", "Actor2 geo long"),
			gdeltFormStr("Actor2GeoFeatureID", "Actor2 geo feature ID"),
		),
		gdeltSection("Action geo",
			gdeltFormInt("ActionGeoType", "Action geo type"),
			gdeltFormStr("ActionGeoFullName", "Action geo full name"),
			gdeltFormStr("ActionGeoCountryCode", "Action geo country code"),
			gdeltFormStr("ActionGeoADM1Code", "Action geo ADM1"),
			gdeltFormStr("ActionGeoADM2Code", "Action geo ADM2"),
			gdeltFormFloat64("ActionGeoLat", "Action geo lat"),
			gdeltFormFloat64("ActionGeoLong", "Action geo long"),
			gdeltFormStr("ActionGeoFeatureID", "Action geo feature ID"),
		),
		gdeltSection("Other",
			gdeltFormInt64("DateAdded", "Date added (epoch seconds)"),
			gdeltFormStr("SourceURL", "Source URL"),
		),
	}
}

// --- Read-only detail: LabelNewline = label above value (column). Sections = Accordion panels. ---

func gdeltDetailStr(label, name string) components.PageInterface {
	key := "$in." + name
	return &components.LabelNewline{
		Title: label,
		Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string](key), Classes: "whitespace-pre-wrap break-all"},
		},
	}
}

func gdeltDetailScalar(label, name string) components.PageInterface {
	key := "$in." + name
	return &components.LabelNewline{
		Title: label,
		Children: []components.PageInterface{
			&components.FieldText{Getter: gdeltFmtAny(key), Classes: "font-mono"},
		},
	}
}

func gdeltDetailAccordionPanel(title string, open bool, fields ...components.PageInterface) components.AccordionItem {
	sectionKey := strings.ReplaceAll(strings.ReplaceAll(title, " ", "_"), "/", "_")
	return components.AccordionItem{
		Open: open,
		Title: components.FieldText{
			Classes: "font-semibold text-base",
			Getter:  getters.Static(title),
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:     components.Page{Key: "seer_gdelt.EventDetailSection." + sectionKey},
				Classes:  "flex flex-col gap-3 w-full max-w-full pt-1",
				Children: fields,
			},
		},
	}
}

func gdeltEventDetailRows() []components.PageInterface {
	return []components.PageInterface{
		&components.Accordion{
			Page:    components.Page{Key: "seer_gdelt.EventDetailAccordion"},
			Classes: "w-full max-w-5xl",
			Items: []components.AccordionItem{
				gdeltDetailAccordionPanel("Record", true,
					evDetailUint("Row ID", "$in.ID"),
					evDetailTime("Created", "$in.CreatedAt"),
					evDetailTime("Updated", "$in.UpdatedAt"),
				),
				gdeltDetailAccordionPanel("Identifiers", false,
					gdeltDetailScalar("Global event ID", "GlobalEventID"),
					gdeltDetailScalar("SQL date (YYYYMMDD)", "SQLDate"),
					gdeltDetailStr("Month-year", "MonthYear"),
					gdeltDetailStr("Year", "Year"),
					gdeltDetailScalar("Fraction date", "FractionDate"),
				),
				gdeltDetailAccordionPanel("Actor 1", false,
					gdeltDetailStr("Actor1 code", "Actor1Code"),
					gdeltDetailStr("Actor1 name", "Actor1Name"),
					gdeltDetailStr("Actor1 country code", "Actor1CountryCode"),
					gdeltDetailStr("Actor1 known group code", "Actor1KnownGroupCode"),
					gdeltDetailStr("Actor1 ethnic code", "Actor1EthnicCode"),
					gdeltDetailStr("Actor1 religion1 code", "Actor1Religion1Code"),
					gdeltDetailStr("Actor1 religion2 code", "Actor1Religion2Code"),
					gdeltDetailStr("Actor1 type1 code", "Actor1Type1Code"),
					gdeltDetailStr("Actor1 type2 code", "Actor1Type2Code"),
					gdeltDetailStr("Actor1 type3 code", "Actor1Type3Code"),
				),
				gdeltDetailAccordionPanel("Actor 2", false,
					gdeltDetailStr("Actor2 code", "Actor2Code"),
					gdeltDetailStr("Actor2 name", "Actor2Name"),
					gdeltDetailStr("Actor2 country code", "Actor2CountryCode"),
					gdeltDetailStr("Actor2 known group code", "Actor2KnownGroupCode"),
					gdeltDetailStr("Actor2 ethnic code", "Actor2EthnicCode"),
					gdeltDetailStr("Actor2 religion1 code", "Actor2Religion1Code"),
					gdeltDetailStr("Actor2 religion2 code", "Actor2Religion2Code"),
					gdeltDetailStr("Actor2 type1 code", "Actor2Type1Code"),
					gdeltDetailStr("Actor2 type2 code", "Actor2Type2Code"),
					gdeltDetailStr("Actor2 type3 code", "Actor2Type3Code"),
				),
				gdeltDetailAccordionPanel("Event", false,
					gdeltDetailScalar("Is root event", "IsRootEvent"),
					gdeltDetailStr("Event code", "EventCode"),
					gdeltDetailStr("Event base code", "EventBaseCode"),
					gdeltDetailStr("Event root code", "EventRootCode"),
					gdeltDetailScalar("Quad class", "QuadClass"),
					gdeltDetailScalar("Goldstein scale", "GoldsteinScale"),
					gdeltDetailScalar("Num mentions", "NumMentions"),
					gdeltDetailScalar("Num sources", "NumSources"),
					gdeltDetailScalar("Num articles", "NumArticles"),
					gdeltDetailScalar("Avg tone", "AvgTone"),
				),
				gdeltDetailAccordionPanel("Actor 1 geo", false,
					gdeltDetailScalar("Actor1 geo type", "Actor1GeoType"),
					gdeltDetailStr("Actor1 geo full name", "Actor1GeoFullName"),
					gdeltDetailStr("Actor1 geo country code", "Actor1GeoCountryCode"),
					gdeltDetailStr("Actor1 geo ADM1", "Actor1GeoADM1Code"),
					gdeltDetailStr("Actor1 geo ADM2", "Actor1GeoADM2Code"),
					gdeltDetailScalar("Actor1 geo lat", "Actor1GeoLat"),
					gdeltDetailScalar("Actor1 geo long", "Actor1GeoLong"),
					gdeltDetailStr("Actor1 geo feature ID", "Actor1GeoFeatureID"),
				),
				gdeltDetailAccordionPanel("Actor 2 geo", false,
					gdeltDetailScalar("Actor2 geo type", "Actor2GeoType"),
					gdeltDetailStr("Actor2 geo full name", "Actor2GeoFullName"),
					gdeltDetailStr("Actor2 geo country code", "Actor2GeoCountryCode"),
					gdeltDetailStr("Actor2 geo ADM1", "Actor2GeoADM1Code"),
					gdeltDetailStr("Actor2 geo ADM2", "Actor2GeoADM2Code"),
					gdeltDetailScalar("Actor2 geo lat", "Actor2GeoLat"),
					gdeltDetailScalar("Actor2 geo long", "Actor2GeoLong"),
					gdeltDetailStr("Actor2 geo feature ID", "Actor2GeoFeatureID"),
				),
				gdeltDetailAccordionPanel("Action geo", false,
					gdeltDetailScalar("Action geo type", "ActionGeoType"),
					gdeltDetailStr("Action geo full name", "ActionGeoFullName"),
					gdeltDetailStr("Action geo country code", "ActionGeoCountryCode"),
					gdeltDetailStr("Action geo ADM1", "ActionGeoADM1Code"),
					gdeltDetailStr("Action geo ADM2", "ActionGeoADM2Code"),
					gdeltDetailScalar("Action geo lat", "ActionGeoLat"),
					gdeltDetailScalar("Action geo long", "ActionGeoLong"),
					gdeltDetailStr("Action geo feature ID", "ActionGeoFeatureID"),
				),
				gdeltDetailAccordionPanel("Other", false,
					gdeltDetailScalar("Date added (epoch seconds)", "DateAdded"),
					gdeltDetailStr("Source URL", "SourceURL"),
				),
			},
		},
	}
}

func evDetailUint(label, path string) components.PageInterface {
	return &components.LabelNewline{
		Title: label,
		Children: []components.PageInterface{
			&components.FieldText{
				Getter: getters.Map(getters.Key[uint](path), func(_ context.Context, v uint) (string, error) {
					return strconv.FormatUint(uint64(v), 10), nil
				}),
				Classes: "font-mono",
			},
		},
	}
}

func evDetailTime(label, path string) components.PageInterface {
	return &components.LabelNewline{
		Title: label,
		Children: []components.PageInterface{
			&components.FieldText{
				Getter: getters.Map(getters.Key[time.Time](path), func(_ context.Context, t time.Time) (string, error) {
					if t.IsZero() {
						return "", nil
					}
					return t.UTC().Format(time.RFC3339), nil
				}),
				Classes: "font-mono text-sm",
			},
		},
	}
}

func eventListColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "ID",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Map(getters.Key[uint]("$row.ID"), func(_ context.Context, id uint) (string, error) {
					return strconv.FormatUint(uint64(id), 10), nil
				}), Classes: "font-mono"},
			},
		},
		{
			Label: "Global ID",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Map(getters.Key[uint64]("$row.GlobalEventID"), func(_ context.Context, id uint64) (string, error) {
					return strconv.FormatUint(id, 10), nil
				}), Classes: "font-mono"},
			},
		},
		{
			Label: "Date",
			Children: []components.PageInterface{
				&components.FieldText{Getter: gdeltListSQLDateGetter(), Classes: "font-mono"},
			},
		},
		{
			Label: "Actors",
			Children: []components.PageInterface{
				&components.FieldText{Getter: gdeltListActorsGetter(), Classes: "whitespace-normal max-w-md"},
			},
		},
		{
			Label: "Action country",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.ActionGeoCountryCode")},
			},
		},
		{
			Label: "Mentions",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Map(getters.Key[int]("$row.NumMentions"), func(_ context.Context, n int) (string, error) {
					return strconv.Itoa(n), nil
				})},
			},
		},
	}
}

func registerGDELTEventPages() {
	registerGDELTEventMenus()
	registerGDELTEventListPage()
	registerGDELTEventDetailPage()
	registerGDELTEventFormPages()
}

func registerGDELTEventMenus() {
	lago.RegistryPage.Register("seer_gdelt.EventDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Event #%d", getters.Any(getters.Key[uint]("gdeltEvent.ID"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("All events"),
			Url:   lago.RoutePath("seer_gdelt.EventListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("seer_gdelt.EventUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
			},
		},
	})
}

func registerGDELTEventListPage() {
	lago.RegistryPage.Register("seer_gdelt.EventTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Event]{
				Page:    components.Page{Key: "seer_gdelt.EventTableBody"},
				UID:     "seer-gdelt-events",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Event]]("gdeltEvents"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_gdelt.EventCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_gdelt.EventDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: eventListColumns(),
			},
		},
	})
}

func registerGDELTEventDetailPage() {
	lago.RegistryPage.Register("seer_gdelt.EventDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.EventDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Event]{
				Getter: getters.Key[Event]("gdeltEvent"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page:    components.Page{Key: "seer_gdelt.EventDetailBody"},
						Classes: "gap-3 w-full max-w-5xl",
						Children: append([]components.PageInterface{
							&components.ContainerRow{
								Page:    components.Page{Key: "seer_gdelt.EventDetailHeader"},
								Classes: "flex flex-wrap justify-between gap-2 items-start w-full",
								Children: []components.PageInterface{
									&components.FieldTitle{
										Getter: getters.Format("GDELT event %d", getters.Any(getters.Key[uint]("$in.ID"))),
									},
									&components.ButtonLink{
										Label: "Edit",
										Link: lago.RoutePath("seer_gdelt.EventUpdateRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("$in.ID")),
										}),
										Classes: "btn-primary btn-sm",
									},
								},
							},
						}, gdeltEventDetailRows()...),
					},
				},
			},
		},
	})
}

func registerGDELTEventFormPages() {
	createFormName := getters.Static("seer_gdelt.EventCreateForm")
	updateFormName := getters.Static("seer_gdelt.EventUpdateForm")
	deleteFormName := getters.Static("seer_gdelt.EventDeleteForm")

	lago.RegistryPage.Register("seer_gdelt.EventCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      createFormName,
				ActionURL: lago.RoutePath("seer_gdelt.EventCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Event]{
						Getter:   getters.Static(Event{}),
						Attr:     getters.FormBubbling(createFormName),
						Title:    "Create GDELT event",
						Subtitle: "Manual row in local seer_gdelt_events table. Global event ID must be unique.",
						Classes:  "@container max-w-5xl",
						ChildrenInput: []components.PageInterface{
							&components.ContainerColumn{
								Page:     components.Page{Key: "seer_gdelt.EventCreateFields"},
								Classes:  "gap-2",
								Children: gdeltEventFormInputs(),
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create event"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.EventUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.EventDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: updateFormName,
				ActionURL: lago.RoutePath("seer_gdelt.EventUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Event]{
						Getter:   getters.Key[Event]("gdeltEvent"),
						Attr:     getters.FormBubbling(updateFormName),
						Title:    "Edit GDELT event",
						Subtitle: "Updates this row. BigQuery searches upsert by global event ID and do not delete other stored events.",
						Classes:  "@container max-w-5xl",
						ChildrenInput: []components.PageInterface{
							&components.ContainerColumn{
								Page:     components.Page{Key: "seer_gdelt.EventUpdateFields"},
								Classes:  "gap-2",
								Children: gdeltEventFormInputs(),
							},
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center w-full",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save changes"},
									&components.ButtonModalForm{
										Label: "Delete",
										Icon:  "trash",
										Name:  deleteFormName,
										Url: lago.RoutePath("seer_gdelt.EventDeleteRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
										}),
										FormPostURL: lago.RoutePath("seer_gdelt.EventDeleteRoute", map[string]getters.Getter[any]{
											"id": getters.Any(getters.Key[uint]("gdeltEvent.ID")),
										}),
										ModalUID: "seer-gdelt-event-delete-modal",
										Classes:  "btn-outline btn-error btn-sm",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_gdelt.EventDeleteForm", &components.Modal{
		UID: "seer-gdelt-event-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete this event row?",
				Message: "Removes only this local database row (not GDELT BigQuery data).",
				Attr:    getters.FormBubbling(deleteFormName),
			},
		},
	})
}
