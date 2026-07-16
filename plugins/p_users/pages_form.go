package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func userFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.Key[string]("$in.Name")},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputEmail{Label: "Email", Name: "Email", Required: true, Getter: getters.Key[string]("$in.Email")},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Phone"),
				Children: []components.PageInterface{
					&components.InputPhone{Label: "Phone", Name: "Phone", Required: true, Getter: getters.Key[string]("$in.Phone")},
				},
			},
			&components.ContainerError{
				Page:  components.Page{Key: "p_users.RoleField"},
				Error: getters.Key[error]("$error.RoleID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[Role]{
						Label:       "Role",
						Name:        "RoleID",
						Url:         lariv.RoutePath("p_users.RoleSelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select a role...",
						Required:    true,
						Getter:      getters.Association[Role](getters.Key[uint]("$in.RoleID")),
					},
				},
			},
		},
	}
}

func selfFormFields() components.ContainerColumn {
	fields := userFormFields()
	components.RemoveChild[*components.ContainerError](&fields, "p_users.RoleField")
	return fields
}

func pageEntriesForms() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.UserFormFields", Value: userFormFields()},
		{Key: "p_users.UserCreateForm", Value: &components.Modal{
			Page: components.Page{
				Key: "p_users.UserCreateModal",
			},
			UID: "user-create-modal",
			Children: []components.PageInterface{
				&components.FormComponent[User]{
					Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

					Title:    "Create User",
					Subtitle: "Create a new user",
					Classes:  "@container",
					ChildrenInput: []components.PageInterface{
						userFormFields(),
					},
					ChildrenAction: []components.PageInterface{
						&components.ContainerRow{
							Classes: "flex justify-end gap-2 mt-2",
							Children: []components.PageInterface{
								&components.ButtonSubmit{Label: "Save User", Classes: "btn-primary"},
							},
						},
					},
				},
			},
		}},
		{Key: "p_users.UserUpdateForm", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name: getters.Static("p_users.UserUpdateForm"),
					ActionURL: lariv.RoutePath("p_users.UpdateRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("user.ID")),
					}),
					Children: []components.PageInterface{
						&components.FormComponent[User]{
							Getter: getters.Key[User]("user"),
							Attr:   getters.FormBubbling(getters.Static("p_users.UserUpdateForm")),

							Title:    "Edit User",
							Subtitle: "Update user details",
							Classes:  "@container",
							ChildrenInput: []components.PageInterface{
								userFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ContainerRow{
									Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
									Children: []components.PageInterface{
										&components.ContainerRow{
											Classes: "flex justify-end gap-2",
											Children: []components.PageInterface{
												&components.ButtonSubmit{Label: "Save User"},
												&components.ButtonModalForm{
													Label:       "Delete",
													Icon:        "trash",
													Name:        getters.Static("p_users.UserDeleteForm"),
													Url:         lariv.RoutePath("p_users.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("user.ID"))}),
													FormPostURL: lariv.RoutePath("p_users.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("user.ID"))}),
													ModalUID:    "user-delete-modal",
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
		{Key: "p_users.SelfUpdateForm", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserSelfMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_users.SelfUpdateForm"),
					ActionURL: lariv.RoutePath("p_users.SelfUpdateRoute", nil),
					Children: []components.PageInterface{
						&components.FormComponent[User]{
							Getter: getters.Key[User]("user"),
							Attr:   getters.FormBubbling(getters.Static("p_users.SelfUpdateForm")),

							Title:    "Edit My Profile",
							Subtitle: "Update your account details",
							Classes:  "@container",
							ChildrenInput: []components.PageInterface{
								selfFormFields(),
							},
							ChildrenAction: []components.PageInterface{
								&components.ButtonSubmit{Label: "Save Profile"},
							},
						},
					},
				},
			},
		}},
		{Key: "p_users.SelfChangePasswordForm", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserSelfMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_users.SelfChangePasswordForm"),
					ActionURL: lariv.RoutePath("p_users.SelfChangePasswordRoute", nil),
					Children: []components.PageInterface{
						&components.FormComponent[User]{
							Getter: getters.Key[User]("user"),
							Attr:   getters.FormBubbling(getters.Static("p_users.SelfChangePasswordForm")),

							Title:    "Change Password",
							Subtitle: "Update your password",
							ChildrenInput: []components.PageInterface{
								&components.ContainerError{
									Error: getters.Key[error]("$error.new_password"),
									Children: []components.PageInterface{
										&components.InputPassword{Name: "new_password", Label: "New Password", Required: true},
									},
								},
								&components.ContainerError{
									Error: getters.Key[error]("$error.confirm_password"),
									Children: []components.PageInterface{
										&components.InputPassword{Name: "confirm_password", Label: "Confirm New Password", Required: true},
									},
								},
							},
							ChildrenAction: []components.PageInterface{
								&components.ButtonSubmit{Label: "Change Password"},
							},
						},
					},
				},
			},
		}},
		{Key: "p_users.ChangePasswordForm", Value: &components.ShellScaffold{
			Sidebar: []components.PageInterface{
				lariv.DynamicPage{Name: "p_users.UserDetailMenu"},
			},
			Children: []components.PageInterface{
				&components.FormListenBoostedPost{
					Name:      getters.Static("p_users.ChangePasswordForm"),
					ActionURL: lariv.RoutePath("p_users.ChangePasswordRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("user.ID"))}),
					Children: []components.PageInterface{
						&components.FormComponent[User]{
							Getter: getters.Key[User]("user"),
							Attr:   getters.FormBubbling(getters.Static("p_users.ChangePasswordForm")),

							Title:    "Change Password",
							Subtitle: "Update user password",
							ChildrenInput: []components.PageInterface{
								&components.ContainerError{
									Error: getters.Key[error]("$error.new_password"),
									Children: []components.PageInterface{
										&components.InputPassword{Name: "new_password", Label: "New Password", Required: true},
									},
								},
								&components.ContainerError{
									Error: getters.Key[error]("$error.confirm_password"),
									Children: []components.PageInterface{
										&components.InputPassword{Name: "confirm_password", Label: "Confirm New Password", Required: true},
									},
								},
							},
							ChildrenAction: []components.PageInterface{
								&components.ButtonSubmit{Label: "Change Password"},
							},
						},
					},
				},
			},
		}},
	}
}
