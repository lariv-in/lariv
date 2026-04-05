package p_nirmancampus_programs

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	courses "github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
)

func universityFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$get.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromMap(s, universityChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programUniversityPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromMap(s, universityChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func universityFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "University",
		Name:    "University",
		Choices: getters.Static(registry.PairsFromMap(universityChoices)),
		Getter:  universityFilterPairGetter(),
	}
}

func universityFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "University",
		Name:     "University",
		Required: false,
		Choices:  getters.Static(registry.PairsFromMap(universityChoices)),
		Getter:   programUniversityPairGetter(),
	}
}

func programTypeFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$get.ProgramType")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromMap(s, programTypeChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programProgramTypePairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.ProgramType")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		if p, ok := registry.PairFromMap(s, programTypeChoices); ok {
			return p, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programTypeFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "Program type",
		Name:    "ProgramType",
		Choices: getters.Static(registry.PairsFromMap(programTypeChoices)),
		Getter:  programTypeFilterPairGetter(),
	}
}

func programTypeFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Program type",
		Name:     "ProgramType",
		Required: false,
		Choices:  getters.Static(registry.PairsFromMap(programTypeChoices)),
		Getter:   programProgramTypePairGetter(),
	}
}

func admissionSessionChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: AdmissionSessionJan, Value: "January"},
		{Key: AdmissionSessionJuly, Value: "July"},
		{Key: AdmissionSessionBoth, Value: "January and July"},
	}
}

func termTypeChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: TermTypeYear, Value: "Year"},
		{Key: TermTypeSession, Value: "Session"},
	}
}

func programAdmissionSessionsPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.AdmissionSessions")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		for _, p := range admissionSessionChoices() {
			if p.Key == s {
				return p, nil
			}
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programTermTypePairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string]("$in.TermType")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		for _, p := range termTypeChoices() {
			if p.Key == s {
				return p, nil
			}
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func admissionSessionsFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Admission sessions",
		Name:     "AdmissionSessions",
		Required: false,
		Choices:  getters.Static(admissionSessionChoices()),
		Getter:   programAdmissionSessionsPairGetter(),
	}
}

func termTypeFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Term type",
		Name:     "TermType",
		Required: false,
		Choices:  getters.Static(termTypeChoices()),
		Getter:   programTermTypePairGetter(),
	}
}

func programAdmissionSessionsDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.AdmissionSessions")(ctx)
		if err != nil || s == "" {
			return "—", nil
		}
		for _, p := range admissionSessionChoices() {
			if p.Key == s {
				return p.Value, nil
			}
		}
		return s, nil
	}
}

func programTermTypeDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.Key[string]("$in.TermType")(ctx)
		if err != nil || s == "" {
			return "—", nil
		}
		for _, p := range termTypeChoices() {
			if p.Key == s {
				return p.Value, nil
			}
		}
		return s, nil
	}
}

func courseListJoinOrDash(g getters.Getter[[]courses.Course]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		list, err := g(ctx)
		if err != nil {
			return "", err
		}
		if len(list) == 0 {
			return "—", nil
		}
		codes := make([]string, 0, len(list))
		for _, c := range list {
			codes = append(codes, c.Code)
		}
		return strings.Join(codes, ", "), nil
	}
}

func programStructureRowsGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		units, err := getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
		if err != nil {
			return nil, err
		}
		if units == nil {
			return []ProgramStructureUnit{}, nil
		}
		return units, nil
	}
}

func programStructureNonEmptyGetter() getters.Getter[bool] {
	return func(ctx context.Context) (bool, error) {
		units, err := getters.Key[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
		if err != nil {
			return false, err
		}
		return len(units) > 0, nil
	}
}

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
	registerStructurePages()
}

