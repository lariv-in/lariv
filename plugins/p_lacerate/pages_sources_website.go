package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func websiteSourceListTableConfig() components.PageInterface {
	return &components.FieldText{
		Getter: getters.Key[string]("$row.Website.URL"),
	}
}

func websiteSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.WebsiteSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.URL"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "URL",
						Name:    "URL",
						Getter:  getters.Key[string]("$in.Website.URL"),
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func websiteSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailWebsiteFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "URL",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.IfOrElse(
							getters.Map(getters.Key[string]("$in.Website.URL"), func(_ context.Context, s string) (string, error) {
								s = strings.TrimSpace(s)
								if s == "" {
									return "", nil
								}
								return s, nil
							}),
							getters.Static("(none)"),
						),
					},
				},
			},
		},
	}
}
