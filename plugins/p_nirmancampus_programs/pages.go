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

func universityChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "IGNOU", Value: "IGNOU"},
		{Key: "MRSPTU", Value: "MRSPTU"},
	}
}

func programTypeChoices() []registry.Pair[string, string] {
	return []registry.Pair[string, string]{
		{Key: "certificate", Value: "Certificate"},
		{Key: "diploma", Value: "Diploma"},
		{Key: "bachelor", Value: "Bachelor"},
		{Key: "masters", Value: "Masters"},
	}
}

func universityFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$get.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programUniversityPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$in.University")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func universityFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "University",
		Name:    "University",
		Choices: getters.GetterStatic(universityChoices()),
		Getter:  universityFilterPairGetter(),
	}
}

func universityFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "University",
		Name:     "University",
		Required: false,
		Choices:  getters.GetterStatic(universityChoices()),
		Getter:   programUniversityPairGetter(),
	}
}

func programTypeFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$get.ProgramType")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		for _, p := range programTypeChoices() {
			if p.Key == s {
				return p, nil
			}
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programProgramTypePairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$in.ProgramType")(ctx)
		if err != nil || s == "" {
			return registry.Pair[string, string]{}, nil
		}
		for _, p := range programTypeChoices() {
			if p.Key == s {
				return p, nil
			}
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func programTypeFilterSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:   "Program type",
		Name:    "ProgramType",
		Choices: getters.GetterStatic(programTypeChoices()),
		Getter:  programTypeFilterPairGetter(),
	}
}

func programTypeFormSelect() *components.InputSelect[string] {
	return &components.InputSelect[string]{
		Label:    "Program type",
		Name:     "ProgramType",
		Required: false,
		Choices:  getters.GetterStatic(programTypeChoices()),
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
		{Key: TermTypeSemester, Value: "Semester"},
	}
}

func programAdmissionSessionsDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		s, err := getters.GetterKey[string]("$in.AdmissionSessions")(ctx)
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
		s, err := getters.GetterKey[string]("$in.TermType")(ctx)
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

func stringSliceJoinOrDash(g getters.Getter[[]string]) getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		sl, err := g(ctx)
		if err != nil {
			return "", err
		}
		if len(sl) == 0 {
			return "—", nil
		}
		return strings.Join(sl, ", "), nil
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
		units, err := getters.GetterKey[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
		if err != nil {
			return nil, err
		}
		if units == nil {
			return []ProgramStructureUnit{}, nil
		}
		return units, nil
	}
}

