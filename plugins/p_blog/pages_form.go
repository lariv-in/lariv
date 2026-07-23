package p_blog

import (
	"context"

	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/plugins/p_users"
	"github.com/lariv-in/lariv/registry"
)

func blogFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Title"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Title",
						Name:     "Title",
						Required: true,
						Getter:   getters.Key[string]("$in.Title"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Slug"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:  "Slug (Auto-generated from title if blank)",
						Name:   "Slug",
						Getter: getters.Key[string]("$in.Slug"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Description"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Description",
						Name:   "Description",
						Rows:   3,
						Getter: getters.Key[string]("$in.Description"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.CreatedByID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_users.User]{
						Label:       "Author (User)",
						Name:        "CreatedByID",
						Required:    true,
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select user author...",
						Url:         lariv.RoutePath("p_users.SelectRoute", nil),
						Getter: getters.Getter[p_users.User](func(ctx context.Context) (p_users.User, error) {
							if u, err := getters.Key[p_users.User]("$in.CreatedBy")(ctx); err == nil && u.ID != 0 {
								return u, nil
							}
							if user, ok := p_users.UserFromContextOptional(ctx); ok {
								return user, nil
							}
							return p_users.User{}, nil
						}),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Tags"),
				Children: []components.PageInterface{
					&components.InputManyToMany[BlogTag]{
						Label:       "Tags",
						Name:        "Tags",
						Required:    false,
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select tags...",
						Url:         lariv.RoutePath("p_blog.TagSelectRoute", nil),
						Getter:      getters.Key[[]BlogTag]("$in.Tags"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Content"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Content (Markdown)",
						Name:   "Content",
						Rows:   12,
						Getter: getters.Key[string]("$in.Content"),
					},
				},
			},
		},
	}
}

func tagFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Tag Name (ltree)",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
					},
				},
			},
		},
	}
}

func pageEntriesForms() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		// Blog Create Page
		{Key: "p_blog.BlogCreatePage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogListMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_blog.BlogCreateForm"),
					ActionURL: lariv.RoutePath("p_blog.BlogCreateRoute", nil),
					Children: []components.PageInterface{
						&components.FormComponent[Blog]{
							Page: components.Page{Key: "p_blog.BlogCreateForm"},
							Attr: getters.FormBubbling(getters.Static("p_blog.BlogCreateForm")),
							ChildrenInput: []components.PageInterface{
								blogFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ButtonSubmit{Label: "Create Blog Post"},
							},
						},
					},
				},
			},
		}},

		// Blog Update Page
		{Key: "p_blog.BlogUpdatePage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name: getters.Static("p_blog.BlogUpdateForm"),
					ActionURL: lariv.RoutePath("p_blog.BlogUpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("blog.ID")),
					}),
					Children: []components.PageInterface{
						&components.FormComponent[Blog]{
							Page:   components.Page{Key: "p_blog.BlogUpdateForm"},
							Attr:   getters.FormBubbling(getters.Static("p_blog.BlogUpdateForm")),
							Getter: getters.Key[Blog]("blog"),
							ChildrenInput: []components.PageInterface{
								blogFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ContainerRow{
									Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
									Children: []components.PageInterface{
										&components.ContainerRow{
											Classes: "flex justify-end gap-2",
											Children: []components.PageInterface{
												&components.ButtonSubmit{Label: "Save Changes"},
												&components.ButtonModalForm{
													Label:       "Delete",
													Icon:        "trash",
													Name:        getters.Static("p_blog.BlogDeleteForm"),
													Url:         lariv.RoutePath("p_blog.BlogDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("blog.ID"))}),
													FormPostURL: lariv.RoutePath("p_blog.BlogDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("blog.ID"))}),
													ModalUID:    "blog-delete-modal",
													Classes:     "btn-error",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},

		// Blog Delete Modal
		{Key: "p_blog.BlogDeleteForm", Value: &components.Modal{
			UID: "blog-delete-modal",
			Children: []components.PageInterface{
				&components.DeleteConfirmation{
					Page:    components.Page{Key: "p_blog.BlogDeleteForm"},
					Title:   "Confirm deletion",
					Message: "Are you sure you want to delete this blog post? This action cannot be undone.",
					Attr:    getters.FormBubbling(getters.Static("p_blog.BlogDeleteForm")),
				},
			},
		}},

		// Tag Create Page
		{Key: "p_blog.TagCreatePage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.BlogListMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_blog.TagCreateForm"),
					ActionURL: lariv.RoutePath("p_blog.TagCreateRoute", nil),
					Children: []components.PageInterface{
						&components.FormComponent[BlogTag]{
							Page: components.Page{Key: "p_blog.TagCreateForm"},
							Attr: getters.FormBubbling(getters.Static("p_blog.TagCreateForm")),
							ChildrenInput: []components.PageInterface{
								tagFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ButtonSubmit{Label: "Create Tag"},
							},
						},
					},
				},
			},
		}},

		// Tag Update Page
		{Key: "p_blog.TagUpdatePage", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_blog.TagDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name: getters.Static("p_blog.TagUpdateForm"),
					ActionURL: lariv.RoutePath("p_blog.TagUpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("tag.ID")),
					}),
					Children: []components.PageInterface{
						&components.FormComponent[BlogTag]{
							Page:   components.Page{Key: "p_blog.TagUpdateForm"},
							Attr:   getters.FormBubbling(getters.Static("p_blog.TagUpdateForm")),
							Getter: getters.Key[BlogTag]("tag"),
							ChildrenInput: []components.PageInterface{
								tagFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ContainerRow{
									Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
									Children: []components.PageInterface{
										&components.ContainerRow{
											Classes: "flex justify-end gap-2",
											Children: []components.PageInterface{
												&components.ButtonSubmit{Label: "Save Changes"},
												&components.ButtonModalForm{
													Label:       "Delete",
													Icon:        "trash",
													Name:        getters.Static("p_blog.TagDeleteForm"),
													Url:         lariv.RoutePath("p_blog.TagDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("tag.ID"))}),
													FormPostURL: lariv.RoutePath("p_blog.TagDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("tag.ID"))}),
													ModalUID:    "tag-delete-modal",
													Classes:     "btn-error",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}},

		// Tag Delete Modal
		{Key: "p_blog.TagDeleteForm", Value: &components.Modal{
			UID: "tag-delete-modal",
			Children: []components.PageInterface{
				&components.DeleteConfirmation{
					Page:    components.Page{Key: "p_blog.TagDeleteForm"},
					Title:   "Confirm deletion",
					Message: "Are you sure you want to delete this tag? This action cannot be undone.",
					Attr:    getters.FormBubbling(getters.Static("p_blog.TagDeleteForm")),
				},
			},
		}},
	}
}
