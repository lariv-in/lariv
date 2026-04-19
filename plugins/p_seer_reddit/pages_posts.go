package p_seer_reddit

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func redditPostListTableColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Post ID",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.PostID")},
			},
		},
		{
			Label: "Title",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.Title")},
			},
		},
		{
			Label: "r/",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.Subreddit")},
			},
		},
	}
}

func registerRedditPostPages() {
	lago.RegistryPage.Register("seer_reddit.RedditPostTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditMenu"},
		},
		Children: []components.PageInterface{
			&components.GetterPage{
				Page:   components.Page{Key: "seer_reddit.RedditPostTableShell"},
				Getter: redditPostListTableShellGetter(false),
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditPostTableBySource", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditSourceDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.GetterPage{
				Page:   components.Page{Key: "seer_reddit.RedditPostTableBySourceShell"},
				Getter: redditPostListTableShellGetter(true),
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditPostDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "seer_reddit.RedditPostDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[RedditPost]{
				Getter: getters.Key[RedditPost]("redditPost"),
				Children: []components.PageInterface{
					&components.GetterPage{
						Page:   components.Page{Key: "seer_reddit.RedditPostDetailShell"},
						Getter: redditPostDetailShellGetter(redditPostDetailContentColumn()),
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("seer_reddit.RedditPostDeleteForm", &components.Modal{
		UID: "seer-reddit-post-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete saved post?",
				Message: "Clears title, body, and metadata in this app. Reddit Post ID stays for deduplication. The post is soft-deleted and hidden from lists.",
				Attr:    getters.FormBubbling(getters.Static("seer_reddit.RedditPostDeleteForm")),
			},
		},
	})
}
