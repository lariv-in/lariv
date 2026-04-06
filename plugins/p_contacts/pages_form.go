package p_contacts

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func contactFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "contacts.ContactFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.Key[string]("$in.Name"),
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Phone"),
						Children: []components.PageInterface{
							&components.InputPhone{
								Label:  "Phone",
								Name:   "Phone",
								Getter: getters.Key[string]("$in.Phone"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputEmail{
								Label:  "Email",
								Name:   "Email",
								Getter: getters.Key[string]("$in.Email"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Address"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Address",
						Name:   "Address",
						Rows:   3,
						Getter: getters.Key[string]("$in.Address"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.Notes"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Notes",
						Name:   "Notes",
						Rows:   4,
						Getter: getters.Key[string]("$in.Notes"),
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("contacts.ContactCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("contacts.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Contact]{
						Attr: getters.FormBubbling(nil),

						Title:    "Create Contact",
						Subtitle: "Add a new contact",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							contactFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Contact"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("contacts.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Contact]{
						Getter: getters.Key[Contact]("contact"),
						Attr:   getters.FormBubbling(nil),

						Title:    "Edit Contact",
						Subtitle: "Update contact details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							contactFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ButtonModalForm{
										Label:       "Delete",
										Icon:        "trash",
										Url:         lago.RoutePath("contacts.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("contact.ID"))}),
										FormPostURL: lago.RoutePath("contacts.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("contact.ID"))}),
										ModalUID:    "contact-delete-modal",
										Classes:     "btn-outline btn-error btn-sm",
									},
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Contact"},
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
}
