package p_users

import (
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/getters"
	"github.com/lariv-in/lago"
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
		Title: getters.GetterStatic("Users"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Users"),
				Url:   lago.GetterRoutePath("users.ListRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Roles"),
				Url:   lago.GetterRoutePath("users.RoleListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("users.UserDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("User: %s", getters.GetterAny(getters.GetterKey[string]("user.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Users"),
			Url:   lago.GetterRoutePath("users.ListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("User Detail"),
				Url: lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit User"),
				Url: lago.GetterRoutePath("users.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete User"),
				Url: lago.GetterRoutePath("users.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("user.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Change Password"),
				Url: lago.GetterRoutePath("users.ChangePasswordRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("user.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("users.UserFilter", &components.FormComponent[User]{
		Url:    lago.GetterRoutePath("users.ListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.GetterKey[string]("$get.Email")},
			&components.InputPhone{Label: "Phone", Name: "Phone", Getter: getters.GetterKey[string]("$get.Phone")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.UserSelectionFilter", &components.FormComponent[User]{
		Url:    lago.GetterRoutePath("users.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
			&components.InputText{Label: "Email", Name: "Email", Getter: getters.GetterKey[string]("$get.Email")},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionFilter", &components.FormComponent[Role]{
		Url:    lago.GetterRoutePath("users.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
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
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.GetterKey[string]("$in.Name")},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputEmail{Label: "Email", Name: "Email", Required: true, Getter: getters.GetterKey[string]("$in.Email")},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Phone"),
						Children: []components.PageInterface{
							&components.InputPhone{Label: "Phone", Name: "Phone", Required: true, Getter: getters.GetterKey[string]("$in.Phone")},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.RoleID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[Role]{
								Label:       "Role",
								Name:        "RoleID",
								Url:         lago.GetterRoutePath("users.RoleSelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select a role...",
								Required:    true,
								Getter:      getters.GetterAssociation[Role](getters.GetterKey[uint]("$in.RoleID")),
							},
						},
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("users.UserFormFields", userFormFields())

	lago.RegistryPage.Register("users.UserCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Url:      lago.GetterRoutePath("users.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create User",
				Subtitle: "Create a new user",
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

	lago.RegistryPage.Register("users.UserUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter: getters.GetterKey[User]("user"),
				Url: lago.GetterRoutePath("users.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
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
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[User]{
				Getter: getters.GetterKey[User]("user"),
				Url:    lago.GetterRoutePath("users.SelfUpdateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Edit My Profile",
				Subtitle: "Update your account details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					userFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Profile"},
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
				Getter:   getters.GetterKey[User]("user"),
				Url:      lago.GetterRoutePath("users.ChangePasswordRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Change Password",
				Subtitle: "Update user password",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.new_password"),
						Children: []components.PageInterface{
							&components.InputPassword{Name: "new_password", Label: "New Password", Required: true},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.confirm_password"),
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
				UID:             "user-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[User]]("users"),
				CreateUrl:       lago.GetterRoutePath("users.CreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "users.UserFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						&components.FieldPhone{Getter: getters.GetterKey[string]("$row.Phone")},
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
				Getter: getters.GetterKey[User]("user"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Email")},
							&components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Superuser",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$in.IsSuperuser")},
								},
							},
							&components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterForeignKey[Role, uint, string](getters.GetterKey[uint]("$in.RoleID"), "Name")},
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
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[User]{
				Page: components.Page{
					Key: "users.SelfDetailContent",
				},
				Getter: getters.GetterKey[User]("user"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Email")},
							&components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Superuser",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$in.IsSuperuser")},
								},
							},
							&components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterForeignKey[Role, uint, string](getters.GetterKey[uint]("$in.RoleID"), "Name")},
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
				CancelUrl: lago.GetterRoutePath("users.DetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("user.ID"))}),
			},
		},
	})
}

// --- Auth (Login / Signup) ---

func registerAuthPages() {
	lago.RegistryPage.Register("users.LoginPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Classes: "w-80", Children: []components.PageInterface{
				&components.FieldTitle{Getter: getters.GetterStatic("Login")},
				&components.FormComponent[User]{
					Page: components.Page{
						Key: "users.AuthForm",
					},
					Getter: getters.GetterKey[User]("user"),
					Url:    getters.GetterNil[string](),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.Email"),
							Children: []components.PageInterface{
								&components.InputEmail{
									Label:    "Email",
									Required: true,
									Getter:   getters.GetterKey[string]("$in.Email"),
									Name:     "Email",
								},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.Password"),
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
							Link:    lago.GetterRoutePath("users.SignupRoute", nil),
							Classes: "w-full",
						},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.SignupPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Classes: "w-96", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Create an Account")},
				&components.FormComponent[User]{
					Getter: getters.GetterKey[User]("user"),
					Url:    getters.GetterNil[string](),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.Name"),
							Children: []components.PageInterface{
								&components.InputText{Label: "Full Name", Required: true, Getter: getters.GetterKey[string]("$in.Name"), Name: "Name"},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.Email"),
							Children: []components.PageInterface{
								&components.InputEmail{Label: "Email", Required: true, Getter: getters.GetterKey[string]("$in.Email"), Name: "Email"},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.Phone"),
							Children: []components.PageInterface{
								&components.InputPhone{Label: "Phone Number", Required: true, Getter: getters.GetterKey[string]("$in.Phone"), Name: "Phone"},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.password1"),
							Children: []components.PageInterface{
								&components.InputPassword{Name: "password1", Label: "Password", Required: true},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.password2"),
							Children: []components.PageInterface{
								&components.InputPassword{Name: "password2", Label: "Confirm Password", Required: true},
							},
						},
						&components.ContainerError{
							Error: getters.GetterKey[error]("$error.terms_accepted"),
							Children: []components.PageInterface{
								&components.InputCheckbox{Name: "terms_accepted", Label: "I accept the terms and conditions", Getter: getters.GetterKey[bool]("$in.terms_accepted"), Required: true},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						&components.ButtonSubmit{Label: "Sign Up", Classes: "w-full"},
						&components.ButtonLink{Label: "Already have an account? Login", Link: lago.GetterRoutePath("users.LoginRoute", nil), Classes: "w-full"},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.UnauthenticatedPage", &components.ShellAuthScaffold{
		Children: []components.PageInterface{
			&components.ContainerColumn{Classes: "w-80 items-center text-center", Children: []components.PageInterface{
				&components.FieldTitle{Getter: getters.GetterStatic("Welcome")},
				&components.FieldSubtitle{Getter: getters.GetterStatic("Please log in or create an account to continue.")},
				&components.ContainerColumn{Classes: "w-full mt-4 gap-2", Children: []components.PageInterface{
					&components.ButtonLink{Label: "Login", Classes: "btn btn-primary text-white w-full", Link: lago.GetterRoutePath("users.LoginRoute", nil)},
					&components.ButtonLink{Label: "Sign Up", Classes: "btn btn-outline w-full", Link: lago.GetterRoutePath("users.SignupRoute", nil)},
				}},
			}},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("users.UserSelectionTable", &components.Modal{
		UID:   "user-selection-modal",
		Title: "Select User",
		Children: []components.PageInterface{
			&components.DataTable[User]{
				UID:             "user-selection-table",
				Data:            getters.GetterKey[components.ObjectList[User]]("users"),
				OnClick:         getters.GetterSelect("UserID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.UserSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionTable", &components.Modal{
		UID:   "role-selection-modal",
		Title: "Select Role",
		Children: []components.PageInterface{
			&components.DataTable[Role]{
				UID:             "role-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Role]]("roles"),
				OnClick:         getters.GetterSelect("RoleID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.RoleSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
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
		Title: getters.GetterFormat("Role: %s", getters.GetterAny(getters.GetterKey[string]("role.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Roles"),
			Url:   lago.GetterRoutePath("users.RoleListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Role Detail"),
				Url:   lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("role.ID"))}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Role"),
				Url:   lago.GetterRoutePath("users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("role.ID"))}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Role"),
				Url:   lago.GetterRoutePath("users.RoleDeleteRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("role.ID"))}),
			},
		},
	})

	// Role Filter
	lago.RegistryPage.Register("users.RoleFilter", &components.FormComponent[Role]{
		Url:    lago.GetterRoutePath("users.RoleListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{Label: "Name", Name: "Name", Getter: getters.GetterKey[string]("$get.Name")},
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
				UID:             "role-table",
				Classes:         "w-full",
				Data:            getters.GetterKey[components.ObjectList[Role]]("roles"),
				CreateUrl:       lago.GetterRoutePath("users.RoleCreateRoute", nil),
				OnClick:         getters.GetterNavigateGetter(lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID"))})),
				FilterComponent: lago.DynamicPage{Name: "users.RoleFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
					}},
				},
			},
		},
	})

	// Role Create Form
	lago.RegistryPage.Register("users.RoleCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Role]{
				Url:      lago.GetterRoutePath("users.RoleCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Role",
				Subtitle: "Create a new role",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.GetterKey[string]("$in.Name")},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Role"},
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
				Getter:   getters.GetterKey[Role]("role"),
				Url:      lago.GetterRoutePath("users.RoleUpdateRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID"))}),
				Method:   http.MethodPost,
				Title:    "Edit Role",
				Subtitle: "Update role details",
				ChildrenInput: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{Label: "Name", Name: "Name", Required: true, Getter: getters.GetterKey[string]("$in.Name")},
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
				Getter: getters.GetterKey[Role]("role"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
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
				CancelUrl: lago.GetterRoutePath("users.RoleDetailRoute", map[string]getters.Getter[any]{"id": getters.GetterAny(getters.GetterKey[uint]("role.ID"))}),
			},
		},
	})
}
