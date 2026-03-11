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
	lago.RegistryPage.Register("users.UserMenu", components.SidebarMenu{
		Title: getters.GetterStatic("Users"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Home"),
			Url:   getters.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("All Users"),
				Url:   getters.GetterStatic(AppUrl),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Roles"),
				Url:   getters.GetterStatic(RoleUrl),
			},
		},
	})

	lago.RegistryPage.Register("users.UserDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("User: %s", getters.GetterKey("user.name")),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Users"),
			Url:   getters.GetterStatic(AppUrl),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("User Detail"),
				Url:   getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit User"),
				Url:   getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete User"),
				Url:   getters.GetterFormat(AppUrl+"%v/delete/", getters.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Change Password"),
				Url:   getters.GetterFormat(AppUrl+"%v/change-password/", getters.GetterKey("user.id")),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("users.UserFilter", components.FormComponent{
		Url:    getters.GetterStatic(AppUrl),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
			components.InputText{Label: "Email", Name: "email", Getter: getters.GetterKey("$get.email")},
			components.InputPhone{Label: "Phone", Name: "phone", Getter: getters.GetterKey("$get.phone")},
			components.InputTernary{
				Label:      "Superuser",
				Name:       "is_superuser",
				TrueLabel:  "Yes",
				FalseLabel: "No",
				NoneLabel:  "All",
				Getter:     getters.GetterKey("$get.is_superuser"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.UserSelectionFilter", components.FormComponent{
		Url:    getters.GetterStatic(AppUrl + "select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
			components.InputText{Label: "Email", Name: "email", Getter: getters.GetterKey("$get.email")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.UserMultiSelectionFilter", components.FormComponent{
		Url:    getters.GetterStatic(AppUrl + "multi-select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
			components.InputText{Label: "Email", Name: "email", Getter: getters.GetterKey("$get.email")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionFilter", components.FormComponent{
		Url:    getters.GetterStatic(RoleUrl + "select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})
}

// --- Form Fields & Forms ---

func userFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					components.InputText{Label: "Name", Name: "name", Required: true, Getter: getters.GetterKey("$in.name")},
					components.InputEmail{Label: "Email", Name: "email", Required: true, Getter: getters.GetterKey("$in.email")},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					components.InputPhone{Label: "Phone", Name: "phone", Required: true, Getter: getters.GetterKey("$in.phone")},
					components.InputForeignKey{
						Label:       "Role",
						Name:        "role_id",
						Url:         getters.GetterStatic(RoleUrl + "select/"),
						DisplayAttr: "name",
						Placeholder: "Select a role...",
						Required:    true,
						Getter:      getters.GetterAssociation("roles", getters.GetterKey("$in.role_id")),
					},
				},
			},
			components.InputTernary{
				Label:      "Superuser",
				Name:       "is_superuser",
				TrueLabel:  "Yes",
				FalseLabel: "No",
				NoneLabel:  "Not Set",
				Getter:     getters.GetterKey("$in.IsSuperuser"),
			},
			components.ButtonSubmit{Label: "Save User"},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("users.UserFormFields", userFormFields())

	lago.RegistryPage.Register("users.UserCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:      getters.GetterStatic(AppUrl + "create/"),
				Method:   http.MethodPost,
				Title:    "Create User",
				Subtitle: "Create a new user",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					userFormFields(),
				},
			},
		},
	})

	lago.RegistryPage.Register("users.UserUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Getter:   getters.GetterKey("user"),
				Url:      getters.GetterFormat(AppUrl+"%v/edit/", getters.GetterKey("$in.id")),
				Method:   http.MethodPost,
				Title:    "Edit User",
				Subtitle: "Update user details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					userFormFields(),
				},
			},
		},
	})

	lago.RegistryPage.Register("users.ChangePasswordForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Getter:   getters.GetterKey("user"),
				Url:      getters.GetterFormat(AppUrl+"%v/change-password/", getters.GetterKey("$in.id")),
				Method:   http.MethodPost,
				Title:    "Change Password",
				Subtitle: "Update user password",
				ChildrenInput: []components.PageInterface{
					components.InputPassword{Name: "new_password", Label: "New Password", Required: true},
					components.InputPassword{Name: "confirm_password", Label: "Confirm New Password", Required: true},
				},
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Change Password"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("users.UserTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "user-table",
				Classes:         "w-full",
				Data:            getters.GetterKey("users"),
				CreateUrl:       getters.GetterStatic(AppUrl + "create/"),
				OnClick:         getters.GetterNavigate(AppUrl+"%v/", getters.GetterKey("$row.ID")),
				FilterComponent: lago.DynamicPage{Name: "users.UserFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Phone")},
					}},
					{Label: "Superuser", Key: "IsSuperuser", Children: []components.PageInterface{
						components.FieldCheckbox{Getter: getters.GetterKey("$row.IsSuperuser")},
					}},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("users.UserDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			components.Detail{
				Getter: getters.GetterKey("user"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Children: []components.PageInterface{
							components.FieldTitle{Getter: getters.GetterKey("$in.Name")},
							components.FieldSubtitle{Getter: getters.GetterKey("$in.Email")},
							components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									components.FieldText{Getter: getters.GetterKey("$in.Phone")},
								},
							},
							components.LabelInline{
								Title: "Superuser",
								Children: []components.PageInterface{
									components.FieldCheckbox{Getter: getters.GetterKey("$in.IsSuperuser")},
								},
							},
							components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									components.FieldText{Getter: getters.GetterForeignKey[Role](getters.GetterKey("$in.RoleID"), "Name")},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.UserDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserDetailMenu"},
		},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this user?",
				CancelUrl: getters.GetterFormat(AppUrl+"%v/", getters.GetterKey("user.id")),
			},
		},
	})
}

