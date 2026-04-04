package p_nirmancampus_academicrecords

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
)

func init() {
	registerMenuPages()
	registerStudentsMenuAcademicRecordsEntry()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

var courseDetailLink = lago.RoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
	"id": getters.Any(getters.Key[uint]("$in.ID")),
})

func tableColumns() []components.TableColumn {
	return []components.TableColumn{
		{Label: "Student", Name: "Student.User.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Student.User.Name")},
		}},
		{Label: "Program", Name: "Program.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Program.Name")},
		}},
		{Label: "Session", Name: "Session.Name", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Session.Name")},
		}},
		{Label: "Status", Name: "Status", Children: []components.PageInterface{
			&components.FieldText{Getter: getters.Key[string]("$row.Status")},
		}},
		{Label: "Term", Name: "Term", Children: []components.PageInterface{
			&components.FieldText{
				Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$row.Term"))),
			},
		}},
	}
}

// --- Form Field Getters ---

// programStructureUnitForIn loads the ProgramStructureUnit for $in.ProgramID
// and $in.Term. When preloadOptionalPool is true, OptionalCourseSelectionPool
// is preloaded (for multi-select URLs).
func optionalCourseCountDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		psu, err := getters.Key[p_nirmancampus_programs.ProgramStructureUnit](academicRecordProgramStructureUnitContextKey)(ctx)
		if err != nil || psu.ID == 0 {
			return "—", nil
		}
		return fmt.Sprintf("%d", psu.OptionalCourseCount), nil
	}
}

func optionalCoursesMultiSelectURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		base, err := lago.RoutePath("courses.MultiSelectRoute", nil)(ctx)
		if err != nil {
			return "", err
		}
		u, errParse := url.Parse(base)
		if errParse != nil {
			return base, nil
		}
		psu, err := getters.Key[p_nirmancampus_programs.ProgramStructureUnit](academicRecordProgramStructureUnitContextKey)(ctx)
		q := u.Query()
		if err != nil || psu.ID == 0 || len(psu.OptionalCourseSelectionPool) == 0 {
			q.Set("pool_course_ids", "")
		} else {
			parts := make([]string, 0, len(psu.OptionalCourseSelectionPool))
			for _, c := range psu.OptionalCourseSelectionPool {
				parts = append(parts, strconv.FormatUint(uint64(c.ID), 10))
			}
			q.Set("pool_course_ids", strings.Join(parts, ","))
		}
		u.RawQuery = q.Encode()
		return u.String(), nil
	}
}

// --- Menus ---

func registerStudentsMenuAcademicRecordsEntry() {
	lago.RegistryPage.Patch("students.StudentMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			return page
		}
		menu.Children = append(menu.Children, &components.SidebarMenuItem{
			Title: getters.Static("Academic Records"),
			Url:   lago.RoutePath("academicrecords.DefaultRoute", nil),
		})
		return menu
	})
}

func registerMenuPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordDetailMenu", &components.SidebarMenu{
		Title: getters.Format("Record: %s", getters.Any(getters.Key[string]("academicrecord.Student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to all Academic Records"),
			Url:   lago.RoutePath("academicrecords.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("Academic Record Detail"),
				Url: lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Edit Academic Record"),
				Url: lago.RoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.Static("Delete Academic Record"),
				Url: lago.RoutePath("academicrecords.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFilter", &components.FormComponent[AcademicRecord]{
		OnSubmit: getters.FormSubmitGet(lago.RoutePath("academicrecords.DefaultRoute", nil)),
		Method:   http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputSelect[string]{
				Label:   "Status",
				Name:    "Status",
				Choices: getters.Static(registry.PairsFromMap(AcademicRecordStatusChoices)),
				Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
					s, err := getters.Key[string]("$get.Status")(ctx)
					if err != nil || s == "" {
						return registry.Pair[string, string]{}, nil
					}
					if p, ok := registry.PairFromMap(s, AcademicRecordStatusChoices); ok {
						return p, nil
					}
					return registry.Pair[string, string]{Key: s, Value: s}, nil
				},
			},
			&components.InputText{
				Label:  "Term",
				Name:   "Term",
				Getter: getters.Key[string]("$get.Term"),
			},
			&components.InputForeignKey[p_nirmancampus_programs.Program]{
				Label:       "Program",
				Name:        "ProgramID",
				Url:         lago.RoutePath("programs.SelectRoute", nil),
				Placeholder: "Filter by program...",
				Display:     getters.Key[string]("$in.Name"),
				Getter: getters.Association[p_nirmancampus_programs.Program](
					getters.Key[uint]("$get.ProgramID"),
				),
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

func createFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.SessionID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[sessions.Session]{
								Label:       "Session",
								Name:        "SessionID",
								Required:    true,
								Url:         lago.RoutePath("sessions.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select a session…",
								Getter: getters.Association[sessions.Session](
									getters.Key[uint]("$in.SessionID"),
								),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.StudentID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_students.Student]{
								Label:       "Student",
								Name:        "StudentID",
								Required:    true,
								Url:         lago.RoutePath("students.SelectRoute", nil),
								Display:     getters.Key[string]("$in.StudentNo"),
								Placeholder: "Select a student...",
								Getter: getters.Association[p_nirmancampus_students.Student](
									getters.Key[uint]("$in.StudentID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.ProgramID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_programs.Program]{
								Label:       "Program",
								Name:        "ProgramID",
								Required:    true,
								Url:         lago.RoutePath("programs.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select a program...",
								Getter: getters.Association[p_nirmancampus_programs.Program](
									getters.Key[uint]("$in.ProgramID"),
								),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Term"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Label:    "Term",
								Name:     "Term",
								Required: true,
								Getter:   getters.Key[uint]("$in.Term"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Status"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Status",
								Name:     "Status",
								Required: true,
								Choices:  getters.Static(registry.PairsFromMap(AcademicRecordStatusChoices)),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.Status")(ctx)
									if err != nil || s == "" {
										if p, ok := registry.PairFromMap(AcademicRecordStatusEnrolled, AcademicRecordStatusChoices); ok {
											return p, nil
										}
										return registry.Pair[string, string]{Key: AcademicRecordStatusEnrolled, Value: AcademicRecordStatusEnrolled}, nil
									}
									if p, ok := registry.PairFromMap(s, AcademicRecordStatusChoices); ok {
										return p, nil
									}
									return registry.Pair[string, string]{Key: s, Value: s}, nil
								},
							},
						},
					},
				},
			},
		},
	}
}

func editFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 mt-4",
				Children: []components.PageInterface{
					&components.LabelInline{
						Title: "Student",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Format(
								"%s (%s)",
								getters.Any(getters.Key[string]("$in.Student.StudentNo")),
								getters.Any(getters.Key[string]("$in.Student.User.Name")),
							)},
						},
					},
					&components.LabelInline{
						Title: "Program",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.ForeignKey[p_nirmancampus_programs.Program, uint, string](
									getters.Key[uint]("$in.ProgramID"),
									"Name",
								),
							},
						},
					},
					&components.LabelInline{
						Title: "Session",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$in.Session.Name")},
						},
					},
					&components.LabelInline{
						Title: "Term",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.Term"))),
							},
						},
					},
				},
			},
			&components.InputForeignKey[p_nirmancampus_students.Student]{
				Hidden: true,
				Name:   "StudentID",
				Getter: getters.Association[p_nirmancampus_students.Student](
					getters.Key[uint]("$in.StudentID"),
				),
			},
			&components.InputForeignKey[p_nirmancampus_programs.Program]{
				Hidden: true,
				Name:   "ProgramID",
				Getter: getters.Association[p_nirmancampus_programs.Program](
					getters.Key[uint]("$in.ProgramID"),
				),
			},
			&components.InputNumber[uint]{
				Hidden:   true,
				Name:     "Term",
				Getter:   getters.Key[uint]("$in.Term"),
				Required: true,
			},
			&components.LabelNewline{
				Title: "Compulsory courses",
				Children: []components.PageInterface{
					&components.FieldManyToMany[p_nirmancampus_courses.Course]{
						Getter:  getters.Key[[]p_nirmancampus_courses.Course]("$in.CompulsoryCourses"),
						Display: getters.Key[string]("$in.Name"),
						Link:    courseDetailLink,
						Classes: "w-full",
					},
				},
			},
			&components.LabelInline{
				Title: "Optional course count",
				Children: []components.PageInterface{
					&components.FieldText{Getter: optionalCourseCountDisplayGetter()},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.OptionalCourses"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_nirmancampus_courses.Course]{
						Label:       "Optional courses",
						Name:        "OptionalCourses",
						Getter:      getters.Key[[]p_nirmancampus_courses.Course]("$in.OptionalCourses"),
						Url:         optionalCoursesMultiSelectURLGetter(),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select optional courses from the program pool…",
						Classes:     "w-full",
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.Status"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Status",
								Name:     "Status",
								Required: true,
								Choices:  getters.Static(registry.PairsFromMap(AcademicRecordStatusChoices)),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.Status")(ctx)
									if err != nil || s == "" {
										return registry.Pair[string, string]{}, nil
									}
									if p, ok := registry.PairFromMap(s, AcademicRecordStatusChoices); ok {
										return p, nil
									}
									return registry.Pair[string, string]{Key: s, Value: s}, nil
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
	lago.RegistryPage.Register("academicrecords.AcademicRecordFormFields", editFormFields())
	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateFormFields", createFormFields())

	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateForm", &components.Modal{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateModal",
		},
		UID: "academicrecords-create-modal",
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				OnSubmit: getters.FormSubmitCloseModal(lago.RoutePath("academicrecords.CreateRoute", nil)),
				Method:   http.MethodPost,
				Title:    "Create Academic Record",
				Subtitle: "Pick student, program, term, and status. Compulsory courses are copied from that term in the program structure.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					createFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Save Academic Record", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				Getter: getters.Key[AcademicRecord]("academicrecord"),
				OnSubmit: getters.FormSubmit(lago.RoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				})),
				Method:   http.MethodPost,
				Title:    "Edit Academic Record",
				Subtitle: "Update status or course selections. Student, program, and term cannot be changed here.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					editFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Academic Record"},
				},
			},
		},
	})
}

