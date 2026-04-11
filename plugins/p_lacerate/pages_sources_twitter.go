package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"gorm.io/datatypes"
)

func twitterSourceListTableConfig() components.PageInterface {
	return &components.FieldList[string]{
		Getter: getters.JSONList[string](getters.Key[datatypes.JSON]("$row.Twitter.Handles")),
		Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row")},
		},
	}
}

func twitterSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.TwitterSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Handles"),
				Children: []components.PageInterface{
					&InputTwitterHandleList{
						InputStringList: components.InputStringList{
							Label:   "Twitter handles",
							Name:    "Handles",
							Classes: "w-full",
							Getter:  getters.Key[datatypes.JSON]("$in.Twitter.Handles"),
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.MaxFreshPosts"),
				Children: []components.PageInterface{
					&components.InputNumber[uint]{
						Label:  "Max fresh posts per fetch",
						Name:   "MaxFreshPosts",
						Getter: getters.Map(getters.Key[uint]("$in.Twitter.MaxFreshPosts"), func(_ context.Context, n uint) (uint, error) {
							if n == 0 {
								return sourceDefaultMaxFreshPosts, nil
							}
							return n, nil
						}),
						Classes: "w-full",
					},
				},
			},
		},
	}
}

func twitterSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailTwitterFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "Handles",
				Children: []components.PageInterface{
					&components.FieldList[string]{
						Getter: getters.Map(getters.JSONList[string](getters.Key[datatypes.JSON]("$in.Twitter.Handles")), func(_ context.Context, items []string) ([]string, error) {
							lines := make([]string, 0, len(items))
							for _, s := range items {
								h := strings.TrimSpace(s)
								if h != "" {
									lines = append(lines, "@"+h)
								}
							}
							if len(lines) == 0 {
								return []string{"(none)"}, nil
							}
							return lines, nil
						}),
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row")},
						},
					},
				},
			},
			&components.LabelInline{
				Title: "Max fresh posts per fetch",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.Map(getters.Key[uint]("$in.Twitter.MaxFreshPosts"), func(_ context.Context, n uint) (string, error) {
							if n == 0 {
								n = sourceDefaultMaxFreshPosts
							}
							return fmt.Sprintf("%d", n), nil
						}),
					},
				},
			},
		},
	}
}
