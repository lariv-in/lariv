package p_users

import (
	"net/http"

	"github.com/lariv-in/components"
	"github.com/lariv-in/lago"
)

func init() {
	lago.RegistryPage.Register("users.LoginPage", components.LayoutAuthScaffold{
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
							Link:    lago.GetterRoute("users.SignupRoute"),
							Classes: "w-full",
						},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.SignupPage", components.LayoutAuthScaffold{
		Children: []components.PageInterface{
			components.ContainerColumn{Classes: "w-80", Children: []components.PageInterface{
				components.FieldTitle{Getter: components.GetterStatic("Create an Account")},
				components.FormComponent{
					Getter: components.GetterKey("user"),
					Url:    components.GetterNil(),
					Method: http.MethodPost,
					ChildrenInput: []components.PageInterface{
						components.ContainerError{
							Error: components.GetterKey("$error.name"),
							Children: []components.PageInterface{
								components.InputText{
									Label:    "Full Name",
									Required: true,
									Getter:   components.GetterKey("$in.name"),
									Name:     "name",
								},
							},
						},
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
							Error: components.GetterKey("$error.phone"),
							Children: []components.PageInterface{
								components.InputPhone{
									Label:    "Phone",
									Required: true,
									Getter:   components.GetterKey("$in.phone"),
									Name:     "phone",
								},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.password1"),
							Children: []components.PageInterface{
								components.InputPassword{
									Name:     "password1",
									Label:    "Password",
									Required: true,
								},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.password2"),
							Children: []components.PageInterface{
								components.InputPassword{
									Name:     "password2",
									Label:    "Confirm your password",
									Required: true,
								},
							},
						},
						components.ContainerError{
							Error: components.GetterKey("$error.terms_accepted"),
							Children: []components.PageInterface{
								components.InputCheckbox{
									Name:     "terms_accepted",
									Label:    "I accept the terms and conditions",
									Getter:   components.GetterKey("$in.terms_accepted"),
									Required: true,
								},
							},
						},
					},
					ChildrenAction: []components.PageInterface{
						components.ButtonSubmit{
							Label:   "Sign Up",
							Classes: "w-full",
						},
						components.ButtonLink{
							Label:   "Already have an account? Login",
							Link:    lago.GetterRoute("users.LoginRoute"),
							Classes: "w-full",
						},
					},
				},
			}},
		},
	})

	lago.RegistryPage.Register("users.UnauthenticatedPage", components.LayoutAuthScaffold{
		Children: []components.PageInterface{},
	})

	lago.RegistryPage.Register("users.AllUsersPage", components.LayoutTopbarScaffold{
		Children: []components.PageInterface{
			components.LayoutSimple{
				Children: []components.PageInterface{
					components.DataTable{
						Title:    "All Users",
						Subtitle: "Users registered in the system",
						Data:     components.GetterKey("users"),
						Columns: []components.TableColumn{
							{
								Label: "Name",
								Key:   "Name",
								Children: []components.PageInterface{
									components.FieldTitle{Getter: components.GetterKey("$row.Name")},
								},
							},
							{
								Label: "Email",
								Key:   "Email",
								Children: []components.PageInterface{
									components.FieldTitle{Getter: components.GetterKey("$row.Email")},
								},
							},
							{
								Label: "Phone",
								Key:   "Phone",
								Children: []components.PageInterface{
									components.FieldTitle{Getter: components.GetterKey("$row.Phone")},
								},
							},
							{
								Label: "Role",
								Key:   "Role.Name",
								Children: []components.PageInterface{
									components.FieldTitle{Getter: components.GetterKey("$row.Role.Name")},
								},
							},
							{
								Label: "Superuser",
								Key:   "IsSuperuser",
								Children: []components.PageInterface{
									components.FieldTitle{Getter: components.GetterFormat("%v", components.GetterKey("$row.IsSuperuser"))},
								},
							},
						},
					},
				},
			},
		},
	})
}
