package p_users

import (
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerAuthPages()
	registerSelectionPages()
	registerRolePages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("users.UserMenu", &components.SidebarMenu{
		Title: getters.Static("Users"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Users"),
				Url:   lago.RoutePath("users.ListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Roles"),
				Url:   lago.RoutePath("users.RoleListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("users.UserDetailMenu", &components.SidebarMenu{
		Title: getters.Format("User: %s", getters.Any(getters.Key[string]("user.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Users"),
			Url:   lago.RoutePath("users.ListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("User Detail"),
				Url: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit User"),
				Url: lago.RoutePath("users.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Delete User"),
				Url: lago.RoutePath("users.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Change Password"),
				Url: lago.RoutePath("users.ChangePasswordRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("user.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("users.UserSelfMenu", &components.SidebarMenu{
		Title: getters.Format("My account: %s", getters.Any(getters.Key[string]("user.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Home"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("My Profile"),
				Url:   lago.RoutePath("users.SelfDetailRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit My Profile"),
				Url:   lago.RoutePath("users.SelfUpdateRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Change Password"),
				Url:   lago.RoutePath("users.SelfChangePasswordRoute", nil),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("users.UserFilter", &components.FormComponent[User]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("users.ListRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
			&components.InputPhone{Label: "Phone", Name: "Phone", Getter: getters.Key[string]("$get.Phone")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.UserSelectionFilter", &components.FormComponent[User]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("users.SelectRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.Key[string]("$get.Email")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionFilter", &components.FormComponent[Role]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("users.SelectRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.Key[string]("$get.Name")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

// --- Form Fields & Forms ---

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
				Page:  components.Page{Key: "users.RoleField"},
				Error: getters.Key[error]("$error.RoleID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[Role]{
						Label:       "Role",
						Name:        "RoleID",
						Url:         lago.RoutePath("users.RoleSelectRoute", nil),
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
	// Remove the Role field — users should not edit their own role
	components.RemoveChild[*components.ContainerError](&fields, "users.RoleField")
	return fields
}

func registerFormPages() {
	lago.RegistryPage.Register("users.UserFormFields", userFormFields())

	lago.RegistryPage.Register("users.UserCreateForm", &components.Modal{
		Page: components.Page{
			Key: "users.UserCreateModal",
		},
		UID: "user-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				OnSubmit: getters.FormSubmitCloseModal(lago.RoutePath("users.CreateRoute", nil)),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("users.UserUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter: getters.Key[User]("user"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("users.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				})),
				Method:   http.MethodPost,
				Title:    "Edit User",
				Subtitle: "Update user details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					userFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save User"},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.SelfUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserSelfMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter:   getters.Key[User]("user"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("users.SelfUpdateRoute", nil)),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("users.SelfChangePasswordForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserSelfMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter:   getters.Key[User]("user"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("users.SelfChangePasswordRoute", nil)),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("users.ChangePasswordForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter:   getters.Key[User]("user"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("users.ChangePasswordRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})),
				Method:   http.MethodPost,
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
	})
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("users.UserTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[User]{
				UID:     "user-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[User]]("users"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.UserFilter"}},
					&components.ButtonModal{
						Url:     lago.RoutePath("users.CreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Email")},
					}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{
						&components.FieldPhone{Getter: getters.Key[string]("$row.Phone")},
					}},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("users.UserDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[User]{
				Page: components.Page{
					Key: "users.UserDetailContent",
				},
				Getter: getters.Key[User]("user"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Email")},
							&components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Superuser",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsSuperuser")},
								},
							},
							&components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.ForeignKey[Role, uint, string](getters.Key[uint]("$in.RoleID"), "Name")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.SelfDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserSelfMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[User]{
				Page: components.Page{
					Key: "users.SelfDetailContent",
				},
				Getter: getters.Key[User]("user"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Email")},
							&components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Page:  components.Page{Roles: []string{"superuser"}},
								Title: "Superuser",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsSuperuser")},
								},
							},
							&components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.ForeignKey[Role, uint, string](getters.Key[uint]("$in.RoleID"), "Name")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.UserDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this user?",
				CancelUrl: lago.RoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("user.ID"))}),
			},
		},
	})
}

// --- Auth (Login / Signup) ---

func registerAuthPages() {
	lago.RegistryPage.Register("users.LoginPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Children: []components.PageInterface{
				&components.FieldTitle{Getter: getters.Static("Login")},
				&components.FormComponent[User]{
					Page: components.Page{
						Key: "users.AuthForm",
					},
					Getter:   getters.Key[User]("user"),
					OnSubmit: getters.FormSubmit(lago.RoutePath("users.LoginRoute", nil)),
					Method:   http.MethodPost,
					ChildrenInput: []components.PageInterface{
						&components.ContainerError{
							Error: getters.Key[error]("$error.Email"),
							Children: []components.PageInterface{
								&components.InputEmail{
									Label:    "Email",
									Required: true,
									Getter:   getters.Key[string]("$in.Email"),
									Name:     "Email",
								},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.Password"),
							Children: []components.PageInterface{
								&components.InputPassword{
									Label:    "Password",
									Required: true,
									Name:     "Password",
								},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						&components.ButtonSubmit{
							Label:   "Login",
							Classes: "w-full",
						},
						&components.ButtonLink{
							Label:   "Don't have an account? Sign up",
							Link:    lago.RoutePath("users.SignupRoute", nil),
							Classes: "w-full",
						},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.SignupPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.Static("Create an Account")},
				&components.FormComponent[User]{
					Getter:   getters.Key[User]("user"),
					OnSubmit: getters.FormSubmit(lago.RoutePath("users.SignupRoute", nil)),
					Method:   http.MethodPost,
					ChildrenInput: []components.PageInterface{
						&components.ContainerError{
							Error: getters.Key[error]("$error.Name"),
							Children: []components.PageInterface{
								&components.InputText{Label: "Full Name", Required: true, Getter: getters.Key[string]("$in.Name"), Name: "Name"},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.Email"),
							Children: []components.PageInterface{
								&components.InputEmail{Label: "Email", Required: true, Getter: getters.Key[string]("$in.Email"), Name: "Email"},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.Phone"),
							Children: []components.PageInterface{
								&components.InputPhone{Label: "Phone Number", Required: true, Getter: getters.Key[string]("$in.Phone"), Name: "Phone"},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.password1"),
							Children: []components.PageInterface{
								&components.InputPassword{Name: "password1", Label: "Password", Required: true},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.password2"),
							Children: []components.PageInterface{
								&components.InputPassword{Name: "password2", Label: "Confirm Password", Required: true},
							},
						},
						&components.ContainerError{
							Error: getters.Key[error]("$error.terms_accepted"),
							Children: []components.PageInterface{
								&components.InputCheckbox{Name: "terms_accepted", Label: "I accept the terms and conditions", Getter: getters.Key[bool]("$in.terms_accepted"), Required: true},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						&components.ButtonSubmit{Label: "Sign Up", Classes: "w-full"},
						&components.ButtonLink{Label: "Already have an account? Login", Link: lago.RoutePath("users.LoginRoute", nil), Classes: "w-full"},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.UnauthenticatedPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Classes: "w-80 items-center text-center", Children: []components.PageInterface{
				&components.FieldTitle{Getter: getters.Static("Welcome")},
				&components.FieldSubtitle{Getter: getters.Static("Please log in or create an account to continue.")},
				&components.ContainerColumn{Classes: "w-full mt-4 gap-2", Children: []components.PageInterface{
					&components.ButtonLink{Label: "Login", Classes: "btn btn-primary text-white w-full", Link: lago.RoutePath("users.LoginRoute", nil)},
					&components.ButtonLink{Label: "Sign Up", Classes: "btn btn-outline w-full", Link: lago.RoutePath("users.SignupRoute", nil)},
				}},
			}},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("users.UserSelectionTable", &components.Modal{
		UID: "user-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[User]{
				UID:     "user-selection-table",
				Title:   "Select User",
				Data:    getters.Key[components.ObjectList[User]]("users"),
				OnClick: getters.Select("UserID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.UserSelectionFilter"}},
					&components.ButtonModal{
						Url:     lago.RoutePath("users.CreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
					{Label: "Email", Name: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Email")},
					}},
					{Label: "Phone", Name: "Phone", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionTable", &components.Modal{
		UID: "role-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Role]{
				UID:     "role-selection-table",
				Title:   "Select Role",
				Data:    getters.Key[components.ObjectList[Role]]("roles"),
				OnClick: getters.Select("RoleID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "users.RoleSelectionFilter"}},
					&components.ButtonModal{
						Url:     lago.RoutePath("users.RoleCreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				Columns: []components.TableColumn{
					{Label: "Name", Name: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.Key[string]("$row.Name")},
					}},
				},
			},
		},
	})
}

// --- Role CRUD Pages ---

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
			&components.SidebarMenuItem{
				Title: getters.Static("Delete Role"),
				Url:   lago.RoutePath("users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
			},
		},
	})

	// Role Filter
	lago.RegistryPage.Register("users.RoleFilter", &components.FormComponent[Role]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("users.RoleListRoute", nil)),
		Method:   http.MethodGet,
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
					&components.ButtonModal{
						Url:     lago.RoutePath("users.RoleCreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$row.ID"))})),
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
				OnSubmit: getters.FormSubmitCloseModal(lago.RoutePath("users.RoleCreateRoute", nil)),
				Method:   http.MethodPost,
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
			&components.FormComponent[Role]{
				Getter:   getters.Key[Role]("role"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))})),
				Method:   http.MethodPost,
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
					&components.ButtonSubmit{Label: "Save Role"},
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
	lago.RegistryPage.Register("users.RoleDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this role?",
				CancelUrl: lago.RoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("role.ID"))}),
			},
		},
	})
}
