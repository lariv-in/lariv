package forms

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
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
		Title: getters.GetterStatic("Forms"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All forms"),
				Url:   lago.GetterRoutePath("forms.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Form: %s", getters.GetterAny(getters.GetterKey[string]("form.Title"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all forms"),
			Url:   lago.GetterRoutePath("forms.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Detail"),
				Url: lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit"),
				Url: lago.GetterRoutePath("forms.UpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Submissions"),
				Url: lago.GetterRoutePath("forms.SubmissionsListRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete"),
				Url: lago.GetterRoutePath("forms.DeleteRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldEditMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Field"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to form"),
			Url: lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
				"form_id": getters.GetterAny(getters.GetterKey[uint]("form_field.FormID")),
			}),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit field"),
				Url: lago.GetterRoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form_field.FormID")),
					"id":      getters.GetterAny(getters.GetterKey[uint]("form_field.ID")),
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
				Page:      components.Page{Key: "forms.FormTableBody"},
				UID:       "forms-table",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[Form]]("forms"),
				CreateUrl: lago.GetterRoutePath("forms.CreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Title",
						Name:  "Title",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Title")},
						},
					},
					{
						Label: "Slug",
						Name:  "Slug",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Slug")},
						},
					},
				},
			},
		},
	})
}

func fieldTableRowNavigateEdit() getters.Getter[string] {
	return getters.GetterNavigateGetter(
		lago.GetterRoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
			"form_id": getters.GetterAny(getters.GetterKey[uint]("$row.FormID")),
			"id":      getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
		}),
	)
}

