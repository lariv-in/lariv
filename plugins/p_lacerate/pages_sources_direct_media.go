package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func directMediaSourceListTableConfig() components.PageInterface {
	return &components.FieldText{
		Getter: getters.Key[string]("$row.DirectMedia.URL"),
	}
}

func directMediaSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.DirectMediaSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.URL"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Direct media URL",
						Name:    "URL",
						Getter:  getters.Key[string]("$in.DirectMedia.URL"),
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func directMediaSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailDirectMediaFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "URL",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.IfOrElse(
							getters.Map(getters.Key[string]("$in.DirectMedia.URL"), func(_ context.Context, s string) (string, error) {
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
