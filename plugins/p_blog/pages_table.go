package p_blog

import (
	"time"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesTables() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		// Blog Articles Table Page
		{Key: "p_blog.BlogListPage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogListMenu"},
			},
			Children: []components.PageInterface{
				&components.DataTable[Blog]{
					UID:      "blogs-table",
					Title:    "Blog Articles",
					Subtitle: "Manage posts and publications",
					Classes:  "w-full",
					Data:     getters.Key[components.ObjectList[Blog]]("blogs"),
					Actions: []components.PageInterface{
						&components.ButtonLink{
							Link:    lariv.RoutePath("p_blog.BlogCreateRoute", nil),
							Icon:    "plus",
							Classes: "btn-square btn-outline btn-sm",
						},
					},
					RowAttr: getters.RowAttrNavigate(lariv.RoutePath("p_blog.BlogDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
					Columns: []components.TableColumn{
						{Label: "Title", Name: "Title", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						}},
						{Label: "Slug", Name: "Slug", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Slug")},
						}},
						{Label: "Author", Name: "CreatedBy", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.CreatedBy.Name")},
						}},
						{Label: "Updated At", Name: "UpdatedAt", Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")},
						}},
					},
				},
			},
		}},

		// Blog Tags Table Page
		{Key: "p_blog.TagListPage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogListMenu"},
			},
			Children: []components.PageInterface{
				&components.DataTable[BlogTag]{
					UID:      "tags-table",
					Title:    "Blog Tags",
					Subtitle: "Manage ltree tags and categories",
					Classes:  "w-full",
					Data:     getters.Key[components.ObjectList[BlogTag]]("tags"),
					Actions: []components.PageInterface{
						&components.ButtonLink{
							Link:    lariv.RoutePath("p_blog.TagCreateRoute", nil),
							Icon:    "plus",
							Classes: "btn-square btn-outline btn-sm",
						},
					},
					RowAttr: getters.RowAttrNavigate(lariv.RoutePath("p_blog.TagDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
					Columns: []components.TableColumn{
						{Label: "Tag Name (ltree)", Name: "Name", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						}},
						{Label: "Updated At", Name: "UpdatedAt", Children: []components.PageInterface{
							&components.FieldDatetime{Getter: getters.Key[time.Time]("$row.UpdatedAt")},
						}},
					},
				},
			},
		}},
	}
}
