package p_nirmancampus_website

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
)

func init() {
	registerImportantLinksAdminFilterPages()
	registerImportantLinksAdminFormPages()
	registerImportantLinksAdminTablePages()
	registerImportantLinksAdminDetailPages()
}

func registerImportantLinksAdminFilterPages() {
	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksFilter", &components.FormComponent[ImportantLink]{
		Attr: getters.FormBoostedGet(lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil)),

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

func importantLinksFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{Key: "nirmancampus_website.ImportantLinksFormFieldsBody"},
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

func registerImportantLinksAdminFormPages() {
	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("nirmancampus_website.ImportantLinksCreateForm"),
				ActionURL: lago.RoutePath("nirmancampus_website.ImportantLinksCreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[ImportantLink]{
						Attr: getters.FormBubbling(getters.Static("nirmancampus_website.ImportantLinksCreateForm")),

						Title:    "Create Important Link",
						Subtitle: "Create a new important link entry",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							importantLinksFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksImportForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("nirmancampus_website.ImportantLinksImportForm"),
				ActionURL: lago.RoutePath("nirmancampus_website.ImportantLinksImportRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[map[string]any]{
						Attr: getters.FormBubbling(getters.Static("nirmancampus_website.ImportantLinksImportForm")),

						Title:    "Import Important Links",
						Subtitle: "Upload a .json file containing an array of important link objects.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							&components.InputFile{
								Label:    "JSON File",
								Name:     "ImportFile",
								Required: true,
								Accept:   ".json,application/json",
							},
						},
						ChildrenAction: []components.PageInterface{
							components.ContainerRow{
								Classes: "flex gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Import"},
									&components.ButtonLink{
										Label:   "Cancel",
										Link:    lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
										Classes: "btn-outline",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.ImportantLinksDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("nirmancampus_website.ImportantLinksUpdateForm"),
				ActionURL: lago.RoutePath("nirmancampus_website.ImportantLinksUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("link.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[ImportantLink]{
						Getter: getters.Key[ImportantLink]("link"),
						Attr:   getters.FormBubbling(getters.Static("nirmancampus_website.ImportantLinksUpdateForm")),

						Title:    "Edit Important Link",
						Subtitle: "Update important link details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							importantLinksFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Update"},
											&components.ButtonModalForm{
												Label: "Delete",
												Icon:  "trash",
												Name:  getters.Static("nirmancampus_website.ImportantLinksDeleteForm"),
												Url: lago.RoutePath("nirmancampus_website.ImportantLinksDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("link.ID")),
												}),
												FormPostURL: lago.RoutePath("nirmancampus_website.ImportantLinksDeleteRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("link.ID")),
												}),
												ModalUID: "nirmancampus-important-links-delete-modal",
												Classes:  "btn-error",
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
}

func registerImportantLinksAdminTablePages() {
	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.WebsiteAdminMenu"},
		},
		Children: []components.PageInterface{
			components.ContainerRow{
				Classes: "flex justify-end mb-2",
				Children: []components.PageInterface{
					&components.ButtonLink{
						Label:   "Import JSON",
						Link:    lago.RoutePath("nirmancampus_website.ImportantLinksImportRoute", nil),
						Icon:    "arrow-up-tray",
						Classes: "btn-outline btn-sm",
					},
				},
			},
			&components.DataTable[ImportantLink]{
				UID:     "important-links-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[ImportantLink]]("links"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "nirmancampus_website.ImportantLinksFilter"}},
					&components.TableButtonCreate{Link: lago.RoutePath("nirmancampus_website.ImportantLinksCreateRoute", nil)},
				},
				RowAttr: getters.RowAttrNavigate(lago.RoutePath("nirmancampus_website.ImportantLinksDetailRoute", map[string]getters.Getter[any]{
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
					{
						Label: "Mode",
						Name:  "IsLink",
						Children: []components.PageInterface{
							&components.FieldCheckbox{Getter: getters.Key[bool]("$row.IsLink")},
						},
					},
					{
						Label: "Value",
						Name:  "Value",
						Children: []components.PageInterface{
							&components.ShowIf{
								Getter: getters.Any(getters.Key[bool]("$row.IsLink")),
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$row.Link")},
								},
							},
							&components.ShowIf{
								Getter: getters.BoolNot(getters.Key[bool]("$row.IsLink")),
								Children: []components.PageInterface{
									&p_filesystem.FieldFile{
										VNode: getters.Association[p_filesystem.VNode](getters.Deref(getters.Key[*uint]("$row.FileID"))),
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

func registerImportantLinksAdminDetailPages() {
	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksDetailMenu", &components.SidebarMenu{
		Title: getters.Static("Important Links"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to Important Links"),
			Url:   lago.RoutePath("nirmancampus_website.ImportantLinksDefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("nirmancampus_website.ImportantLinksUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("link.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "nirmancampus_website.ImportantLinksDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[ImportantLink]{
				Getter: getters.Key[ImportantLink]("link"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "nirmancampus_website.ImportantLinksDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Order",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[int]("$in.Order")))},
								},
							},
							&components.LabelInline{
								Title: "Is Link",
								Children: []components.PageInterface{
									&components.FieldCheckbox{Getter: getters.Key[bool]("$in.IsLink")},
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
								Getter: getters.BoolNot(getters.Key[bool]("$in.IsLink")),
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

	lago.RegistryPage.Register("nirmancampus_website.ImportantLinksDeleteForm", &components.Modal{
		UID: "nirmancampus-important-links-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this important link?",
				Attr:    getters.FormBubbling(getters.Key[string]("$get.name")),
			},
		},
	})
}
