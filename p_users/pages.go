package p_users

import (
	"net/http"

	"github.com/lariv-in/components"
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
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("users.UserMenu", components.SidebarMenu{
		Title: components.GetterStatic("Users"),
		Back: &components.SidebarMenuItem{
			Title: components.GetterStatic("Back to Home"),
			Url:   components.GetterStatic("/apps/"),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: components.GetterStatic("All Users"),
				Url:   components.GetterStatic(AppUrl),
			},
		},
	})

	lago.RegistryPage.Register("users.UserDetailMenu", components.SidebarMenu{
		Title: components.GetterFormat("User: %s", components.GetterKey("user.name")),
		Back: &components.SidebarMenuItem{
			Title: components.GetterStatic("Back to All Users"),
			Url:   components.GetterStatic(AppUrl),
		},
		Children: []components.PageInterface{
			components.SidebarMenuItem{
				Title: components.GetterStatic("User Detail"),
				Url:   components.GetterFormat(AppUrl+"%v/", components.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: components.GetterStatic("Edit User"),
				Url:   components.GetterFormat(AppUrl+"%v/edit/", components.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: components.GetterStatic("Delete User"),
				Url:   components.GetterFormat(AppUrl+"%v/delete/", components.GetterKey("user.id")),
			},
			components.SidebarMenuItem{
				Title: components.GetterStatic("Change Password"),
				Url:   components.GetterFormat(AppUrl+"%v/change-password/", components.GetterKey("user.id")),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("users.UserFilter", components.FormComponent{
		Url:    components.GetterStatic(AppUrl),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name"},
			components.InputText{Label: "Email", Name: "email"},
			components.InputPhone{Label: "Phone", Name: "phone"},
			components.InputTernary{
				Label:      "Active",
				Name:       "is_active",
				TrueLabel:  "Active Only",
				FalseLabel: "Inactive Only",
				NoneLabel:  "All",
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
		Url:    components.GetterStatic(AppUrl + "select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name"},
			components.InputText{Label: "Email", Name: "email"},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.UserMultiSelectionFilter", components.FormComponent{
		Url:    components.GetterStatic(AppUrl + "multi-select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name"},
			components.InputText{Label: "Email", Name: "email"},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				components.ButtonSubmit{Label: "Apply"},
				components.InputClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("users.RoleSelectionFilter", components.FormComponent{
		Url:    components.GetterStatic(AppUrl + "roles/select/"),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			components.InputText{Label: "Name", Name: "name"},
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
					components.InputText{Label: "Name", Name: "name", Required: true, Getter: components.GetterKey("$in.name")},
					components.InputEmail{Label: "Email", Name: "email", Required: true, Getter: components.GetterKey("$in.email")},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					components.InputPhone{Label: "Phone", Name: "phone", Required: true, Getter: components.GetterKey("$in.phone")},
					components.InputForeignKey{
						Label:       "Role",
						Name:        "role",
						Url:         components.GetterStatic(AppUrl + "roles/select/"),
						DisplayAttr: "name",
						Placeholder: "Select a role...",
						Required:    true,
						Getter:      components.GetterAssociation("roles", components.GetterKey("$in.role_id")),
					},
				},
			},
			components.InputTernary{
				Label:      "Active",
				Name:       "is_active",
				TrueLabel:  "Yes",
				FalseLabel: "No",
				NoneLabel:  "Not Set",
				Getter:     components.GetterKey("$in.IsActive"),
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
				Url:      components.GetterStatic(AppUrl + "create/"),
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
				Getter:   components.GetterKey("user"),
				Url:      components.GetterFormat(AppUrl+"%v/edit/", components.GetterKey("$in.id")),
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
				Getter:   components.GetterKey("user"),
				Url:      components.GetterFormat(AppUrl+"%v/change-password/", components.GetterKey("$in.id")),
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
				Data:            components.GetterKey("users"),
				CreateUrl:       components.GetterStatic(AppUrl + "create/"),
				OnClick:         components.GetterNavigate(AppUrl+"%v/", components.GetterKey("$row.id")),
				FilterComponent: lago.DynamicPage{Name: "users.UserFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Phone")},
					}},
					{Label: "Active", Key: "IsActive", Children: []components.PageInterface{
						components.FieldCheckbox{Getter: components.GetterKey("$row.IsActive")},
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
				Getter: components.GetterKey("user"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Children: []components.PageInterface{
							components.FieldTitle{Getter: components.GetterKey("$in.Name")},
							components.FieldSubtitle{Getter: components.GetterKey("$in.Email")},
							components.LabelInline{
								Title:   "Phone",
								Classes: "mt-2",
								Children: []components.PageInterface{
									components.FieldText{Getter: components.GetterKey("$in.Phone")},
								},
							},
							components.LabelInline{
								Title: "Active",
								Children: []components.PageInterface{
									components.FieldCheckbox{Getter: components.GetterKey("$in.IsActive")},
								},
							},
							components.LabelInline{
								Title: "Role",
								Children: []components.PageInterface{
									components.FieldText{Getter: components.GetterKey("$in.Role.Name")},
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
				CancelUrl: components.GetterFormat(AppUrl+"%v/", components.GetterKey("user.id")),
			},
		},
	})
}

// --- Auth (Login / Signup) ---

func registerAuthPages() {
	lago.RegistryPage.Register("users.LoginPage", components.ShellAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{Classes: "w-80", Children: []components.PageInterface{
				components.FieldTitle{Getter: components.GetterStatic("Login")},
				components.FormComponent{
					Getter: components.GetterKey("user"),
					Url:    components.GetterNil(),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						components.ContainerError{
							Error: components.GetterKey("$error.email"),
							Children: []components.PageInterface{
								components.InputEmail{
									Label:    "Email",
									Required: true,
									Getter:   components.GetterKey("$in.email"),
									Name:     "email",
								},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.password"),
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
				components.FieldTitle{Getter: components.GetterStatic("Create an Account")},
				components.FormComponent{
					Getter: components.GetterKey("user"),
					Url:    components.GetterNil(),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						components.ContainerError{
							Error: components.GetterKey("$error.name"),
							Children: []components.PageInterface{
								components.InputText{Label: "Full Name", Required: true, Getter: components.GetterKey("$in.name"), Name: "name"},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.email"),
							Children: []components.PageInterface{
								components.InputEmail{Label: "Email", Required: true, Getter: components.GetterKey("$in.email"), Name: "email"},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.phone"),
							Children: []components.PageInterface{
								components.InputPhone{Label: "Phone Number", Required: true, Getter: components.GetterKey("$in.phone"), Name: "phone"},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.password1"),
							Children: []components.PageInterface{
								components.InputPassword{Name: "password1", Label: "Password", Required: true},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.password2"),
							Children: []components.PageInterface{
								components.InputPassword{Name: "password2", Label: "Confirm Password", Required: true},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.terms_accepted"),
							Children: []components.PageInterface{
								components.InputCheckbox{Name: "terms_accepted", Label: "I accept the terms and conditions", Getter: components.GetterKey("$in.terms_accepted"), Required: true},
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
				components.FieldTitle{Getter: components.GetterStatic("Welcome")},
				components.FieldSubtitle{Getter: components.GetterStatic("Please log in or create an account to continue.")},
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
				Data:            components.GetterKey("users"),
				OnClick:         components.GetterSelect("user", components.GetterKey("$row.id"), components.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.UserSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Email")},
					}},
					{Label: "Phone", Key: "Phone", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Phone")},
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
				Data:            components.GetterKey("users"),
				OnClick:         components.GetterMultiSelect("role", components.GetterKey("$row.id"), components.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.UserMultiSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Name")},
					}},
					{Label: "Email", Key: "Email", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Email")},
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
				Data:            components.GetterKey("roles"),
				OnClick:         components.GetterSelect("role", components.GetterKey("$row.id"), components.GetterKey("$row.Name")),
				FilterComponent: lago.DynamicPage{Name: "users.RoleSelectionFilter"},
				Columns: []components.TableColumn{
					{Label: "Name", Key: "Name", Children: []components.PageInterface{
						components.FieldText{Getter: components.GetterKey("$row.Name")},
					}},
				},
			},
		},
	})
}