func formFieldTableColumns() []components.TableColumn {
	return []components.TableColumn{
		{
			Label: "Label",
			Name:  "Label",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.GetterKey[string]("$row.Label")},
			},
		},
		{
			Label: "Type",
			Name:  "FieldType",
			Children: []components.PageInterface{
				&components.FieldText{Getter: getters.GetterKey[string]("$row.FieldType")},
			},
		},
		{
			Label: "",
			Name:  "",
			Children: []components.PageInterface{
				&components.ContainerRow{
					Page:    components.Page{Key: "forms.FieldMoveRow"},
					Classes: "items-center justify-center",
					Children: []components.PageInterface{
						&components.ShowIf{
							Page:   components.Page{Key: "forms.FieldMoveUp"},
							Getter: getters.GetterBoolNot(getters.GetterKey[bool]("$isFirstRow")),
							Children: []components.PageInterface{
								&components.ButtonPost{
									Page:        components.Page{Key: "forms.FieldMoveUpCell.post"},
									Label:       "",
									Icon:        "arrow-up",
									IconClasses: "w-4 h-4",
									URL: lago.GetterRoutePath("forms.FieldMoveUpRoute", map[string]getters.Getter[any]{
										"form_id": getters.GetterAny(getters.GetterKey[uint]("$row.FormID")),
										"id":      getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
									}),
									Classes: "btn-xs btn-square btn-outline",
								},
							},
						},
						&components.ShowIf{
							Page:   components.Page{Key: "forms.FieldMoveDown"},
							Getter: getters.GetterBoolNot(getters.GetterKey[bool]("$isLastRow")),
							Children: []components.PageInterface{
								&components.ButtonPost{
									Page:        components.Page{Key: "forms.FieldMoveDownCell.post"},
									Label:       "",
									Icon:        "arrow-down",
									IconClasses: "w-4 h-4",
									URL: lago.GetterRoutePath("forms.FieldMoveDownRoute", map[string]getters.Getter[any]{
										"form_id": getters.GetterAny(getters.GetterKey[uint]("$row.FormID")),
										"id":      getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
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
				Getter:   getters.GetterKey[string]("$in.Title"),
			},
			&components.InputTextarea{
				Label:  "Description",
				Name:   "Description",
				Rows:   3,
				Getter: getters.GetterKey[string]("$in.Description"),
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
				Url:      lago.GetterRoutePath("forms.CreateRoute", nil),
				Method:   http.MethodPost,
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
				Getter: getters.GetterKey[Form]("form"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Title")},
							&components.LabelInline{
								Title: "Slug",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Slug")},
								},
							},
							&components.LabelInline{
								Title: "Public URL",
								Children: []components.PageInterface{
									&components.FieldLink{
										Page:    components.Page{Key: "forms.FormDetailPublicURL"},
										Href:    lago.GetterRoutePath("forms.PublicFormRoute", map[string]getters.Getter[any]{"slug": getters.GetterAny(getters.GetterKey[string]("$in.Slug"))}),
										Classes: "link link-primary link-hover break-all",
									},
								},
							},
							&components.LabelInline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter:  getters.GetterKey[string]("$in.Description"),
										Classes: "whitespace-pre-wrap",
									},
								},
							},
							&components.FieldTitle{
								Page:    components.Page{Key: "forms.FormDetailFieldsHeading"},
								Getter:  getters.GetterStatic("Fields"),
								Classes: "mt-6",
							},
							&components.DataTable[FormField]{
								Page:    components.Page{Key: "forms.FormDetailFieldsTable"},
								UID:     "form-detail-fields-table",
								Classes: "w-full [&_th]:text-center [&_th]:align-middle [&_td]:text-center [&_td]:align-middle",
								Data:    getters.GetterKey[components.ObjectList[FormField]](FormFieldsObjectListContextKey),
								CreateUrl: lago.GetterRoutePath("forms.FieldCreateRoute", map[string]getters.Getter[any]{
									"form_id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
								}),
								OnClick: fieldTableRowNavigateEdit(),
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
				Getter: getters.GetterKey[Form]("form"),
				Url: lago.GetterRoutePath("forms.UpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method: http.MethodPost,
				Title:  "Edit form",
				ChildrenInput: []components.PageInterface{
					formDefinitionFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save"},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FormDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete form",
				Message: "This will delete the form, its field definitions, and stored submissions.",
				CancelUrl: lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
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
			Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterParseUint(getters.GetterKey[string]("$path.form_id")))),
		},
	}
	children = append(children,
		&components.InputText{
			Label:    "Label",
			Name:     "Label",
			Required: true,
			Getter:   getters.GetterFormat("%v", getters.GetterAny(getters.GetterKey[any]("$in.Label"))),
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
					Choices:  getters.GetterStatic(FieldTypeRegistryPairs),
					Getter: getters.GetterMap(getters.GetterKey[string]("$in.FieldType"), func(_ context.Context, ft string) (registry.Pair[string, string], error) {
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
							Getter: getters.GetterJsonArray[string](getters.GetterKey[string]("$in.Options")),
						},
					},
				},
			},
		},
		&components.InputCheckbox{
			Label:  "Required",
			Name:   "Required",
			Getter: getters.GetterKey[bool]("$in.Required"),
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
				Url: lago.GetterRoutePath("forms.FieldCreateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form.ID")),
				}),
				Method:   http.MethodPost,
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
				Getter: getters.GetterKey[FormField]("form_field"),
				Url: lago.GetterRoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form_field.FormID")),
					"id":      getters.GetterAny(getters.GetterKey[uint]("form_field.ID")),
				}),
				Method: http.MethodPost,
				Title:  "Edit field",
				ChildrenInput: []components.PageInterface{
					formFieldEditorBody(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save"},
				},
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FieldEditMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete field",
				Message: "Remove this field from the form?",
				CancelUrl: lago.GetterRoutePath("forms.DetailRoute", map[string]getters.Getter[any]{
					"form_id": getters.GetterAny(getters.GetterKey[uint]("form_field.FormID")),
				}),
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
				Data:    getters.GetterKey[components.ObjectList[FormSubmission]]("form_submissions"),
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("forms.SubmissionDetailRoute", map[string]getters.Getter[any]{
						"form_id": getters.GetterAny(getters.GetterKey[uint]("$row.FormID")),
						"id":      getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "ID",
						Name:  "ID",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$row.ID")))},
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
				Getter: getters.GetterKey[FormSubmission]("form_submission"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Children: []components.PageInterface{
							&components.LabelInline{
								Title: "Submitted at",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterFormat("%v", getters.GetterAny(getters.GetterKey[any]("$in.CreatedAt")))},
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
						ActionURL: lago.GetterRoutePath("forms.PublicFormRoute", map[string]getters.Getter[any]{"slug": getters.GetterAny(getters.GetterKey[string](ContextKeyPublicLoadedForm + ".Slug"))}),
					},
				},
			},
		},
	})
}
