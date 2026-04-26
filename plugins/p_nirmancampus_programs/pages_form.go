package p_nirmancampus_programs

import (
	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/registry"
)

func universityFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "University",
		Name:     "University",
		Required: false,
		Choices:  getters.Static(UniversityChoices),
		Getter:   registry.PairFromGetter(getters.Key[string]("$in.University"), UniversityChoices),
	}
}

func programTypeFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Program type",
		Name:     "ProgramType",
		Required: false,
		Choices:  getters.Static(programTypeChoices),
		Getter:   registry.PairFromGetter(getters.Key[string]("$in.ProgramType"), programTypeChoices),
	}
}

func admissionSessionsFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Admission sessions",
		Name:     "AdmissionSessions",
		Required: false,
		Choices:  getters.Static(admissionSessionChoices),
		Getter:   registry.PairFromGetter(getters.Key[string]("$in.AdmissionSessions"), admissionSessionChoices),
	}
}

func termTypeFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Term type",
		Name:     "TermType",
		Required: false,
		Choices:  getters.Static(termTypeChoices),
		Getter:   registry.PairFromGetter(getters.Key[string]("$in.TermType"), termTypeChoices),
	}
}

func programFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "programs.ProgramFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
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
					&components.ContainerError{
						Error: getters.Key[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Code",
								Name:     "Code",
								Required: true,
								Getter:   getters.Key[string]("$in.Code"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Description"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Description",
								Name:   "Description",
								Rows:   3,
								Getter: getters.Key[string]("$in.Description"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.University"),
						Children: []components.PageInterface{
							universityFormSelect(),
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.ProgramType"),
						Children: []components.PageInterface{
							programTypeFormSelect(),
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.AdmissionSessions"),
						Children: []components.PageInterface{
							admissionSessionsFormSelect(),
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.TermType"),
						Children: []components.PageInterface{
							termTypeFormSelect(),
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.ProgramFee"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Label:    "Program fee (₹)",
								Name:     "ProgramFee",
								Required: false,
								Getter:   getters.Key[uint]("$in.ProgramFee"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.ProgramMedia"),
						Children: []components.PageInterface{
							&components.InputManyToMany[ProgramMedia]{
								Label:       "Media languages",
								Name:        "ProgramMedia",
								Getter:      getters.Key[[]ProgramMedia]("$in.ProgramMedia"),
								Url:         lago.RoutePath("programs.ProgramMediaMultiSelectRoute", nil),
								Display:     getters.Key[string]("$in.Language"),
								Placeholder: "Select instruction languages…",
								Classes:     "w-full",
							},
						},
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("programs.ProgramFormFields", programFormFields())

	lago.RegistryPage.Register("programs.ProgramCreateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name:      getters.Static("programs.ProgramCreateForm"),
				ActionURL: lago.RoutePath("programs.CreateRoute", nil),
				Children: []components.PageInterface{
					&components.FormComponent[Program]{
						Attr: getters.FormBubbling(getters.Static("programs.ProgramCreateForm")),

						Title:    "Create Program",
						Subtitle: "Create a new program",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							programFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Program"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("programs.ProgramUpdateForm"),
				ActionURL: lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[Program]{
						Getter: getters.Key[Program]("program"),
						Attr:   getters.FormBubbling(getters.Static("programs.ProgramUpdateForm")),

						Title:    "Edit Program",
						Subtitle: "Update program details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							programFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Program"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("programs.ProgramDeleteForm"),
												Url:         lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}),
												FormPostURL: lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("program.ID"))}),
												ModalUID:    "program-delete-modal",
												Classes:     "btn-error",
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
