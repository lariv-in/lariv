package forms

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
	gomponents "maragu.dev/gomponents"
)

func init() {
	registerMenus()
	registerFormListPages()
	registerFormCRUDPages()
	registerFieldPages()
	registerSubmissionPages()
	registerPublicPage()
}

func registerMenus() {
	lago.RegistryPage.Register("forms.FormMenu", &components.SidebarMenu{
		Title: getters.Static("Forms"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All forms"),
				Url:   lago.RoutePath("forms.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Form: %s", getters.Any(getters.Key[string]("form.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all forms"),
			Url:   lago.RoutePath("forms.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Detail"),
				Url: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Edit"),
				Url: lago.RoutePath("forms.UpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Submissions"),
				Url: lago.RoutePath("forms.SubmissionsListRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldEditMenu", &components.SidebarMenu{
		Title: getters.Static("Field"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to form"),
			Url: lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
			}),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Edit field"),
				Url: lago.RoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
					"id":      getters.Any(getters.Key[uint]("form_field.ID")),
				}),
			},
		},
	})
}

func registerFormListPages() {
	lago.RegistryPage.Register("forms.FormTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Form]{
				Page:    components.Page{Key: "forms.FormTableBody"},
				UID:     "forms-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Form]]("forms"),
				Actions: []components.PageInterface{
					&components.TableButtonCreate{
						Link: lago.RoutePath("forms.CreateRoute", nil),
					},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Title")},
						},
					},
					{
						Label: "Slug",
						Name:  "Slug",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Slug")},
						},
					},
				},
			},
		},
	})
}

func fieldTableRowNavigateEdit() getters.Getter[gomponents.Node] {
	return getters.RowAttrNavigate(
		lago.RoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
			"form_id": getters.Any(getters.Key[uint]("$row.FormID")),
			"id":      getters.Any(getters.Key[uint]("$row.ID")),
		}),
	)
}

func formFieldTableColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Label",
			Name:  "Label",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.Label")},
			},
		},
		{
			Label: "Type",
			Name:  "FieldType",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.Key[string]("$row.FieldType")},
			},
		},
		{
			Label: "",
			Name:  "",
			Children: []components.PageInterface{
				&components.ContainerRow{
					Page: components.Page{Key: "forms.FieldMoveRow"},
					Children: []components.PageInterface{
						&components.ShowIf{
							Page:   components.Page{Key: "forms.FieldMoveUp"},
							Getter: getters.BoolNot(getters.Key[bool]("$isFirstRow")),
							Children: []components.PageInterface{
								&components.ButtonPost{
									Page:        components.Page{Key: "forms.FieldMoveUpCell.post"},
									Label:       "",
									Icon:        "arrow-up",
									IconClasses: "w-4 h-4",
									URL: lago.RoutePath("forms.FieldMoveUpRoute", map[string]getters.Getter[any]{
										"form_id": getters.Any(getters.Key[uint]("$row.FormID")),
										"id":      getters.Any(getters.Key[uint]("$row.ID")),
									}),
									Classes: "btn-xs btn-square btn-outline",
								},
							},
						},
						&components.ShowIf{
							Page:   components.Page{Key: "forms.FieldMoveDown"},
							Getter: getters.BoolNot(getters.Key[bool]("$isLastRow")),
							Children: []components.PageInterface{
								&components.ButtonPost{
									Page:        components.Page{Key: "forms.FieldMoveDownCell.post"},
									Label:       "",
									Icon:        "arrow-down",
									IconClasses: "w-4 h-4",
									URL: lago.RoutePath("forms.FieldMoveDownRoute", map[string]getters.Getter[any]{
										"form_id": getters.Any(getters.Key[uint]("$row.FormID")),
										"id":      getters.Any(getters.Key[uint]("$row.ID")),
									}),
									Classes: "btn-xs btn-square btn-outline",
								},
							},
						},
					},
				},
			},
		},
	}
}

