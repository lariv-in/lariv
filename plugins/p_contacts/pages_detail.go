package p_contacts

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerDetailPages() {
	lago.RegistryPage.Register("contacts.ContactDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Contact]{
				Getter: getters.Key[Contact]("contact"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "contacts.ContactDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Email")},
								},
							},
							&components.LabelInline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.Key[string]("$in.Address"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.LabelInline{
								Title: "Notes",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.Key[string]("$in.Notes"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactDeleteForm", &components.Modal{
		UID: "contact-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this contact?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
