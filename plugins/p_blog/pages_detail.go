package p_blog

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesDetail() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		// Blog Article Detail Page
		{Key: "p_blog.BlogDetailPage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.Detail[Blog]{
					Page:   components.Page{Key: "p_blog.BlogDetailContent"},
					Getter: getters.Key[Blog]("blog"),
					Children: []components.PageInterface{
						&components.ContainerColumn{
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
								&components.LabelInline{
									Title:   "Slug",
									Classes: "mt-2 block",
									Children: []components.PageInterface{
										&components.FieldText{Getter: getters.Key[string]("$in.Slug")},
									},
								},
								&components.LabelInline{
									Title:   "Description",
									Classes: "mt-4 block",
									Children: []components.PageInterface{
										&components.FieldText{Getter: getters.Key[string]("$in.Description")},
									},
								},
								&components.LabelInline{
									Title:   "Created By (Author)",
									Classes: "mt-4 block",
									Children: []components.PageInterface{
										&components.FieldText{Getter: getters.Key[string]("$in.CreatedBy.Name")},
									},
								},
								&components.LabelInline{
									Title:   "Tags",
									Classes: "mt-4 block",
									Children: []components.PageInterface{
										&components.FieldManyToMany[BlogTag]{
											Getter:  getters.Key[[]BlogTag]("$in.Tags"),
											Display: getters.Key[string]("$in.Name"),
										},
									},
								},
								&components.LabelInline{
									Title:   "Content (Markdown)",
									Classes: "mt-6 block",
									Children: []components.PageInterface{
										&components.FieldMarkdown{
											Getter: getters.Key[string]("$in.Content"),
										},
									},
								},
							},
						},
					},
				},
			},
		}},

		// Blog Tag Detail Page
		{Key: "p_blog.TagDetailPage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.TagDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.Detail[BlogTag]{
					Page:   components.Page{Key: "p_blog.TagDetailContent"},
					Getter: getters.Key[BlogTag]("tag"),
					Children: []components.PageInterface{
						&components.ContainerColumn{
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
								&components.LabelInline{
									Title:   "Associated Articles",
									Classes: "mt-4 block",
									Children: []components.PageInterface{
										&components.FieldManyToMany[Blog]{
											Getter:  getters.Key[[]Blog]("$in.Blogs"),
											Display: getters.Key[string]("$in.Title"),
										},
									},
								},
							},
						},
					},
				},
			},
		}},
	}
}
