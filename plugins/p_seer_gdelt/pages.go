package p_seer_gdelt

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

var gdeltSortChoices = []registry.Pair[string, string]{
	{Key: gdeltSortDateDesc, Value: "Newest first"},
	{Key: gdeltSortDateAsc, Value: "Oldest first"},
	{Key: gdeltSortMentionsDesc, Value: "Most mentions"},
}

func init() {
	registerGDELTMenuPages()
	registerGDELTSearchPages()
	registerGDELTMapPages()
	registerGDELTEventPages()
}

func registerGDELTMenuPages() {
	lago.RegistryPage.Register("seer_gdelt.Menu", &components.SidebarMenu{
		Title: getters.Static("GDELT"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Search"),
				Url:   lago.RoutePath("seer_gdelt.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Events"),
				Url:   lago.RoutePath("seer_gdelt.EventListRoute", nil),
			},
			&gdeltMapSidebarLink{Page: components.Page{Key: "seer_gdelt.MenuMapLink"}},
		},
	})
}

func registerGDELTSearchPages() {
	lago.RegistryPage.Register("seer_gdelt.SearchPage", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_gdelt.Menu"},
		},
		Children: []components.PageInterface{
			&components.ContainerColumn{
				Page:    components.Page{Key: "seer_gdelt.SearchPageBody"},
				Classes: "container max-w-6xl mx-auto gap-4",
				Children: []components.PageInterface{
					&components.FormComponent[map[string]any]{
						Page:     components.Page{Key: "seer_gdelt.SearchForm"},
						Attr:     getters.FormBoostedGet(lago.RoutePath("seer_gdelt.SearchRoute", nil)),
						Title:    "GDELT search",
						Subtitle: "Search GDELT events in BigQuery with guided filters. Configure [Plugins.p_seer_gdelt] projectID and optional credentialsFile, then share searches by URL.",
						Classes:  "@container rounded-box border border-base-300 bg-base-100 p-4",
						ChildrenInput: []components.PageInterface{
							gdeltSearchFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Page:    components.Page{Key: "seer_gdelt.SearchActions"},
								Classes: "flex-wrap gap-2 mt-3",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Search GDELT"},
									&components.ButtonLink{
										Label:   "Clear",
										Link:    lago.RoutePath("seer_gdelt.DefaultRoute", nil),
										Classes: "btn-ghost",
									},
								},
							},
						},
					},
					&gdeltSearchResultsSection{},
				},
			},
		},
	})
}