// --- Tables ---

func registerTablePages() {
	academicRecordsSessionEnvironment := &components.Environment[uint]{
		Label:   "Session",
		Key:     getters.Static(academicRecordsEnvironmentSessionKey),
		Options: AcademicSessionsListGetter,
		Default: academicRecordsSessionEnvironmentDefault,
	}
	lago.RegistryPage.Register("academicrecords.AcademicRecordTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "students.StudentMenu"},
		},
		Children: []components.PageInterface{
			academicRecordsSessionEnvironment,
			&components.DataTable[AcademicRecord]{
				Page:    components.Page{Key: "academicrecords.AcademicRecordTableBody"},
				UID:     "academicrecords-table",
				Classes: "w-full",
				Data:    getters.Key[components.ObjectList[AcademicRecord]]("academicrecords"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "academicrecords.AcademicRecordFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
					&components.ButtonModal{
						Page:    components.Page{Roles: []string{"admin", "superuser"}},
						Url:     lago.RoutePath("academicrecords.CreateRoute", nil),
						Icon:    "plus",
						Classes: "btn-square btn-outline btn-sm",
					},
				},
				OnClick: getters.NavigateGetter(
					lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.Any(getters.Key[uint]("$row.ID")),
					}),
				),
				Columns: tableColumns(),
			},
		},
	})
}

// --- Detail & Delete ---

func registerDetailPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[AcademicRecord]{
				Getter: getters.Key[AcademicRecord]("academicrecord"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "academicrecords.AcademicRecordDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.Key[string]("$in.Student.User.Name")},
							&components.FieldSubtitle{Getter: getters.Key[string]("$in.Student.StudentNo")},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Program.Name")},
								},
							},
							&components.LabelInline{
								Title: "Session",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Session.Name")},
								},
							},
							&components.LabelInline{
								Title: "Status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.Key[string]("$in.Status")},
								},
							},
							&components.LabelInline{
								Title: "Term",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.Term"))),
									},
								},
							},
							&components.LabelNewline{
								Title: "Compulsory courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.Key[[]p_nirmancampus_courses.Course]("$in.CompulsoryCourses"),
										Display: getters.Key[string]("$in.Name"),
										Link:    courseDetailLink,
										Classes: "w-full",
									},
								},
							},
							&components.LabelNewline{
								Title: "Optional courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.Key[[]p_nirmancampus_courses.Course]("$in.OptionalCourses"),
										Display: getters.Key[string]("$in.Name"),
										Link:    courseDetailLink,
										Classes: "w-full",
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to delete this academic record?",
				CancelUrl: lago.RoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Selection Tables ---

func registerSelectionPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordSelectionTable", &components.Modal{
		UID: "academicrecords-selection-modal",
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page:  components.Page{Key: "academicrecords.AcademicRecordSelectionTableBody"},
				UID:   "academicrecords-selection-table",
				Title: "Select Academic Record",
				Data:  getters.Key[components.ObjectList[AcademicRecord]]("academicrecords"),
				OnClick: getters.Select("AcademicRecordID", getters.Key[uint]("$row.ID"), getters.Format(
					"%s (%s) · term %s",
					getters.Any(getters.Key[string]("$row.Program.Name")),
					getters.Any(getters.Key[string]("$row.Status")),
					getters.Any(getters.Format("%d", getters.Any(getters.Key[uint]("$row.Term")))),
				)),
				Columns: tableColumns(),
			},
		},
	})
}
