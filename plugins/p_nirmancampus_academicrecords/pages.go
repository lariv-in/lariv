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
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func init() {
	registerMenuPages()
	registerFilterPages()
	registerFormPages()
	registerTablePages()
	registerDetailPages()
	registerSelectionPages()
}

// --- Shared Helpers ---

// statusPairGetter reads a status string from the given context key, matches it
// against AcademicRecordStatusChoices, and optionally defaults to the first
// choice when the value is empty.
func statusPairGetter(key string, defaultToFirst bool) getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.Key[string](key)(ctx)
		if err != nil || s == "" {
			if defaultToFirst {
				if choices := AcademicRecordStatusChoices(); len(choices) > 0 {
					return choices[0], nil
				}
			}
			return registry.Pair[string, string]{}, nil
		}
		for _, p := range AcademicRecordStatusChoices() {
			if p.Key == s {
				return p, nil
			}
		}
		return registry.Pair[string, string]{Key: s, Value: s}, nil
	}
}

func statusInputSelect(forCreate bool) *components.InputSelect[string] {
	var g getters.Getter[registry.Pair[string, string]]
	if forCreate {
		g = statusPairGetter("$in.Status", true)
	} else {
		g = statusPairGetter("$in.Status", false)
	}
	return &components.InputSelect[string]{
		Label:    "Status",
		Name:     "Status",
		Required: true,
		Choices:  getters.Static(AcademicRecordStatusChoices()),
		Getter:   g,
	}
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
func programStructureUnitForIn(ctx context.Context, preloadOptionalPool bool) (p_nirmancampus_programs.ProgramStructureUnit, error) {
	var psu p_nirmancampus_programs.ProgramStructureUnit
	programID, err := getters.Key[uint]("$in.ProgramID")(ctx)
	if err != nil || programID == 0 {
		return psu, err
	}
	term, err := getters.Key[uint]("$in.Term")(ctx)
	if err != nil {
		return psu, err
	}
	db, ok := ctx.Value("$db").(*gorm.DB)
	if !ok || db == nil {
		return psu, fmt.Errorf("no db in context")
	}
	q := db.Where("program_id = ? AND term_number = ?", programID, term)
	if preloadOptionalPool {
		err = q.Preload("OptionalCourseSelectionPool").First(&psu).Error
	} else {
		err = q.Select("optional_course_count").First(&psu).Error
	}
	return psu, err
}

func optionalCourseCountDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		psu, err := programStructureUnitForIn(ctx, false)
		if err != nil {
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
		psu, err := programStructureUnitForIn(ctx, true)
		q := u.Query()
		if err != nil || len(psu.OptionalCourseSelectionPool) == 0 {
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

func registerMenuPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordMenu", &components.SidebarMenu{
		Title: getters.Static("Academic Records"),
		Back: &components.SidebarMenuItem{
			Title: getters.Static("Back to All Apps"),
			Url:   lago.RoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.Static("All Academic Records"),
				Url:   lago.RoutePath("academicrecords.DefaultRoute", nil),
			},
		},
	})

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
		Url:    lago.RoutePath("academicrecords.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputSelect[string]{
				Label:   "Status",
				Name:    "Status",
				Choices: getters.Static(AcademicRecordStatusChoices()),
				Getter:  statusPairGetter("$get.Status", false),
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
							statusInputSelect(true),
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
							statusInputSelect(false),
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

	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				Url:      lago.RoutePath("academicrecords.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Academic Record",
				Subtitle: "Pick student, program, term, and status. Compulsory courses are copied from that term in the program structure.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					createFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Academic Record"},
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
				Url: lago.RoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("$in.ID")),
				}),
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
	lago.RegistryPage.Register("academicrecords.AcademicRecordTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
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
					&components.TableButtonCreate{
						Link: lago.RoutePath("academicrecords.CreateRoute", nil),
						Page: components.Page{Roles: []string{"admin", "superuser"}},
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
		UID:   "academicrecords-selection-modal",
		Title: "Select Academic Record",
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page: components.Page{Key: "academicrecords.AcademicRecordSelectionTableBody"},
				UID:  "academicrecords-selection-table",
				Data: getters.Key[components.ObjectList[AcademicRecord]]("academicrecords"),
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
