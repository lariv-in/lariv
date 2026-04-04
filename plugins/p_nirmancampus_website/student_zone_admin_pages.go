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
		v, err := getters.Key[bool]("$in.IsLink")(ctx)
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
		Title: getters.Static("Student Zone"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Website"),
			Url:   lago.RoutePath("nirmancampus_website.AppLandingRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Sections"),
				Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("All Items"),
				Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Section: %s", getters.Any(getters.Key[string]("section.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Sections"),
			Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Section Detail"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Section"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("section.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Delete Section"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("section.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminItemDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Item: %s", getters.Any(getters.Key[string]("item.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Items"),
			Url:   lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Item Detail"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit Item"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("item.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Delete Item"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerStudentZoneAdminFilterPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionFilter", &components.FormComponent[StudentZoneSection]{
		Url:    lago.RoutePath("nirmancampus_website.StudentZoneAdminDefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.Key[string]("$get.Title"),
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
		Url:    lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.Key[string]("$get.Title"),
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
		Url:    lago.RoutePath("nirmancampus_website.StudentZoneAdminItemListRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Title",
				Name:   "Title",
				Getter: getters.Key[string]("$get.Title"),
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
						Error: getters.Key[error]("$error.Title"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Title",
								Name:     "Title",
								Required: true,
								Getter:   getters.Key[string]("$in.Title"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Order"),
						Children: []components.PageInterface{
							&components.InputNumber[int]{
								Label:  "Order",
								Name:   "Order",
								Getter: getters.Key[int]("$in.Order"),
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
				Error: getters.Key[error]("$error.Title"),
				Children: []components.PageInterface{
					&components.InputText{
						Label:    "Title",
						Name:     "Title",
						Required: true,
						Getter:   getters.Key[string]("$in.Title"),
					},
				},
			},

			&components.ContainerError{
				Error: getters.Key[error]("$error.StudentZoneSectionID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[StudentZoneSection]{
						Label:       "Section",
						Name:        "StudentZoneSectionID",
						Required:    true,
						Getter:      getters.Association[StudentZoneSection](getters.Key[uint]("$in.StudentZoneSectionID")),
						Url:         lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionSelectRoute", nil),
						Display:     getters.Key[string]("$in.Title"),
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
						Getter: getters.Key[bool]("$in.IsLink"),
						XModel: "isLink",
					},

					&components.ClientIf{
						Condition: "isLink",
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.Link"),
								Children: []components.PageInterface{
									&components.InputText{
										Label:  "Link URL",
										Name:   "Link",
										Getter: getters.Key[string]("$in.Link"),
									},
								},
							},
						},
					},

					&components.ClientIf{
						Condition: "!isLink",
						Children: []components.PageInterface{
							&components.ContainerError{
								Error: getters.Key[error]("$error.FileID"),
								Children: []components.PageInterface{
									&p_filesystem.InputVNode{
										Label: "File",
										Name:  "FileID",
										VNode: getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$in.FileID"))),
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
				Url:      lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionCreateRoute", nil),
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
				Getter: getters.Key[StudentZoneSection]("section"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
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
				Url:      lago.RoutePath("nirmancampus_website.StudentZoneAdminItemCreateRoute", nil),
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
				Getter: getters.Key[StudentZoneItem]("item"),
				Url: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
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
				UID:     "section-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[StudentZoneSection]]("sections"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionCreateRoute", nil)},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$row.ID")),
				})),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Order",
						Name:  "Order",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int]("$row.Order")))},
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
				UID:     "item-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[StudentZoneItem]]("items"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminItemFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemCreateRoute", nil)},
				},
				OnClick: getters.NavigateGetter(lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$row.ID")),
				})),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Section",
						Name:  "Section",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.StudentZoneSection.Title")},
						},
					},
					{
						Label: "Is Link",
						Name:  "IsLink",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsLink")},
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
				Getter: getters.Key[StudentZoneSection]("section"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminSectionDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Order",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int]("$in.Order")))},
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
				CancelUrl: lago.RoutePath("nirmancampus_website.StudentZoneAdminSectionDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("section.ID")),
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
				Getter: getters.Key[StudentZoneItem]("item"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "nirmancampus_website.StudentZoneAdminItemDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Section",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.StudentZoneSection.Title")},
								},
							},
							&components.ShowIf{
								Getter: getters.Any(getters.Key[bool]("$in.IsLink")),
								Children: []components.PageInterface{
									&components.LabelInline{
										Title: "Link",
										Children: []components.PageInterface{
											&components.FieldText{Getter: getters.Key[string]("$in.Link")},
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
												VNode: getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$in.FileID"))),
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
				CancelUrl: lago.RoutePath("nirmancampus_website.StudentZoneAdminItemDetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("item.ID")),
				}),
			},
		},
	})
}

// --- Selection ---

func registerStudentZoneAdminSelectionPages() {
	lago.RegistryPage.Register("nirmancampus_website.StudentZoneAdminSectionSelectionTable", &components.Modal{
		UID: "section-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[StudentZoneSection]{
				UID:   "section-selection-table",
				Title: "Select Section",
				Data:  getters.Key[components.ObjectList[StudentZoneSection]]("sections"),
				OnClick: getters.Select("StudentZoneSectionID",
					getters.Key[uint]("$row.ID"),
					getters.Key[string]("$row.Title"),
				),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "nirmancampus_website.StudentZoneAdminSectionSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Order",
						Name:  "Order",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int]("$row.Order")))},
						},
					},
				},
			},
		},
	})
}
