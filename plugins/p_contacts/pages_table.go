package p_contacts

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

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
