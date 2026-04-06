package forms

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
)

func registerFieldPages() {
	lago.RegistryPage.Register("forms.FieldCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FormDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("forms.FieldCreateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[FormField]{
						Attr: getters.FormBubbling(nil),

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
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "forms.FieldEditMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				ActionURL: lago.RoutePath("forms.FieldUpdateRoute", map[string]getters.Getter[any]{
					"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
					"id":      getters.Any(getters.Key[uint]("form_field.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[FormField]{
						Getter: getters.Key[FormField]("form_field"),
						Attr:   getters.FormBubbling(nil),

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
											"form_id": getters.Any(getters.Key[uint]("form_field.FormID")),
											"id":      getters.Any(getters.Key[uint]("form_field.ID")),
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
			},
		},
	})

	lago.RegistryPage.Register("forms.FieldDeleteForm", &components.Modal{
		UID: "forms-field-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Delete field",
				Message: "Remove this field from the form?",
				Attr:    getters.FormBubbling(nil),
			},
		},
	})
}
