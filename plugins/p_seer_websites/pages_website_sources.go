package p_seer_websites

import (
	"context"
	"fmt"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	gomponents "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func websiteSourceFetchButtonAttrGetter() getters.Getter[gomponents.Node] {
	return func(ctx context.Context) (gomponents.Node, error) {
		id, err := getters.Key[uint]("websiteSource.ID")(ctx)
		if err != nil {
			return nil, err
		}
		if WebsiteSourceCrawlIsRunning(id) {
			return Disabled(), nil
		}
		return nil, nil
	}
}

func websiteSourceDetailWorkerLabel(ctx context.Context) (string, error) {
	ws, err := getters.Key[WebsiteSource]("websiteSource")(ctx)
	if err != nil {
		return "", err
	}
	if ws.WebsiteRunnerID == nil || *ws.WebsiteRunnerID == 0 {
		return "—", nil
	}
	if ws.WebsiteRunner != nil {
		return ws.WebsiteRunner.Name, nil
	}
	db, err := getters.DBFromContext(ctx)
	if err != nil {
		return "", err
	}
	var wr WebsiteRunner
	if err := db.WithContext(ctx).Where("id = ?", *ws.WebsiteRunnerID).Take(&wr).Error; err != nil {
		return fmt.Sprintf("id %d", *ws.WebsiteRunnerID), nil
	}
	return wr.Name, nil
}

func websiteSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "seer_websites.WebsiteSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Page:  components.Page{Key: "seer_websites.WebsiteSourceForm.WebsiteRunnerID"},
				Error: getters.Key[error]("$error.WebsiteRunnerID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[WebsiteRunner]{
						Label:       "Worker",
						Name:        "WebsiteRunnerID",
						Url:         lago.RoutePath("seer_websites.WebsiteRunnerSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Optional worker…",
						Required:    false,
						Getter:      getters.Association[WebsiteRunner](getters.Deref(getters.Key[*uint]("$in.WebsiteRunnerID"))),
						Classes:     "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.URL"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Seed URL",
						Name:     "URL",
						Required: true,
						Getter:   websiteURLStringFromMapURLField("$in.URL"),
						Classes:  "w-full max-w-xl",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Depth"),
				Children: []components.PageInterface{
					&components.InputNumber[uint]{
						Label:    "Link depth",
						Name:     "Depth",
						Getter:   getters.Key[uint]("$in.Depth"),
						Classes:  "w-full max-w-md",
					},
				},
			},
		},
	}
}

func registerWebsiteSourcePages() {
	lago.RegistryPage.Register("seer_websites.WebsiteSourceTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[WebsiteSource]{
				Page:    components.Page{Key: "seer_websites.WebsiteSourceTableBody"},
				UID:     "seer-website-sources-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[WebsiteSource]]("websiteSources"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{Link: lago.RoutePath("seer_websites.WebsiteSourceCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("seer_websites.WebsiteSourceDetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "URL",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter:  websiteURLStringFromMapURLField("$row.URL"),
								Classes: "break-all max-w-prose",
							},
						},
					},
					{
						Label: "Depth",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.Depth")))},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteSourceDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[WebsiteSource]{
				Getter: getters.Key[WebsiteSource]("websiteSource"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "seer_websites.WebsiteSourceDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{
								Getter: getters.Format("Source #%d", getters.Any(getters.Key[uint]("$in.ID"))),
							},
							&components.LabelInline{
								Title: "Seed URL",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  websiteURLStringFromMapURLField("$in.URL"),
										Classes: "break-all",
									},
								},
							},
							&components.LabelInline{
								Title: "Link depth",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.Depth")))},
								},
							},
							&components.LabelInline{
								Title: "Worker",
								Children: []components.PageInterface{
									&components.FieldText{Getter: websiteSourceDetailWorkerLabel},
								},
							},
							&components.ContainerRow{
								Page:    components.Page{Key: "seer_websites.WebsiteSourceFetchRow"},
								Classes: "flex flex-wrap gap-2 items-center mt-4",
								Children: []components.PageInterface{
									&components.ButtonPost{
										Label: "Run crawl now",
										URL: lago.RoutePath("seer_websites.WebsiteSourceFetchRoute", map[string]getters.Getter[any]{
											"source_id": getters.Any(getters.Key[uint]("websiteSource.ID")),
										}),
										Icon:    "arrow-path",
										Classes: "btn-outline btn-primary btn-sm",
										Attr:    websiteSourceFetchButtonAttrGetter(),
									},
								},
							},
						},
					},
				},
			},
		},
	})
}