func registerMenuPages() {
	lago.RegistryPage.Register("programs.ProgramMenu", &components.SidebarMenu{
		Title: getters.Static("Programs"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Programs"),
				Url:   lago.RoutePath("programs.DefaultRoute", nil),
			},
			&components.SidebarMenuItem{
				Title: getters.Static("Courses"),
				Url:   lago.RoutePath("courses.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Program: %s", getters.Any(getters.Key[string]("program.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Programs"),
			Url:   lago.RoutePath("programs.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Program Detail"),
				Url: lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Program"),
				Url: lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit program structure"),
				Url: lago.RoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("programs.ProgramFilter", &components.FormComponent[Program]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(lago.RoutePath("programs.DefaultRoute", nil))),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			universityFilterSelect(),
			programTypeFilterSelect(),
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

	lago.RegistryPage.Register("programs.ProgramSelectionFilter", &components.FormComponent[Program]{
		Attr: getters.FormAttr(http.MethodGet, getters.FormSubmitGet(lago.RoutePath("programs.SelectRoute", nil))),

		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.Key[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.Key[string]("$get.Code"),
			},
			universityFilterSelect(),
			programTypeFilterSelect(),
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
		},
	}
}

func programCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.Key[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.RoutePath("programs.CreateRoute", nil)(ctx)
		}
		return "", fmt.Errorf("you do not have permission to do this action")
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
			&components.FormComponent[Program]{
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("programs.CreateRoute", nil))),

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
	})

	lago.RegistryPage.Register("programs.ProgramUpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[Program]{
				Getter: getters.Key[Program]("program"),
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmit(lago.RoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				}))),

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
							&components.ButtonModal{
								Page:    components.Page{Roles: []string{"admin", "superuser"}},
								Label:   "Delete",
								Icon:    "trash",
								Url:     lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("$in.ID"))}),
								Classes: "btn-outline btn-error btn-sm",
							},
							&components.ContainerRow{
								Classes: "flex justify-end gap-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Save Program"},
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
	lago.RegistryPage.Register("programs.ProgramTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:    components.Page{Key: "programs.ProgramTableBody"},
				UID:     "program-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramFilter"}},
					&components.TableButtonCreate{Link: programCreateUrlGetter()},
				},
				RowAttr: getters.RowAttrNavigate(
					lago.RoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
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
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ProgramType")},
						},
					},
					{
						Label: "Description",
						Name:  "Description",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Description")},
						},
					},
				},
			},
		},
	})
}

func registerDetailPages() {
	lago.RegistryPage.Register("programs.ProgramDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[Program]{
				Getter: getters.Key[Program]("program"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "programs.ProgramDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Code")},
							&components.LabelInline{
								Title: "University",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.University")},
								},
							},
							&components.LabelInline{
								Title: "Program type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.ProgramType")},
								},
							},
							&components.LabelInline{
								Title: "Admission sessions",
								Children: []components.PageInterface{
									&components.FieldText{Getter: programAdmissionSessionsDisplayGetter()},
								},
							},
							&components.LabelInline{
								Title: "Term type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: programTermTypeDisplayGetter()},
								},
							},
							&components.LabelNewline{
								Title: "Description",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Description")},
								},
							},
							&components.LabelNewline{
								Title: "Program structure",
								Children: []components.PageInterface{
									&components.ShowIf{
										Getter: getters.Any(programStructureNonEmptyGetter()),
										Children: []components.PageInterface{
											&components.FieldList{
												Getter:  programStructureRowsGetter(),
												Classes: "flex flex-col gap-2",
												Children: []components.PageInterface{
													&components.ContainerColumn{
														Classes: "rounded-box border border-base-300 p-2 card card-body",
														Children: []components.PageInterface{
															&components.LabelInline{
																Title: "Term",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: getters.Format(
																			"%d",
																			getters.Any(getters.Key[uint]("$row.TermNumber")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Compulsory",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListJoinOrDash(
																			getters.Key[[]courses.Course]("$row.CompulsoryCourses"),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional count",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: getters.Format(
																			"%d",
																			getters.Any(getters.Key[uint]("$row.OptionalCourseCount")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional course pool",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListJoinOrDash(
																			getters.Key[[]courses.Course]("$row.OptionalCourseSelectionPool"),
																		),
																	},
																},
															},
														},
													},
												},
											},
										},
									},
									&components.ShowIf{
										Getter: getters.BoolNot(programStructureNonEmptyGetter()),
										Children: []components.PageInterface{
											&components.ButtonLink{
												Page:  components.Page{Roles: []string{"admin", "superuser"}},
												Label: "Add Program Structure",
												Link: lago.RoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
													"id": getters.Any(getters.Key[uint]("$in.ID")),
												}),
												Classes: "btn-primary btn-sm w-fit",
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

	lago.RegistryPage.Register("programs.ProgramDeleteForm", &components.Modal{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		UID:  "program-delete-modal",
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this program?",
				Attr: getters.FormAttr(http.MethodPost, getters.FormSubmitCloseModal(lago.RoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("program.ID")),
				}))),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("programs.ProgramSelectionTable", &components.Modal{
		UID: "program-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:    components.Page{Key: "programs.ProgramSelectionTableBody"},
				UID:     "program-selection-table",
				Title:   "Select Program",
				Data:    getters.Key[components.ObjectList[Program]]("programs"),
				RowAttr: getters.RowAttrSelect("ProgramID", getters.Key[uint]("$row.ID"), getters.Key[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramSelectionFilter"}},
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
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$row.ProgramType")},
						},
					},
				},
			},
		},
	})
}
