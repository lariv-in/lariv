package p_lacerate

import (
	"context"
	"fmt"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"gorm.io/datatypes"
)

func redditSourceListTableConfig() components.PageInterface {
	return &components.FieldList[string]{
		Getter: getters.JSONList[string](getters.Key[datatypes.JSON]("$row.Reddit.Subreddits")),
		Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row")},
		},
	}
}

func redditSourceFormFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.RedditSourceFormFields"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Subreddits"),
				Children: []components.PageInterface{
					&InputSubredditList{
						InputStringList: components.InputStringList{
							Label:   "Subreddits",
							Name:    "Subreddits",
							Classes: "w-full",
							Getter:  getters.Key[datatypes.JSON]("$in.Reddit.Subreddits"),
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.SearchQuery"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:   "Search query",
						Name:    "SearchQuery",
						Getter:  getters.Key[string]("$in.Reddit.SearchQuery"),
						Classes: "w-full",
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.MaxFreshPosts"),
				Children: []components.PageInterface{
					&components.InputNumber[uint]{
						Label:  "Max fresh posts per fetch",
						Name:   "MaxFreshPosts",
						Getter: getters.Map(getters.Key[uint]("$in.Reddit.MaxFreshPosts"), func(_ context.Context, n uint) (uint, error) {
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

func redditSourceDetailFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "lacerate.SourceDetailRedditFields"},
		Children: []components.PageInterface{
			&components.LabelInline{
				Title: "Subreddits",
				Children: []components.PageInterface{
					&components.FieldList[string]{
						Getter: getters.Map(getters.JSONList[string](getters.Key[datatypes.JSON]("$in.Reddit.Subreddits")), func(_ context.Context, items []string) ([]string, error) {
							if len(items) == 0 {
								return []string{"(none)"}, nil
							}
							return items, nil
						}),
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row")},
						},
					},
				},
			},
			&components.LabelInline{
				Title: "Search query",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.IfOrElse(
							getters.Map(getters.Key[string]("$in.Reddit.SearchQuery"), func(_ context.Context, s string) (string, error) {
								s = strings.TrimSpace(s)
								if s == "" {
									return "", nil
								}
								return s, nil
							}),
							getters.Static("(none — uses default subreddit listing)"),
						),
					},
				},
			},
			&components.LabelInline{
				Title: "Max fresh posts per fetch",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.Map(getters.Key[uint]("$in.Reddit.MaxFreshPosts"), func(_ context.Context, n uint) (string, error) {
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
