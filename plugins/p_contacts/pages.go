package p_contacts

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
	registerSelectionPages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("contacts.ContactMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Contacts"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Contacts"),
				Url:   lago.GetterRoutePath("contacts.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Contact: %s", getters.GetterAny(getters.GetterKey[string]("contact.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Contacts"),
			Url:   lago.GetterRoutePath("contacts.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Contact Detail"),
				Url: lago.GetterRoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("contact.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Contact"),
				Url: lago.GetterRoutePath("contacts.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("contact.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Contact"),
				Url: lago.GetterRoutePath("contacts.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("contact.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("contacts.ContactFilter", &components.FormComponent[Contact]{
		Url:    lago.GetterRoutePath("contacts.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.GetterKey[string]("$get.Email"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply Filters"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactSelectionFilter", &components.FormComponent[Contact]{
		Url:    lago.GetterRoutePath("contacts.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.GetterKey[string]("$get.Email"),
			},
		},
		ChildrenAction: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex gap-2",
				Children: []components.PageInterface{
					&components.ButtonSubmit{Label: "Apply"},
					&components.ButtonClear{Label: "Clear"},
				},
			},
		},
	})
}

func contactFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "contacts.ContactFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Name"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Name",
						Name:     "Name",
						Required: true,
						Getter:   getters.GetterKey[string]("$in.Name"),
					},
				},
			},
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Phone"),
						Children: []components.PageInterface{
							&components.InputPhone{
								Label:  "Phone",
								Name:   "Phone",
								Getter: getters.GetterKey[string]("$in.Phone"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Email"),
						Children: []components.PageInterface{
							&components.InputEmail{
								Label:  "Email",
								Name:   "Email",
								Getter: getters.GetterKey[string]("$in.Email"),
							},
						},
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Address"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Address",
						Name:   "Address",
						Rows:   3,
						Getter: getters.GetterKey[string]("$in.Address"),
					},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Notes"),
				Children: []components.PageInterface{
					&components.InputTextarea{
						Label:  "Notes",
						Name:   "Notes",
						Rows:   4,
						Getter: getters.GetterKey[string]("$in.Notes"),
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
			&components.FormComponent[Contact]{
				Url:      lago.GetterRoutePath("contacts.CreateRoute", nil),
				Method:   http.MethodPost,
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
	})

	lago.RegistryPage.Register("contacts.ContactUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Contact]{
				Getter: getters.GetterKey[Contact]("contact"),
				Url: lago.GetterRoutePath("contacts.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Contact",
				Subtitle: "Update contact details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					contactFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Contact"},
				},
			},
		},
	})
}

func registerTablePages() {
	lago.RegistryPage.Register("contacts.ContactTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Contact]{
				Page:      components.Page{Key: "contacts.ContactTableBody"},
				UID:       "contact-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Contact]]("contacts"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "contacts.ContactFilter"}},
					&components.TableButtonCreate{Link: lago.GetterRoutePath("contacts.CreateRoute", nil)},
				},
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Email")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("contacts.ContactDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Contact]{
				Getter: getters.GetterKey[Contact]("contact"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "contacts.ContactDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.LabelInline{
								Title: "Phone",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Phone")},
								},
							},
							&components.LabelInline{
								Title: "Email",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Email")},
								},
							},
							&components.LabelInline{
								Title: "Address",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.GetterKey[string]("$in.Address"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.LabelInline{
								Title: "Notes",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.GetterKey[string]("$in.Notes"),
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

	lago.RegistryPage.Register("contacts.ContactDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this contact?",
				CancelUrl: lago.GetterRoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("contact.ID")),
				}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("contacts.ContactSelectionTable", &components.Modal{
		UID:   "contact-selection-modal",
		Title: "Select Contact",
		Children: []components.PageInterface{
			&components.DataTable[Contact]{
				Page:            components.Page{Key: "contacts.ContactSelectionTableBody"},
				UID:             "contact-selection-table",
				Data:            getters.GetterKey[components.ObjectList[Contact]]("contacts"),
				OnClick: getters.GetterSelect("ContactID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "contacts.ContactSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Phone")},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Email")},
						},
					},
				},
			},
		},
	})
}
