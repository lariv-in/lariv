package p_nirmancampus_student_zone

import (
	"context"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/p_filesystem"
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
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("student_zone.Menu", &components.SidebarMenu{
		Title: getters.GetterStatic("Student Zone"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Sections"),
				Url:   lago.GetterRoutePath("student_zone.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Items"),
				Url:   lago.GetterRoutePath("student_zone.ItemListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("student_zone.SectionDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Section: %s", getters.GetterAny(getters.GetterKey[string]("section.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Sections"),
			Url:   lago.GetterRoutePath("student_zone.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Section Detail"),
				Url: lago.GetterRoutePath("student_zone.SectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Section"),
				Url: lago.GetterRoutePath("student_zone.SectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Section"),
				Url: lago.GetterRoutePath("student_zone.SectionDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("student_zone.ItemDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Item: %s", getters.GetterAny(getters.GetterKey[string]("item.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Items"),
			Url:   lago.GetterRoutePath("student_zone.ItemListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Item Detail"),
				Url: lago.GetterRoutePath("student_zone.ItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Item"),
				Url: lago.GetterRoutePath("student_zone.ItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Item"),
				Url: lago.GetterRoutePath("student_zone.ItemDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("student_zone.SectionFilter", &components.FormComponent[StudentZoneSection]{
		Url:    lago.GetterRoutePath("student_zone.DefaultRoute", nil),
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

	lago.RegistryPage.Register("student_zone.SectionSelectionFilter", &components.FormComponent[StudentZoneSection]{
		Url:    lago.GetterRoutePath("student_zone.SectionSelectRoute", nil),
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

	lago.RegistryPage.Register("student_zone.ItemFilter", &components.FormComponent[StudentZoneItem]{
		Url:    lago.GetterRoutePath("student_zone.ItemListRoute", nil),
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

func sectionFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "student_zone.SectionFormFieldsBody"},
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

func itemFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "student_zone.ItemFormFieldsBody"},
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
						Url:         lago.GetterRoutePath("student_zone.SectionSelectRoute", nil),
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

func registerFormPages() {
	lago.RegistryPage.Register("student_zone.SectionFormFields", sectionFormFields())
	lago.RegistryPage.Register("student_zone.ItemFormFields", itemFormFields())

	lago.RegistryPage.Register("student_zone.SectionCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneSection]{
				Url:      lago.GetterRoutePath("student_zone.SectionCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Section",
				Subtitle: "Create a new student zone section",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					sectionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Section"},
				},
			},
		},
	})

	lago.RegistryPage.Register("student_zone.SectionUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.SectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneSection]{
				Getter: getters.GetterKey[StudentZoneSection]("section"),
				Url: lago.GetterRoutePath("student_zone.SectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Section",
				Subtitle: "Update section details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					sectionFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Update Section"},
				},
			},
		},
	})

	lago.RegistryPage.Register("student_zone.ItemCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.Menu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneItem]{
				Url:      lago.GetterRoutePath("student_zone.ItemCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Item",
				Subtitle: "Create a new student zone item",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					itemFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Item"},
				},
			},
		},
	})

	lago.RegistryPage.Register("student_zone.ItemUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.ItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[StudentZoneItem]{
				Getter: getters.GetterKey[StudentZoneItem]("item"),
				Url: lago.GetterRoutePath("student_zone.ItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Item",
				Subtitle: "Update item details",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					itemFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Update Item"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	lago.RegistryPage.Register("student_zone.SectionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneSection]{
				UID:       "section-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[StudentZoneSection]]("sections"),
				CreateUrl: lago.GetterRoutePath("student_zone.SectionCreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("student_zone.SectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "student_zone.SectionFilter"},
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

	lago.RegistryPage.Register("student_zone.ItemTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.Menu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneItem]{
				UID:       "item-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[StudentZoneItem]]("items"),
				CreateUrl: lago.GetterRoutePath("student_zone.ItemCreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("student_zone.ItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "student_zone.ItemFilter"},
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

func registerDetailPages() {
	lago.RegistryPage.Register("student_zone.SectionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.SectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentZoneSection]{
				Getter: getters.GetterKey[StudentZoneSection]("section"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "student_zone.SectionDetailContent"},
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

	lago.RegistryPage.Register("student_zone.SectionDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.SectionDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this section? All items in this section will also be deleted.",
				CancelUrl: lago.GetterRoutePath("student_zone.SectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("section.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("student_zone.ItemDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.ItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[StudentZoneItem]{
				Getter: getters.GetterKey[StudentZoneItem]("item"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "student_zone.ItemDetailContent"},
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

	lago.RegistryPage.Register("student_zone.ItemDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "student_zone.ItemDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this item?",
				CancelUrl: lago.GetterRoutePath("student_zone.ItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Selection ---

func registerSelectionPages() {
	lago.RegistryPage.Register("student_zone.SectionSelectionTable", &components.Modal{
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
				FilterComponent: lago.DynamicPage{Name: "student_zone.SectionSelectionFilter"},
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
