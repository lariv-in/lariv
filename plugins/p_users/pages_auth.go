package p_users

import (
	"github.com/lariv-in/lariv"
	"github.com/lariv-in/lariv/components"
	"github.com/lariv-in/lariv/getters"
	"github.com/lariv-in/lariv/registry"
)

func pageEntriesAuth() []registry.Pair[string, components.PageInterface] {
	return []registry.Pair[string, components.PageInterface]{
		{Key: "p_users.LoginPage", Value: &components.ShellAuthScaffold{
			Children: []components.PageInterface{
				&components.ContainerColumn{Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("Login")},
					&components.FormListenBoostedPost{
						Name:      getters.Static("p_users.LoginPage"),
						ActionURL: lariv.RoutePath("p_users.LoginRoute", nil),
						Children: []components.PageInterface{
							&components.FormComponent[User]{
								Page: components.Page{
									Key: "p_users.AuthForm",
								},
								Getter: getters.Key[User]("user"),
								Attr:   getters.FormBubbling(getters.Static("p_users.LoginPage")),
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
										Classes: "w-full mb-4",
									},
									&components.ButtonLink{
										Page:    components.Page{Key: "p_users.AuthSignupLink"},
										Label:   getters.Static("Don't have an account? Sign up"),
										Link:    lariv.RoutePath("p_users.SignupRoute", nil),
										Classes: "w-full",
									},
								},
							},
						},
					},
				}},
			},
		}},
		{Key: "p_users.SignupPage", Value: &components.ShellAuthScaffold{
			Children: []components.PageInterface{
				&components.ContainerColumn{Children: []components.PageInterface{
					components.FieldTitle{Getter: getters.Static("Create an Account")},
					&components.FormListenBoostedPost{
						Name:      getters.Static("p_users.SignupPage"),
						ActionURL: lariv.RoutePath("p_users.SignupRoute", nil),
						Children: []components.PageInterface{
							&components.FormComponent[User]{
								Getter: getters.Key[User]("user"),
								Attr:   getters.FormBubbling(getters.Static("p_users.SignupPage")),
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
									&components.ButtonLink{Label: getters.Static("Already have an account? Login"), Link: lariv.RoutePath("p_users.LoginRoute", nil), Classes: "w-full"},
								},
							},
						},
					},
				}},
			},
		}},
		{Key: "p_users.UnauthenticatedPage", Value: &components.ShellAuthScaffold{
			Children: []components.PageInterface{
				&components.ContainerColumn{Classes: "w-80 items-center text-center", Children: []components.PageInterface{
					&components.FieldTitle{Getter: getters.Static("Welcome")},
					&components.FieldSubtitle{Getter: getters.Static("Please log in or create an account to continue.")},
					&components.ContainerColumn{Classes: "w-full mt-4 gap-2", Children: []components.PageInterface{
						&components.ButtonLink{Label: getters.Static("Login"), Classes: "btn btn-primary text-white w-full", Link: lariv.RoutePath("p_users.LoginRoute", nil)},
						&components.ButtonLink{
							Page:    components.Page{Key: "p_users.AuthSignupLink"},
							Label:   getters.Static("Sign Up"),
							Classes: "btn btn-outline w-full",
							Link:    lariv.RoutePath("p_users.SignupRoute", nil),
						},
					}},
				}},
			},
		}},
	}
}