func gdeltSearchFields() components.PageInterface {
	return &components.ContainerColumn{
		Page:    components.Page{Key: "seer_gdelt.SearchFields"},
		Classes: "gap-2",
		Children: []components.PageInterface{
			&components.ContainerError{
				Page:  components.Page{Key: "seer_gdelt.SearchField.Query"},
				Error: getters.Key[error]("$error.Query"),
				Children: []components.PageInterface{
					&components.InputText{
						Page:     components.Page{Key: "seer_gdelt.Query"},
						Label:    "Keywords or phrase",
						Name:     "Query",
						Getter:   getters.Key[string]("$get.Query"),
						Required: false,
						Classes:  "w-full",
					},
				},
			},
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.FilterRow1"},
				Classes: "flex-col @lg:flex-row gap-3",
				Children: []components.PageInterface{
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.Domain"},
						Error: getters.Key[error]("$error.Domain"),
						Children: []components.PageInterface{
							&components.InputText{
								Page:    components.Page{Key: "seer_gdelt.Domain"},
								Label:   "Domain",
								Name:    "Domain",
								Getter:  getters.Key[string]("$get.Domain"),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.ActionCountry"},
						Error: getters.Key[error]("$error.ActionCountry"),
						Children: []components.PageInterface{
							&components.InputText{
								Page:    components.Page{Key: "seer_gdelt.ActionCountry"},
								Label:   "Action country code (FIPS)",
								Name:    "ActionCountry",
								Getter:  getters.Key[string]("$get.ActionCountry"),
								Classes: "w-full",
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_gdelt.FilterRow2"},
				Classes: "flex-col @lg:flex-row gap-3",
				Children: []components.PageInterface{
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.StartDate"},
						Error: getters.Key[error]("$error.StartDate"),
						Children: []components.PageInterface{
							&components.InputDate{
								Page:    components.Page{Key: "seer_gdelt.StartDate"},
								Label:   "Start date",
								Name:    "StartDate",
								Getter:  gdeltDateGetter("StartDate"),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.EndDate"},
						Error: getters.Key[error]("$error.EndDate"),
						Children: []components.PageInterface{
							&components.InputDate{
								Page:    components.Page{Key: "seer_gdelt.EndDate"},
								Label:   "End date",
								Name:    "EndDate",
								Getter:  gdeltDateGetter("EndDate"),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.MinMentions"},
						Error: getters.Key[error]("$error.MinMentions"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Page:    components.Page{Key: "seer_gdelt.MinMentions"},
								Label:   "Minimum mentions",
								Name:    "MinMentions",
								Getter:  gdeltUintGetter("MinMentions", 0),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.Sort"},
						Error: getters.Key[error]("$error.Sort"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Page:    components.Page{Key: "seer_gdelt.Sort"},
								Label:   "Sort",
								Name:    "Sort",
								Choices: getters.Static(gdeltSortChoices),
								Getter:  gdeltPairGetterWithDefault("Sort", gdeltSortChoices, gdeltSortDateDesc),
								Classes: "w-full",
							},
						},
					},
					&components.ContainerError{
						Page:  components.Page{Key: "seer_gdelt.SearchField.MaxRecords"},
						Error: getters.Key[error]("$error.MaxRecords"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Page:    components.Page{Key: "seer_gdelt.MaxRecords"},
								Label:   "Max records",
								Name:    "MaxRecords",
								Getter:  gdeltUintGetter("MaxRecords", defaultGDELTMaxRecords),
								Classes: "w-full",
							},
						},
					},
				},
			},
		},
	}
}

type gdeltSearchResultsSection struct {
	components.Page
}

func (e gdeltSearchResultsSection) GetKey() string {
	return "seer_gdelt.SearchResultsSection"
}

func (e gdeltSearchResultsSection) GetRoles() []string {
	return nil
}

func (e gdeltSearchResultsSection) Build(ctx context.Context) Node {
	state := gdeltSearchStateFromContext(ctx)
	if !state.Searched {
		return Group{}
	}
	results, _ := gdeltResultsGetter(ctx)
	summary := fmt.Sprintf("Showing %d result(s).", len(results.Items))
	if len(results.Items) == 0 {
		summary = "No matching articles found for current filters."
	}
	return Div(
		Class("rounded-box border border-base-300 bg-base-100 p-4"),
		Div(Class("mb-3"),
			Div(Class("text-xl font-semibold"), Text("Results")),
			Div(Class("text-sm text-gray-500"), Text(summary)),
		),
		components.Render(components.TableListContent[Event]{
			Page: components.Page{Key: "seer_gdelt.ResultsTable"},
			Data: getters.Static(results),
			Columns: []components.TableColumn{
				{
					Label: "Actors",
					Children: []components.PageInterface{
						components.FieldText{Getter: gdeltActorsGetter(), Classes: "whitespace-normal"},
					},
				},
				{
					Label: "Date",
					Children: []components.PageInterface{
						components.FieldText{Getter: gdeltRowDateGetter()},
					},
				},
				{
					Label: "Event",
					Children: []components.PageInterface{
						components.FieldText{Getter: gdeltEventGetter()},
					},
				},
				{
					Label: "Country",
					Children: []components.PageInterface{
						components.FieldText{Getter: getters.Key[string]("$row.ActionGeoCountryCode")},
					},
				},
				{
					Label: "Mentions",
					Children: []components.PageInterface{
						components.FieldText{Getter: gdeltRowMentionsGetter()},
					},
				},
				{
					Label: "Source",
					Children: []components.PageInterface{
						components.FieldLink{
							Href:    getters.Key[string]("$row.SourceURL"),
							Label:   gdeltSourceLabelGetter(),
							Classes: "link link-primary whitespace-normal",
						},
					},
				},
			},
		}, ctx),
	)
}

func gdeltPairGetter(field string, choices []registry.Pair[string, string]) getters.Getter[registry.Pair[string, string]] {
	return gdeltPairGetterWithDefault(field, choices, "")
}

func gdeltPairGetterWithDefault(field string, choices []registry.Pair[string, string], fallback string) getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return registry.Pair[string, string]{}, nil
		}
		s = strings.TrimSpace(s)
		if s == "" {
			s = fallback
		}
		if s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromPairs(s, choices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func gdeltDateGetter(field string) getters.Getter[time.Time] {
	return func(ctx context.Context) (time.Time, error) {
		raw, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return time.Time{}, nil
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return time.Time{}, nil
		}
		t, err := time.Parse(time.DateOnly, raw)
		if err != nil {
			return time.Time{}, nil
		}
		return t, nil
	}
}

func gdeltUintGetter(field string, fallback uint) getters.Getter[uint] {
	return func(ctx context.Context) (uint, error) {
		raw, err := getters.Key[string]("$get." + field)(ctx)
		if err != nil {
			return fallback, nil
		}
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return fallback, nil
		}
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fallback, nil
		}
		return uint(n), nil
	}
}

func gdeltActorsGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		actor1, _ := getters.Key[string]("$row.Actor1Name")(ctx)
		actor2, _ := getters.Key[string]("$row.Actor2Name")(ctx)
		actor1 = strings.TrimSpace(actor1)
		actor2 = strings.TrimSpace(actor2)
		switch {
		case actor1 != "" && actor2 != "":
			return actor1 + " / " + actor2, nil
		case actor1 != "":
			return actor1, nil
		case actor2 != "":
			return actor2, nil
		default:
			return "Unknown actors", nil
		}
	}
}

func gdeltEventGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		eventCode, _ := getters.Key[string]("$row.EventCode")(ctx)
		action, _ := getters.Key[string]("$row.ActionGeoFullName")(ctx)
		eventCode = strings.TrimSpace(eventCode)
		action = strings.TrimSpace(action)
		switch {
		case eventCode != "" && action != "":
			return eventCode + " in " + action, nil
		case eventCode != "":
			return eventCode, nil
		case action != "":
			return action, nil
		default:
			return "Event", nil
		}
	}
}

func gdeltSourceLabelGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		url, _ := getters.Key[string]("$row.SourceURL")(ctx)
		url = strings.TrimSpace(url)
		if url == "" {
			return "", nil
		}
		normalized := strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://"), "www.")
		if idx := strings.Index(normalized, "/"); idx >= 0 {
			normalized = normalized[:idx]
		}
		if normalized != "" {
			return normalized, nil
		}
		return url, nil
	}
}

func gdeltRowDateGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, err := getters.Key[int]("$row.SQLDate")(ctx)
		if err != nil || n == 0 {
			return "", nil
		}
		s := strconv.Itoa(n)
		if len(s) != 8 {
			return s, nil
		}
		return s[:4] + "-" + s[4:6] + "-" + s[6:], nil
	}
}

func gdeltRowMentionsGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		n, err := getters.Key[int]("$row.NumMentions")(ctx)
		if err != nil {
			return "", nil
		}
		return strconv.Itoa(n), nil
	}
}
