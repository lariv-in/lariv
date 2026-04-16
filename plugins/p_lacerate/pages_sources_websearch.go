package p_lacerate

import (
	"context"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
)

func websearchSourceListTableConfig() components.PageInterface {
	return &components.FieldText{
		Getter: getters.Key[string]("$row.Websearch.Query"),
	}
}

func websearchSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.WebsearchSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Query"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Query",
						Name:    "Query",
						Getter:  getters.Key[string]("$in.Websearch.Query"),
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func websearchSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailWebsearchFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "Query",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.IfOrElse(
							getters.Map(getters.Key[string]("$in.Websearch.Query"), func(_ context.Context, s string) (string, error) {
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