func formDefinitionFields() components.PageInterface {
	return &components.ContainerColumn{
		Page: components.Page{Key: "forms.FormDefinitionFields"},
		Children: []components.PageInterface{
			&components.InputText{
				Label:    "Title",
				Name:     "Title",
				Required: true,
				Getter:   getters.Key[string]("$in.Title"),
			},
			&components.InputTextarea{
				Label:  "Description",
				Name:   "Description",
				Rows:   3,
				Getter: getters.Key[string]("$in.Description"),
			},
		},
	}
}

func registerFormCRUDPages() {
	lago.RegistryPage.Register("forms.FormCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Form]{
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("forms.CreateRoute", nil))),

				Title:    "Create form",
				Subtitle: "The public URL slug is generated from the title",
				ChildrenInput: []components.PageInterface{
					formDefinitionFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save"},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Form]{
				Getter: getters.Key[Form]("form"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Title")},
							&components.LabelInline{
								Title: "Slug",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Slug")},
								},
							},
							&components.LabelInline{
								Title: "Public URL",
								Children: []components.PageInterface{
									&components.FieldLink{
										Page:    components.Page{Key: "forms.FormDetailPublicURL"},
										Href:    lago.RoutePath("forms.PublicFormRoute", map[string]getters.Getter[any]{"slug": getters.Any(getters.Key[string]("$in.Slug"))}),
										Classes: "link link-primary link-hover break-all",
									},
								},
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.Key[string]("$in.Description"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.FieldTitle{
								Page:    components.Page{Key: "forms.FormDetailFieldsHeading"},
								Getter:  getters.Static("Fields"),
								Classes: "mt-6",
							},
							&components.DataTable[FormField]{
								Page: components.Page{Key: "forms.FormDetailFieldsTable"},
								UID:  "form-detail-fields-table",
								Data: getters.Key[components.ObjectList[FormField]](FormFieldsObjectListContextKey),
								Actions: []components.PageInterface{
									&components.TableButtonCreate{
										Link: lago.RoutePath("forms.FieldCreateRoute", map[string]getters.Getter[any]{
											"form_id": getters.Any(getters.Key[uint]("$in.ID")),
										}),
									},
								},
								RowAttr: fieldTableRowNavigateEdit(),
								Columns: formFieldTableColumns(),
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FormUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Form]{
				Getter: getters.Key[Form]("form"),
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("forms.UpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("$in.ID")),
				}))),

				Title: "Edit form",
				ChildrenInput: []components.PageInterface{
					formDefinitionFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							&components.ButtonModal{
								Label:   "Delete",
								Icon:    "trash",
								Url:     lago.RoutePath("forms.DeleteRoute", map[string]getters.Getter[any]{"form_id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Classes: "btn-outline btn-error btn-sm",
							},
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDeleteForm", &components.Modal{
		UID: "forms-form-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete form",
				Message: "This will delete the form, its field definitions, and stored submissions.",
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("forms.DeleteRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}))),
			},
		},
	})
}

func formFieldEditorBody() components.PageInterface {
	children := []components.PageInterface{
		&components.InputText{
			Label:  "Form ID",
			Name:   "FormID",
			Hidden: true,
			Getter: getters.Format("%d", getters.Any(getters.ParseUint(getters.Key[string]("$path.form_id")))),
		},
	}
	children = append(children,
		&components.InputText{
			Label:    "Label",
			Name:     "Label",
			Required: true,
			Getter:   getters.Format("%v", getters.Any(getters.Key[any]("$in.Label"))),
		},
		&components.ClientData{
			Page: components.Page{Key: "forms.FormFieldTypeOptionsScope"},
			Data: `{ fieldType: '' }`,
			Init: `(() => { const s = $el.querySelector('select[name=FieldType]'); if (!s) return; fieldType = s.value || ''; s.addEventListener('change', e => { fieldType = e.target.value; }); })()`,
			Children: []components.PageInterface{
				&components.InputSelect[string]{
					Label:    "Field type",
					Name:     "FieldType",
					Required: true,
					Choices:  getters.Static(FieldTypeRegistryPairs),
					Getter: getters.Map(getters.Key[string]("$in.FieldType"), func(_ context.Context, ft string) (registry.Pair[string, string], error) {
						for _, p := range FieldTypeRegistryPairs {
							if p.Key == ft {
								return p, nil
							}
						}
						return registry.Pair[string, string]{Key: ft, Value: ft}, nil
					}),
				},
				&components.ClientShow{
					Page:      components.Page{Key: "forms.FormFieldOptionsWhenSelect"},
					Condition: fmt.Sprintf("fieldType === '%s'", FieldTypeSelect),
					Children: []components.PageInterface{
						&components.InputStringList{
							Page:   components.Page{Key: "forms.FormFieldOptionsList"},
							Label:  "Select options",
							Name:   "Options",
							Getter: getters.JSONArray[string](getters.Key[string]("$in.Options")),
						},
					},
				},
			},
		},
		&components.InputCheckbox{
			Label:  "Required",
			Name:   "Required",
			Getter: getters.Key[bool]("$in.Required"),
		},
	)
	return &components.ContainerColumn{
		Page:     components.Page{Key: "forms.FormFieldEditorFields"},
		Children: children,
	}
}

