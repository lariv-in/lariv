package p_seer_websites

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_seer_intel"
)

// websiteURLStringFromMapURLField reads a flattened struct URL field from context (map from [getters.MapFromStruct]).
// structMapKey is e.g. "$in.URL" or "$row.URL". Value is [lago.PageURL] or a plain string (e.g. form merge).
func websiteURLStringFromMapURLField(structMapKey string) getters.Getter[string] {
	return getters.Map(getters.Key[any](structMapKey), func(_ context.Context, v any) (string, error) {
		switch t := v.(type) {
		case nil:
			return "", nil
		case string:
			return strings.TrimSpace(t), nil
		case lago.PageURL:
			return t.String(), nil
		default:
			return strings.TrimSpace(fmt.Sprint(t)), nil
		}
	})
}

// websiteURLStringFromInContext reads $in.URL for [components.Detail] / [components.FormComponent] children.
func websiteURLStringFromInContext() getters.Getter[string] {
	return websiteURLStringFromMapURLField("$in.URL")
}

// websiteURLStringFromRowContext reads $row.URL for [components.DataTable] row cells.
func websiteURLStringFromRowContext() getters.Getter[string] {
	return websiteURLStringFromMapURLField("$row.URL")
}

func websiteIntelMissingGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		site, err := getters.Key[Website]("website")(ctx)
		if err != nil {
			return false, err
		}
		if site.ID == 0 {
			return false, nil
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			return false, err
		}
		exists, err := p_seer_intel.IntelExistsForSource(ctx, db, (Website{}).Kind(), site.ID)
		if err != nil {
			return false, err
		}
		return !exists, nil
	}
}

func websiteIntelPresentGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		site, err := getters.Key[Website]("website")(ctx)
		if err != nil {
			return false, err
		}
		if site.ID == 0 {
			return false, nil
		}
		db, err := getters.DBFromContext(ctx)
		if err != nil {
			return false, err
		}
		ok, err := p_seer_intel.IntelExistsForSource(ctx, db, (Website{}).Kind(), site.ID)
		return ok, err
	}
}

func websiteIntelDetailHrefGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		site, err := getters.Key[Website]("website")(ctx)
		if err != nil || site.ID == 0 {
			return "", err
		}
		return p_seer_intel.IntelDetailPathForSource(ctx, (Website{}).Kind(), site.ID)
	}
}

func websiteDetailContentColumn() components.PageInterface {
	return components.ContainerColumn{
		Page: components.Page{Key: "seer_websites.WebsiteDetailContent"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Page:    components.Page{Key: "seer_websites.WebsiteDetailHeader"},
				Classes: "flex flex-wrap justify-between items-start gap-2 w-full",
				Children: []components.PageInterface{
					&components.FieldTitle{Getter: websiteURLStringFromInContext()},
					&components.ShowIf{
						Page:   components.Page{Key: "seer_websites.WebsiteDetailAddIntelWrap"},
						Getter: websiteIntelMissingGetter(),
						Children: []components.PageInterface{
							&components.ButtonPost{
								Page:    components.Page{Key: "seer_websites.WebsiteDetailAddIntelBtn"},
								Label:   "Add to Intel",
								URL:     lago.RoutePath("seer_websites.WebsiteAddIntelRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Icon:    "document-plus",
								Classes: "btn-outline btn-primary btn-sm shrink-0",
							},
						},
					},
					&components.ButtonModalForm{
						Page:        components.Page{Key: "seer_websites.WebsiteDetailDeleteBtn"},
						Label:       "Delete",
						Icon:        "trash",
						Name:        getters.Static("seer_websites.WebsiteDeleteForm"),
						Url:         lago.RoutePath("seer_websites.WebsiteDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("website.ID"))}),
						FormPostURL: lago.RoutePath("seer_websites.WebsiteDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("website.ID"))}),
						ModalUID:    "seer-website-delete-modal",
						Classes:     "btn-outline btn-error btn-sm shrink-0",
					},
				},
			},
			&components.LabelInline{
				Title: "Markdown",
				Children: []components.PageInterface{
					&components.FieldMarkdown{
						Getter:  getters.Key[string]("$in.Markdown"),
						Classes: "prose prose-sm max-w-none max-h-96 overflow-y-auto",
					},
				},
			},
			&components.ShowIf{
				Page:   components.Page{Key: "seer_websites.WebsiteDetailIntelLinkWrap"},
				Getter: websiteIntelPresentGetter(),
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Intel",
						Children: []components.PageInterface{
							&components.FieldLink{
								Page:    components.Page{Key: "seer_websites.WebsiteDetailIntelLink"},
								Href:    websiteIntelDetailHrefGetter(),
								Label:   getters.Static("View intel"),
								Classes: "link link-primary",
							},
						},
					},
				},
			},
		},
	}
}

func registerWebsiteDetailPages() {
	lago.RegistryPage.Register("seer_websites.WebsiteDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_websites.WebsiteDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Website]{
				Getter: getters.Key[Website]("website"),
				Children: []components.PageInterface{
					websiteDetailContentColumn(),
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_websites.WebsiteDeleteForm", &components.Modal{
		UID: "seer-website-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete saved website?",
				Message: "Clears markdown in this app. URL kept for audit trail. Row is soft-deleted and hidden from lists.",
				Attr:    getters.FormBubbling(getters.Static("seer_websites.WebsiteDeleteForm")),
			},
		},
	})
}
