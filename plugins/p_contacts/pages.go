package p_contacts

import (

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
		Title: getters.Static("Contacts"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Contacts"),
				Url:   lago.RoutePath("contacts.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("contacts.ContactDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Contact: %s", getters.Any(getters.Key[string]("contact.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Contacts"),
			Url:   lago.RoutePath("contacts.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Contact Detail"),
				Url: lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Contact"),
				Url: lago.RoutePath("contacts.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("contact.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("contacts.ContactFilter", &components.FormComponent[Contact]{
		Attr: getters.FormBoostedGet(lago.RoutePath("contacts.DefaultRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$get.Email"),
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
		Attr: getters.FormBoostedGet(lago.RoutePath("contacts.SelectRoute", nil)),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Email",
				Name:   "Email",
				Getter: getters.Key[string]("$get.Email"),
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
				Attr: getters.FormBubbling(nil),


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

func registerTablePages() {
	lago.RegistryPage.Register("contacts.ContactTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "contacts.ContactMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Contact]{
				Page:    components.Page{Key: "contacts.ContactTableBody"},
				UID:     "contact-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Contact]]("contacts"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "contacts.ContactFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("contacts.CreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("contacts.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Email")},
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
				Attr: getters.FormBubbling(nil),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("contacts.ContactSelectionTable", &components.Modal{
		UID: "contact-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Contact]{
				Page:    components.Page{Key: "contacts.ContactSelectionTableBody"},
				UID:     "contact-selection-table",
				Title:   "Select Contact",
				Data:    getters.Key[components.ObjectList[Contact]]("contacts"),
				RowAttr: getters.RowAttrSelect("ContactID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "contacts.ContactSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Name")},
						},
					},
					{
						Label: "Phone",
						Name:  "Phone",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Phone")},
						},
					},
					{
						Label: "Email",
						Name:  "Email",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Email")},
						},
					},
				},
			},
		},
	})
}
