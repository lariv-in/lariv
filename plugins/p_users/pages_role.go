package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerRolePages() {
	// Role Menu
	lago.RegistryPage.Register("users.RoleDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Role: %s", getters.Any(getters.Key[string]("role.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Roles"),
			Url:   lago.RoutePath("users.RoleListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Role Detail"),
				Url:   lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Role"),
				Url:   lago.RoutePath("users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
			},
		},
	})

	// Role Filter
	lago.RegistryPage.Register("users.RoleFilter", &components.FormComponent[Role]{
		Attr: getters.FormBoostedGet(lago.RoutePath("users.RoleListRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	// Role Table
	lago.RegistryPage.Register("users.RoleTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Role]{
				UID:     "role-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Role]]("roles"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.RoleFilter"}},
					&components.ButtonModalForm{
						Name:        getters.Static("users.RoleCreateForm"),
						Url:         lago.RoutePath("users.RoleCreateRoute", nil),
						FormPostURL: lago.RoutePath("users.RoleCreateRoute", nil),
						ModalUID:    "role-create-modal",
						Icon:        "plus",
						Classes:     "btn-square btn-outline btn-sm",
					},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
				},
			},
		},
	})

	// Role Create Form
	lago.RegistryPage.Register("users.RoleCreateForm", &components.Modal{
		Page: components.Page{
			Key: "users.RoleCreateModal",
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
	})

	// Role Update Form
	lago.RegistryPage.Register("users.RoleUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("users.RoleUpdateForm"),
				ActionURL: lago.RoutePath("users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
				Children: []components.PageInterface{
					&components.FormComponent[Role]{
						Getter: getters.Key[Role]("role"),
						Attr:   getters.FormBubbling(getters.Static("users.RoleUpdateForm")),

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
												Name:        getters.Static("users.RoleDeleteForm"),
												Url:         lago.RoutePath("users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
												FormPostURL: lago.RoutePath("users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
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
	})

	// Role Detail
	lago.RegistryPage.Register("users.RoleDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
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
	})

	// Role Delete
	lago.RegistryPage.Register("users.RoleDeleteForm", &components.Modal{
		UID: "role-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this role?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