// --- Auth (Login / Signup) ---

func registerAuthPages() {
	lago.RegistryPage.Register("users.LoginPage", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{Classes: "w-80", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Login")},
				components.FormComponent{
					Getter: getters.GetterKey("user"),
					Url:    getters.GetterNil(),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						components.ContainerError{
							Error: getters.GetterKey("$error.email"),
							Children: []components.PageInterface{
								components.InputEmail{
									Label:    "Email",
									Required: true,
									Getter:   getters.GetterKey("$in.email"),
									Name:     "email",
								},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.password"),
							Children: []components.PageInterface{
								components.InputPassword{
									Label:    "Password",
									Required: true,
									Name:     "password",
								},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						components.ButtonSubmit{
							Label:   "Login",
							Classes: "w-full",
						},
						components.ButtonLink{
							Label:   "Don't have an account? Sign up",
							Link:    lago.RoutePathGetter("users.SignupRoute"),
							Classes: "w-full",
						},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.SignupPage", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{Classes: "w-96", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Create an Account")},
				components.FormComponent{
					Getter: getters.GetterKey("user"),
					Url:    getters.GetterNil(),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						components.ContainerError{
							Error: getters.GetterKey("$error.name"),
							Children: []components.PageInterface{
								components.InputText{Label: "Full Name", Required: true, Getter: getters.GetterKey("$in.name"), Name: "name"},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.email"),
							Children: []components.PageInterface{
								components.InputEmail{Label: "Email", Required: true, Getter: getters.GetterKey("$in.email"), Name: "email"},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.phone"),
							Children: []components.PageInterface{
								components.InputPhone{Label: "Phone Number", Required: true, Getter: getters.GetterKey("$in.phone"), Name: "phone"},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.password1"),
							Children: []components.PageInterface{
								components.InputPassword{Name: "password1", Label: "Password", Required: true},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.password2"),
							Children: []components.PageInterface{
								components.InputPassword{Name: "password2", Label: "Confirm Password", Required: true},
							},
						},
						components.ContainerError{
							Error: getters.GetterKey("$error.terms_accepted"),
							Children: []components.PageInterface{
								components.InputCheckbox{Name: "terms_accepted", Label: "I accept the terms and conditions", Getter: getters.GetterKey("$in.terms_accepted"), Required: true},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						components.ButtonSubmit{Label: "Sign Up", Classes: "w-full"},
						components.ButtonLink{Label: "Already have an account? Login", Link: lago.RoutePathGetter("users.LoginRoute"), Classes: "w-full"},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.UnauthenticatedPage", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{Classes: "w-80 items-center text-center", Children: []components.PageInterface{
				components.FieldTitle{Getter: getters.GetterStatic("Welcome")},
				components.FieldSubtitle{Getter: getters.GetterStatic("Please log in or create an account to continue.")},
				components.ContainerColumn{Classes: "w-full mt-4 gap-2", Children: []components.PageInterface{
					components.ButtonLink{Label: "Login", Classes: "btn btn-primary text-white w-full", Link: lago.RoutePathGetter("users.LoginRoute")},
					components.ButtonLink{Label: "Sign Up", Classes: "btn btn-outline w-full", Link: lago.RoutePathGetter("users.SignupRoute")},
				}},
			}},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("users.UserSelectionTable", components.Modal{
		UID:   "user-selection-modal",
		Title: "Select User",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "user-selection-table",
				Data:            getters.GetterKey("users"),
				OnClick:         getters.GetterSelect("user", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.UserSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Phone")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.UserMultiSelectionTable", components.Modal{
		UID:   "user-multi-selection-modal",
		Title: "Select Users",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "user-multi-selection-table",
				Data:            getters.GetterKey("users"),
				OnClick:         getters.GetterMultiSelect("role", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.UserMultiSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Email")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.RoleMultiSelectionTable", components.Modal{
		UID:   "role-multi-selection-modal",
		Title: "Select Roles",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "role-multi-selection-table",
				Data:            getters.GetterKey("roles"),
				OnClick:         getters.GetterMultiSelect("role", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.RoleSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
				},
			},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionTable", components.Modal{
		UID:   "role-selection-modal",
		Title: "Select Role",
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "role-selection-table",
				Data:            getters.GetterKey("roles"),
				OnClick:         getters.GetterSelect("role_id", getters.GetterKey("$row.ID"), getters.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.RoleSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
				},
			},
		},
	})
}

// --- Role CRUD Pages ---

func registerRolePages() {
	// Role Menu
	lago.RegistryPage.Register("users.RoleDetailMenu", components.SidebarMenu{
		Title: getters.GetterFormat("Role: %s", getters.GetterKey("role.name")),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Roles"),
			Url:   getters.GetterStatic(RoleUrl),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Role Detail"),
				Url:   getters.GetterFormat(RoleUrl+"%v/", getters.GetterKey("role.id")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Role"),
				Url:   getters.GetterFormat(RoleUrl+"%v/edit/", getters.GetterKey("role.id")),
			},
			components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Role"),
				Url:   getters.GetterFormat(RoleUrl+"%v/delete/", getters.GetterKey("role.id")),
			},
		},
	})

	// Role Filter
	lago.RegistryPage.Register("users.RoleFilter", components.FormComponent{
		Url:    getters.GetterStatic(RoleUrl),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name", Getter: getters.GetterKey("$get.name")},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply Filters"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	// Role Table
	lago.RegistryPage.Register("users.RoleTable", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			components.DataTable{
				UID:             "role-table",
				Classes:         "w-full",
				Data:            getters.GetterKey("roles"),
				CreateUrl:       lago.RoutePathGetter("users.RoleCreateRoute"),
				OnClick:         getters.GetterNavigate(RoleUrl+"%v/", getters.GetterKey("$row.ID")),
				FilterComponent: lago.DynamicPage{Name: "users.RoleFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: getters.GetterKey("$row.Name")},
					}},
				},
			},
		},
	})

	// Role Create Form
	lago.RegistryPage.Register("users.RoleCreateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.UserMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Url:      getters.GetterStatic(RoleUrl + "create/"),
				Method:   http.MethodPost,
				Title:    "Create Role",
				Subtitle: "Create a new role",
				ChildrenInput: []components.PageInterface{
					components.InputText{Label: "Name", Name: "name", Required: true, Getter: getters.GetterKey("$in.name")},
				},
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Save Role"},
				},
			},
		},
	})

	// Role Update Form
	lago.RegistryPage.Register("users.RoleUpdateForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
		},
		Children: []components.PageInterface{
			components.FormComponent{
				Getter:   getters.GetterKey("role"),
				Url:      getters.GetterFormat(RoleUrl+"%v/edit/", getters.GetterKey("$in.id")),
				Method:   http.MethodPost,
				Title:    "Edit Role",
				Subtitle: "Update role details",
				ChildrenInput: []components.PageInterface{
					components.InputText{Label: "Name", Name: "name", Required: true, Getter: getters.GetterKey("$in.name")},
				},
				ChildrenAction: []components.PageInterface{
					components.ButtonSubmit{Label: "Save Role"},
				},
			},
		},
	})

	// Role Detail
	lago.RegistryPage.Register("users.RoleDetail", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
		},
		Children: []components.PageInterface{
			components.Detail{
				Getter: getters.GetterKey("role"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Children: []components.PageInterface{
							components.FieldTitle{Getter: getters.GetterKey("$in.Name")},
						},
					},
				},
			},
		},
	})

	// Role Delete
	lago.RegistryPage.Register("users.RoleDeleteForm", components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "users.RoleDetailMenu"},
		},
		Children: []components.PageInterface{
			components.DeleteConfirmation{
				Title:     "Confirm Deletion",
				Message:   "Are you sure you want to delete this role?",
				CancelUrl: getters.GetterFormat(RoleUrl+"%v/", getters.GetterKey("role.id")),
			},
		},
	})
}