func registerFieldPages() {
	lago.RegistryPage.Register("forms.FieldCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[FormField]{
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("forms.FieldCreateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}))),

				Title:    "Add field",
				Subtitle: "Define name, label, and type",
				ChildrenInput: []components.PageInterface{
					formFieldEditorBody(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save field"},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FieldEditMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[FormField]{
				Getter: getters.Key[FormField]("form_field"),
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
					"id":      getters.Any(getters.Key[uint]("form_field.ID")),
				}))),

				Title: "Edit field",
				ChildrenInput: []components.PageInterface{
					formFieldEditorBody(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
						Children: []components.PageInterface{
							&components.ButtonModal{
								Label: "Delete",
								Icon:  "trash",
								Url: lago.RoutePath("forms.FieldDeleteRoute", map[string]getters.Getter[any]{
									"form_id": getters.Any(getters.Key[uint]("$in.FormID")),
									"id":      getters.Any(getters.Key[uint]("$in.ID")),
								}),
								Classes: "btn-outline btn-error btn-sm",
							},
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save"},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldDeleteForm", &components.Modal{
		UID: "forms-field-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete field",
				Message: "Remove this field from the form?",
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("forms.FieldDeleteRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
					"id":      getters.Any(getters.Key[uint]("form_field.ID")),
				}))),
			},
		},
	})
}

func registerSubmissionPages() {
	lago.RegistryPage.Register("forms.SubmissionTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[FormSubmission]{
				Page:    components.Page{Key: "forms.SubmissionTableBody"},
				UID:     "form-submissions-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[FormSubmission]]("form_submissions"),
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("forms.SubmissionDetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.Any(getters.Key[uint]("$row.FormID")),
						"id":      getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Name:  "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.ID")))},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.SubmissionDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[FormSubmission]{
				Getter: getters.Key[FormSubmission]("form_submission"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.LabelInline{
								Title: "Submitted at",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Format("%v", getters.Any(getters.Key[any]("$in.CreatedAt")))},
								},
							},
							&SubmissionAnswersDisplay{
								Page: components.Page{Key: "forms.SubmissionAnswersDisplay"},
							},
						},
					},
				},
			},
		},
	})
}

func registerPublicPage() {
	lago.RegistryPage.Register("forms.PublicSubmitPage", &components.ShellBase{
		Page: components.Page{Key: "forms.PublicSubmitPage"},
		Children: []components.PageInterface{
			&components.LayoutCard{
				Page: components.Page{Key: "forms.PublicSubmitCard"},
				Children: []components.PageInterface{
					&PublicSubmitForm{
						Page:      components.Page{Key: "forms.PublicSubmitFormBody"},
						ActionURL: lago.RoutePath("forms.PublicFormRoute", map[string]getters.Getter[any]{"slug": getters.Any(getters.Key[string](ContextKeyPublicLoadedForm + ".Slug"))}),
					},
				},
			},
		},
	})
}
