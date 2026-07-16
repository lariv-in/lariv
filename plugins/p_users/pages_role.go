package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesRole() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.RoleDetailMenu", Value: &components.SidebarMenu{
			Title: getters.Format("Role: %s", getters.Any(getters.Key[string]("role.Name"))),
			Back: &components.SidebarMenuItem{
				Title: getters.Static("Back to All Roles"),
				Url:   lariv.RoutePath("p_users.RoleListRoute", nil),
			},
			Children: []components.PageInterface{
				&components.SidebarMenuItem{
					Title: getters.Static("Role Detail"),
					Url:   lariv.RoutePath("p_users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
				},
				&components.SidebarMenuItem{
					Title: getters.Static("Edit Role"),
					Url:   lariv.RoutePath("p_users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
				},
			},
		}},
		{Key: "p_users.RoleFilter", Value: &components.FormComponent[Role]{
			Attr: getters.FormBoostedGet(lariv.RoutePath("p_users.RoleListRoute", nil)),

			ChildrenInput: []components.PageInterface{
				&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			},
			ChildrenAction: []components.PageInterface{
				&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				}},
			},
		}},
		{Key: "p_users.RoleTable", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserMenu"},
			},
			Children: []components.PageInterface{
				&components.DataTable[Role]{
					UID:     "role-table",
					Classes: "w-full",
					Data:    getters.Key[components.ObjectList[Role]]("roles"),
					Actions: []components.PageInterface{
						&components.TableButtonFilter{Child: lariv.DynamicPage{Name: "p_users.RoleFilter"}},
						&components.ButtonModalForm{
							Name:        getters.Static("p_users.RoleCreateForm"),
							Url:         lariv.RoutePath("p_users.RoleCreateRoute", nil),
							FormPostURL: lariv.RoutePath("p_users.RoleCreateRoute", nil),
							ModalUID:    "role-create-modal",
							Icon:        "plus",
							Classes:     "btn-square btn-outline btn-sm",
						},
					},
					RowAttr: getters.RowAttrNavigate(lariv.RoutePath("p_users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
					Columns: []components.TableColumn{
						{Label: "Name", Name: "Name", Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						}},
					},
				},
			},
		}},
		{Key: "p_users.RoleCreateForm", Value: &components.Modal{
			Page: components.Page{
				Key: "p_users.RoleCreateModal",
			},
			UID: "role-create-modal",
			Children: []components.PageInterface{
				&components.FormComponent[Role]{
					Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

					Title:    "Create Role",
					Subtitle: "Create a new role",
					ChildrenInput: []components.PageInterface{
						&components.ContainerError{
							Error: getters.Key[error]("$error.Name"),
							Children: []components.PageInterface{
								&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						&components.ContainerRow{
							Classes: "flex justify-end gap-2 mt-2",
							Children: []components.PageInterface{
								&components.ButtonSubmit{Label: "Save Role", Classes: "btn-primary"},
							},
						},
					},
				},
			},
		}},
		{Key: "p_users.RoleUpdateForm", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.RoleDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_users.RoleUpdateForm"),
					ActionURL: lariv.RoutePath("p_users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
					Children: []components.PageInterface{
						&components.FormComponent[Role]{
							Getter: getters.Key[Role]("role"),
							Attr:   getters.FormBubbling(getters.Static("p_users.RoleUpdateForm")),

							Title:    "Edit Role",
							Subtitle: "Update role details",
							ChildrenInput: []components.PageInterface{
								&components.ContainerError{
									Error: getters.Key[error]("$error.Name"),
									Children: []components.PageInterface{
										&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
									},
								},
							},
							ChildrenAction: []components.PageInterface{
								&components.ContainerRow{
									Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
									Children: []components.PageInterface{
										&components.ContainerRow{
											Classes: "flex justify-end gap-2",
											Children: []components.PageInterface{
												&components.ButtonSubmit{Label: "Save Role"},
												&components.ButtonModalForm{
													Label:       "Delete",
													Icon:        "trash",
													Name:        getters.Static("p_users.RoleDeleteForm"),
													Url:         lariv.RoutePath("p_users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
													FormPostURL: lariv.RoutePath("p_users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
													ModalUID:    "role-delete-modal",
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
		{Key: "p_users.RoleDetail", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.RoleDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.Detail[Role]{
					Getter: getters.Key[Role]("role"),
					Children: []components.PageInterface{
						&components.ContainerColumn{
							Children: []components.PageInterface{
								&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							},
						},
					},
				},
			},
		}},
		{Key: "p_users.RoleDeleteForm", Value: &components.Modal{
			UID: "role-delete-modal",
			Children: []components.PageInterface{
				&components.DeleteConfirmation{
					Title:   "Confirm Deletion",
					Message: "Are you sure you want to delete this role?",
					Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
				},
			},
		}},
	}
}
