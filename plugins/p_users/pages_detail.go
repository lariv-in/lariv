package p_users

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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

	lago.RegistryPage.Register("users.UserDeleteForm", &components.Modal{
		UID: "user-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this user?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}

// --- Auth (Login / Signup) ---