func programStructureNonEmptyGetter() getters.Getter[any] {
	return func(ctx context.Context) (any, error) {
		units, err := getters.GetterKey[[]ProgramStructureUnit]("$in.ProgramStructureUnits")(ctx)
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
		Title: getters.GetterStatic("Programs"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Programs"),
				Url:   lago.GetterRoutePath("programs.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Program: %s", getters.GetterAny(getters.GetterKey[string]("program.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Programs"),
			Url:   lago.GetterRoutePath("programs.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Program Detail"),
				Url: lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit Program"),
				Url: lago.GetterRoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit program structure"),
				Url: lago.GetterRoutePath("programs.StructureEditRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Delete Program"),
				Url: lago.GetterRoutePath("programs.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
		},
	})
}

func registerFilterPages() {
	lago.RegistryPage.Register("programs.ProgramFilter", &components.FormComponent[Program]{
		Url:    lago.GetterRoutePath("programs.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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
		Url:    lago.GetterRoutePath("programs.SelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputText{
				Label:  "Name",
				Name:   "Name",
				Getter: getters.GetterKey[string]("$get.Name"),
			},
			&components.InputText{
				Label:  "Code",
				Name:   "Code",
				Getter: getters.GetterKey[string]("$get.Code"),
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
						Error: getters.GetterKey[error]("$error.Name"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Name",
								Name:     "Name",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Name"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Code"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Code",
								Name:     "Code",
								Required: true,
								Getter:   getters.GetterKey[string]("$in.Code"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Description"),
						Children: []components.PageInterface{
							&components.InputTextarea{
								Label:  "Description",
								Name:   "Description",
								Rows:   3,
								Getter: getters.GetterKey[string]("$in.Description"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.University"),
						Children: []components.PageInterface{
							universityFormSelect(),
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.ProgramType"),
						Children: []components.PageInterface{
							programTypeFormSelect(),
						},
					},
				},
			},
		},
	}
}

func programCreateUrlGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		role, err := getters.GetterKey[string]("$role")(ctx)
		if err != nil {
			return "", err
		}
		if role == "superuser" || role == "admin" {
			return lago.GetterRoutePath("programs.CreateRoute", nil)(ctx)
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
				Url:      lago.GetterRoutePath("programs.CreateRoute", nil),
				Method:   http.MethodPost,
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
				Getter: getters.GetterKey[Program]("program"),
				Url: lago.GetterRoutePath("programs.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Program",
				Subtitle: "Update program details",
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
				Data:    getters.GetterKey[components.ObjectList[Program]]("programs"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramFilter"}},
					&components.TableButtonCreate{Link: programCreateUrlGetter()},
				},
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.ProgramType")},
						},
					},
					{
						Label: "Description",
						Name:  "Description",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Description")},
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
				Getter: getters.GetterKey[Program]("program"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "programs.ProgramDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Code")},
							&components.LabelInline{
								Title: "University",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.University")},
								},
							},
							&components.LabelInline{
								Title: "Program type",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.ProgramType")},
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
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Description")},
								},
							},
							&components.LabelNewline{
								Title: "Program structure",
								Children: []components.PageInterface{
									&components.ShowIf{
										Getter: programStructureNonEmptyGetter(),
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
																		Getter: getters.GetterFormat(
																			"%d",
																			getters.GetterAny(getters.GetterKey[int]("$row.TermNumber")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Compulsory",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListJoinOrDash(
																			getters.GetterKey[[]courses.Course]("$row.CompulsoryCourses"),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional count",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: getters.GetterFormat(
																			"%d",
																			getters.GetterAny(getters.GetterKey[int]("$row.OptionalCourseCount")),
																		),
																	},
																},
															},
															&components.LabelInline{
																Title: "Optional course pool",
																Children: []components.PageInterface{
																	&components.FieldText{
																		Getter: courseListJoinOrDash(
																			getters.GetterKey[[]courses.Course]("$row.OptionalCourseSelectionPool"),
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
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("programs.ProgramDeleteForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "programs.ProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this program?",
				CancelUrl: lago.GetterRoutePath("programs.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				}),
			},
		},
	})
}

func registerSelectionPages() {
	lago.RegistryPage.Register("programs.ProgramSelectionTable", &components.Modal{
		UID:   "program-selection-modal",
		Title: "Select Program",
		Children: []components.PageInterface{
			&components.DataTable[Program]{
				Page:    components.Page{Key: "programs.ProgramSelectionTableBody"},
				UID:     "program-selection-table",
				Data:    getters.GetterKey[components.ObjectList[Program]]("programs"),
				OnClick: getters.GetterSelect("ProgramID", getters.GetterKey[uint]("$row.ID"), getters.GetterKey[string]("$row.Name")),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{Child: lago.DynamicPage{Name: "programs.ProgramSelectionFilter"}},
				},
				Columns: []components.TableColumn{
					{
						Label: "Name",
						Name:  "Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Name")},
						},
					},
					{
						Label: "Code",
						Name:  "Code",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Code")},
						},
					},
					{
						Label: "University",
						Name:  "University",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.University")},
						},
					},
					{
						Label: "Program type",
						Name:  "ProgramType",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.ProgramType")},
						},
					},
				},
			},
		},
	})
}
