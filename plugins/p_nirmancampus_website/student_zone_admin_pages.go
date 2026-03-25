package p_nirmancampus_website

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func getterNotIsLink() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		v, err := getters.GetterKey[bool]("$in.IsLink")(ctx)
		if err != nil {
			return true, nil
		}
		return !v, nil
	}
}

func init() {
	registerStudentZoneAdminMenuPages()
	registerStudentZoneAdminFilterPages()
	registerStudentZoneAdminFormPages()
	registerStudentZoneAdminTablePages()
	registerStudentZoneAdminDetailPages()
	registerStudentZoneAdminSelectionPages()
}

// --- Menus ---

func registerStudentZoneAdminMenuPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Student Zone"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Website"),
			Url:   lago.GetterRoutePath("nirmancampus_website.AppLandingRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Sections"),
				Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Items"),
				Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Section: %s", getters.GetterAny(getters.GetterKey[string]("section.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Sections"),
			Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Section Detail"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Section"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Section"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Item: %s", getters.GetterAny(getters.GetterKey[string]("item.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Items"),
			Url:   lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Item Detail"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Item"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Item"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerStudentZoneAdminFilterPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionFilter", &components.FormComponent[StudentZoneSection]{
		Url:    lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.GetterKey[string]("$get.Title"),
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

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionSelectionFilter", &components.FormComponent[StudentZoneSection]{
		Url:    lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.GetterKey[string]("$get.Title"),
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

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemFilter", &components.FormComponent[StudentZoneItem]{
		Url:    lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.GetterKey[string]("$get.Title"),
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
}

// --- Form Fields ---

func studentZoneAdminSectionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminSectionFormFieldsBody"},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Title"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Title",
								Name:     "Title",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Title"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Order"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:  "Order",
								Name:   "Order",
								Getter: getters.GetterKey[int]("$in.Order"),
							},
						},
					},
				},
			},
		},
	}
}

func studentZoneAdminItemFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminItemFormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.Title"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Title",
						Name:     "Title",
						Required: true,
						Getter:   getters.GetterKey[string]("$in.Title"),
					},
				},
			},

			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.StudentZoneSectionID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[StudentZoneSection]{
						Label:       "Section",
						Name:        "StudentZoneSectionID",
						Required:    true,
						Getter:      getters.GetterAssociation[StudentZoneSection](getters.GetterKey[uint]("$in.StudentZoneSectionID")),
						Url:         lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionSelectRoute", nil),
						Display:     getters.GetterKey[string]("$in.Title"),
						Placeholder: "Select a section...",
					},
				},
			},

			&components.ClientData{
				Data: "{ isLink: false }",
				Init: "isLink = $el.querySelector('[name=IsLink]')?.checked ?? false",
				Children: []components.PageInterface{
					&components.InputCheckbox{
						Label:  "Is Link",
						Name:   "IsLink",
						Getter: getters.GetterKey[bool]("$in.IsLink"),
						XModel: "isLink",
					},

					&components.ClientIf{
						Condition: "isLink",
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.GetterKey[error]("$error.Link"),
								Children: []components.PageInterface{
									&components.InputText{
										Label:  "Link URL",
										Name:   "Link",
										Getter: getters.GetterKey[string]("$in.Link"),
									},
								},
							},
						},
					},

					&components.ClientIf{
						Condition: "!isLink",
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.GetterKey[error]("$error.FileID"),
								Children: []components.PageInterface{
									&p_filesystem.InputVNode{
										Label: "File",
										Name:  "FileID",
										VNode: getters.GetterAssociation[p_filesystem.VNode](getters.GetterDeref(getters.GetterKey[*uint]("$in.FileID"))),
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerStudentZoneAdminFormPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionFormFields", studentZoneAdminSectionFormFields())
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemFormFields", studentZoneAdminItemFormFields())

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneSection]{
				Url:      lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Section",
				Subtitle: "Create a new student zone section",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					studentZoneAdminSectionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Section"},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneSection]{
				Getter: getters.GetterKey[StudentZoneSection]("section"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Section",
				Subtitle: "Update section details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					studentZoneAdminSectionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Update Section"},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneItem]{
				Url:      lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Item",
				Subtitle: "Create a new student zone item",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					studentZoneAdminItemFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Item"},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneItem]{
				Getter: getters.GetterKey[StudentZoneItem]("item"),
				Url: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Item",
				Subtitle: "Update item details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					studentZoneAdminItemFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Update Item"},
				},
			},
		},
	})
}

// --- Tables ---

func registerStudentZoneAdminTablePages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneSection]{
				UID:       "section-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[StudentZoneSection]]("sections"),
				CreateUrl: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionCreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Order",
						Name:  "Order",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Order")))},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneItem]{
				UID:       "item-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[StudentZoneItem]]("items"),
				CreateUrl: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemCreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminItemFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Section",
						Name:  "Section",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.StudentZoneSection.Title")},
						},
					},
					{
						Label: "Is Link",
						Name:  "IsLink",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.GetterKey[bool]("$row.IsLink")},
						},
					},
				},
			},
		},
	})
}

// --- Detail & Delete ---

func registerStudentZoneAdminDetailPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentZoneSection]{
				Getter: getters.GetterKey[StudentZoneSection]("section"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminSectionDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Title")},
							&components.LabelInline{
								Title: "Order",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$in.Order")))},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this section? All items in this section will also be deleted.",
				CancelUrl: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentZoneItem]{
				Getter: getters.GetterKey[StudentZoneItem]("item"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminItemDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Title")},
							&components.LabelInline{
								Title: "Section",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.StudentZoneSection.Title")},
								},
							},
							&components.ShowIf{
								Getter: getters.GetterAny(getters.GetterKey[bool]("$in.IsLink")),
								Children: []components.PageInterface{
									&components.LabelInline{
										Title: "Link",
										Children: []components.PageInterface{
											&components.FieldText{Getter: getters.GetterKey[string]("$in.Link")},
										},
									},
								},
							},
							&components.ShowIf{
								Getter: getterNotIsLink(),
								Children: []components.PageInterface{
									&components.LabelInline{
										Title: "File",
										Children: []components.PageInterface{
											&p_filesystem.FieldFile{
												VNode: getters.GetterAssociation[p_filesystem.VNode](getters.GetterDeref(getters.GetterKey[*uint]("$in.FileID"))),
											},
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

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this item?",
				CancelUrl: lago.GetterRoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Selection ---

func registerStudentZoneAdminSelectionPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionSelectionTable", &components.Modal{
		UID:   "section-selection-modal",
		Title: "Select Section",
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneSection]{
				UID:  "section-selection-table",
				Data: getters.GetterKey[components.ObjectList[StudentZoneSection]]("sections"),
				OnClick: getters.GetterSelect("StudentZoneSectionID",
					getters.GetterKey[uint]("$row.ID"),
					getters.GetterKey[string]("$row.Title"),
				),
				FilterComponent: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Order",
						Name:  "Order",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Order")))},
						},
					},
				},
			},
		},
	})
}
