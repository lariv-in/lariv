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

// --- Menus ---

func registerMenuPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Academic Records"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to All Apps"),
			Url:   lago.GetterRoutePath("dashboard.AppsPage", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Academic Records"),
				Url:   lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat("Record: %s", getters.GetterAny(getters.GetterKey[string]("academicrecord.Student.User.Name"))),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to all Academic Records"),
			Url:   lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Academic Record Detail"),
				Url: lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Edit Academic Record"),
				Url: lago.GetterRoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Roles: []string{"admin", "superuser"}},
				Title: getters.GetterStatic("Delete Academic Record"),
				Url: lago.GetterRoutePath("academicrecords.DeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
				}),
			},
		},
	})
}

// --- Filters ---

func registerFilterPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFilter", &components.FormComponent[AcademicRecord]{
		Url:    lago.GetterRoutePath("academicrecords.DefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputSelect[string]{
				Label:    "Status",
				Name:     "Status",
				Required: false,
				Choices:  getters.GetterStatic(AcademicRecordStatusChoices()),
				Getter:   academicRecordStatusFilterPairGetter(),
			},
			&components.InputText{
				Label:  "Term",
				Name:   "Term",
				Getter: getters.GetterKey[string]("$get.Term"),
			},
			&components.InputForeignKey[p_nirmancampus_programs.Program]{
				Label:       "Program",
				Name:        "ProgramID",
				Url:         lago.GetterRoutePath("programs.SelectRoute", nil),
				Placeholder: "Filter by program...",
				Display:     getters.GetterKey[string]("$in.Name"),
				Getter: getters.GetterAssociation[p_nirmancampus_programs.Program](
					getters.GetterKey[uint]("$get.ProgramID"),
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

func academicRecordStatusFilterPairGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$get.Status")(ctx)
		if err != nil || s == "" {
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

func academicRecordStatusPairFromInGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$in.Status")(ctx)
		if err != nil || s == "" {
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

func academicRecordStatusPairCreateGetter() getters.Getter[registry.Pair[string, string]] {
	return func(ctx context.Context) (registry.Pair[string, string], error) {
		s, err := getters.GetterKey[string]("$in.Status")(ctx)
		if err != nil || s == "" {
			choices := AcademicRecordStatusChoices()
			if len(choices) > 0 {
				return choices[0], nil
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

func academicRecordStatusInputSelect(forCreate bool) *components.InputSelect[string] {
	var g getters.Getter[registry.Pair[string, string]]
	if forCreate {
		g = academicRecordStatusPairCreateGetter()
	} else {
		g = academicRecordStatusPairFromInGetter()
	}
	return &components.InputSelect[string]{
		Label:    "Status",
		Name:     "Status",
		Required: true,
		Choices:  getters.GetterStatic(AcademicRecordStatusChoices()),
		Getter:   g,
	}
}

func academicRecordEditStudentDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		sid, err := getters.GetterKey[uint]("$in.StudentID")(ctx)
		if err != nil || sid == 0 {
			return "—", nil
		}
		dbVal := ctx.Value("$db")
		db, ok := dbVal.(*gorm.DB)
		if !ok || db == nil {
			return fmt.Sprintf("Student #%d", sid), nil
		}
		var st p_nirmancampus_students.Student
		if err := db.Preload("User").First(&st, sid).Error; err != nil {
			return fmt.Sprintf("Student #%d", sid), nil
		}
		name := ""
		if st.User.Name != "" {
			name = st.User.Name
		}
		if name == "" {
			return st.StudentNo, nil
		}
		return fmt.Sprintf("%s · %s", st.StudentNo, name), nil
	}
}

func academicRecordEditTermDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		if t, err := getters.GetterKey[int]("$in.Term")(ctx); err == nil {
			return fmt.Sprintf("%d", t), nil
		}
		if s, err := getters.GetterKey[string]("$in.Term")(ctx); err == nil && s != "" {
			return s, nil
		}
		return "—", nil
	}
}

func academicRecordTermHiddenIntGetter() getters.Getter[int] {
	return func(ctx context.Context) (int, error) {
		if t, err := getters.GetterKey[int]("$in.Term")(ctx); err == nil {
			return t, nil
		}
		if s, err := getters.GetterKey[string]("$in.Term")(ctx); err == nil && s != "" {
			n, err := strconv.Atoi(s)
			if err != nil {
				return 0, nil
			}
			return n, nil
		}
		return 0, nil
	}
}

// programStructureUnitForAcademicIn loads the ProgramStructureUnit for $in.ProgramID and $in.Term.
// When preloadOptionalPool is true, OptionalCourseSelectionPool is preloaded (for multi-select URLs).
func programStructureUnitForAcademicIn(ctx context.Context, preloadOptionalPool bool) (p_nirmancampus_programs.ProgramStructureUnit, error) {
	var psu p_nirmancampus_programs.ProgramStructureUnit
	programID, err := getters.GetterKey[uint]("$in.ProgramID")(ctx)
	if err != nil || programID == 0 {
		return psu, err
	}
	term, err := academicRecordTermHiddenIntGetter()(ctx)
	if err != nil {
		return psu, err
	}
	dbVal := ctx.Value("$db")
	db, ok := dbVal.(*gorm.DB)
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

func academicRecordOptionalCourseCountDisplayGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		psu, err := programStructureUnitForAcademicIn(ctx, false)
		if err != nil {
			return "—", nil
		}
		return strconv.Itoa(psu.OptionalCourseCount), nil
	}
}

func academicRecordOptionalCoursesMultiSelectURLGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		base, err := lago.GetterRoutePath("courses.MultiSelectRoute", nil)(ctx)
		if err != nil {
			return "", err
		}
		u, errParse := url.Parse(base)
		if errParse != nil {
			return base, nil
		}
		psu, err := programStructureUnitForAcademicIn(ctx, true)
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

// --- Form Fields ---

// academicRecordCreateFormFields is used on create only: Student, Program, Term, Status.
// Compulsory courses are filled server-side from the matching ProgramStructureUnit.
func academicRecordCreateFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.StudentID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_students.Student]{
								Label:       "Student",
								Name:        "StudentID",
								Required:    true,
								Url:         lago.GetterRoutePath("students.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.StudentNo"),
								Placeholder: "Select a student...",
								Getter: getters.GetterAssociation[p_nirmancampus_students.Student](
									getters.GetterKey[uint]("$in.StudentID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.ProgramID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[p_nirmancampus_programs.Program]{
								Label:       "Program",
								Name:        "ProgramID",
								Required:    true,
								Url:         lago.GetterRoutePath("programs.SelectRoute", nil),
								Display:     getters.GetterKey[string]("$in.Name"),
								Placeholder: "Select a program...",
								Getter: getters.GetterAssociation[p_nirmancampus_programs.Program](
									getters.GetterKey[uint]("$in.ProgramID"),
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
						Error: getters.GetterKey[error]("$error.Term"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Term",
								Name:     "Term",
								Required: true,
								Getter:   getters.GetterKey[int]("$in.Term"),
							},
						},
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Status"),
						Children: []components.PageInterface{
							academicRecordStatusInputSelect(true),
						},
					},
				},
			},
		},
	}
}

func academicRecordFormFields() components.ContainerColumn {
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
							&components.FieldText{Getter: academicRecordEditStudentDisplayGetter()},
						},
					},
					&components.LabelInline{
						Title: "Program",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterForeignKey[p_nirmancampus_programs.Program, uint, string](
									getters.GetterKey[uint]("$in.ProgramID"),
									"Name",
								),
							},
						},
					},
					&components.LabelInline{
						Title: "Term",
						Children: []components.PageInterface{
							&components.FieldText{Getter: academicRecordEditTermDisplayGetter()},
						},
					},
				},
			},
			&components.InputForeignKey[p_nirmancampus_students.Student]{
				Hidden: true,
				Name:   "StudentID",
				Getter: getters.GetterAssociation[p_nirmancampus_students.Student](
					getters.GetterKey[uint]("$in.StudentID"),
				),
			},
			&components.InputForeignKey[p_nirmancampus_programs.Program]{
				Hidden: true,
				Name:   "ProgramID",
				Getter: getters.GetterAssociation[p_nirmancampus_programs.Program](
					getters.GetterKey[uint]("$in.ProgramID"),
				),
			},
			&components.InputNumber{
				Hidden:   true,
				Name:     "Term",
				Getter:   academicRecordTermHiddenIntGetter(),
				Required: true,
			},
			&components.LabelNewline{
				Title: "Compulsory courses",
				Children: []components.PageInterface{
					&components.FieldManyToMany[p_nirmancampus_courses.Course]{
						Getter:  getters.GetterKey[[]p_nirmancampus_courses.Course]("$in.CompulsoryCourses"),
						Display: getters.GetterKey[string]("$in.Name"),
						Link: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
							"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
						}),
						Classes: "w-full",
					},
				},
			},
			&components.LabelInline{
				Title: "Optional course count",
				Children: []components.PageInterface{
					&components.FieldText{Getter: academicRecordOptionalCourseCountDisplayGetter()},
				},
			},
			&components.ContainerError{
				Error: getters.GetterKey[error]("$error.OptionalCourses"),
				Children: []components.PageInterface{
					&components.InputManyToMany[p_nirmancampus_courses.Course]{
						Label:       "Optional courses",
						Name:        "OptionalCourses",
						Required:    false,
						Getter:      getters.GetterKey[[]p_nirmancampus_courses.Course]("$in.OptionalCourses"),
						Url:         academicRecordOptionalCoursesMultiSelectURLGetter(),
						Display:     getters.GetterKey[string]("$in.Name"),
						Placeholder: "Select optional courses from the program pool…",
						Classes:     "w-full",
					},
				},
			},
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:max-w-md",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.Status"),
						Children: []components.PageInterface{
							academicRecordStatusInputSelect(false),
						},
					},
				},
			},
		},
	}
}

// --- Form Pages ---

func registerFormPages() {
	lago.RegistryPage.Register("academicrecords.AcademicRecordFormFields", academicRecordFormFields())
	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateFormFields", academicRecordCreateFormFields())

	lago.RegistryPage.Register("academicrecords.AcademicRecordCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[AcademicRecord]{
				Url:      lago.GetterRoutePath("academicrecords.CreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Create Academic Record",
				Subtitle: "Pick student, program, term, and status. Compulsory courses are copied from that term in the program structure.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					academicRecordCreateFormFields(),
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
				Getter: getters.GetterKey[AcademicRecord]("academicrecord"),
				Url: lago.GetterRoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Academic Record",
				Subtitle: "Update status or course selections. Student, program, and term cannot be changed here.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					academicRecordFormFields(),
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
	createURLGetter := lago.GetterRoutePath("academicrecords.CreateRoute", nil)

	lago.RegistryPage.Register("academicrecords.AcademicRecordTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[AcademicRecord]{
				Page:    components.Page{Key: "academicrecords.AcademicRecordTableBody"},
				UID:     "academicrecords-table",
				Classes: "w-full",
				Data:    getters.GetterKey[components.ObjectList[AcademicRecord]]("academicrecords"),
				Actions: []components.PageInterface{
					&components.TableButtonFilter{
						Child: lago.DynamicPage{Name: "academicrecords.AcademicRecordFilter"},
						Page:  components.Page{Roles: []string{"admin", "superuser"}},
					},
					&components.TableButtonCreate{
						Link: createURLGetter,
						Page: components.Page{Roles: []string{"admin", "superuser"}},
					},
				},
				OnClick: getters.GetterNavigateGetter(
					lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
						"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
					}),
				),
				Columns: []components.TableColumn{
					{
						Label: "Student",
						Name:  "Student.User.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Student.User.Name")},
						},
					},
					{
						Label: "Program",
						Name:  "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "Status",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Status")},
						},
					},
					{
						Label: "Term",
						Name:  "Term",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Term"))),
							},
						},
					},
				},
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
				Getter: getters.GetterKey[AcademicRecord]("academicrecord"),
				Children: []components.PageInterface{
					components.ContainerColumn{
						Page: components.Page{Key: "academicrecords.AcademicRecordDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Student.User.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Student.StudentNo")},
							&components.LabelInline{
								Title: "Program",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Program.Name")},
								},
							},
							&components.LabelInline{
								Title: "Status",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Status")},
								},
							},
							&components.LabelInline{
								Title: "Term",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$in.Term"))),
									},
								},
							},
							&components.LabelNewline{
								Title: "Compulsory courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.GetterKey[[]p_nirmancampus_courses.Course]("$in.CompulsoryCourses"),
										Display: getters.GetterKey[string]("$in.Name"),
										Link: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
										}),
										Classes: "w-full",
									},
								},
							},
							&components.LabelNewline{
								Title: "Optional courses",
								Children: []components.PageInterface{
									&components.FieldManyToMany[p_nirmancampus_courses.Course]{
										Getter:  getters.GetterKey[[]p_nirmancampus_courses.Course]("$in.OptionalCourses"),
										Display: getters.GetterKey[string]("$in.Name"),
										Link: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
											"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
										}),
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
				CancelUrl: lago.GetterRoutePath("academicrecords.DetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("academicrecord.ID")),
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
				Data: getters.GetterKey[components.ObjectList[AcademicRecord]]("academicrecords"),
				OnClick: getters.GetterSelect("AcademicRecordID", getters.GetterKey[uint]("$row.ID"), getters.GetterFormat(
					"%s · %s · term %d",
					getters.GetterAny(getters.GetterKey[string]("$row.Program.Name")),
					getters.GetterAny(getters.GetterKey[string]("$row.Status")),
					getters.GetterAny(getters.GetterKey[int]("$row.Term")),
				)),
				Columns: []components.TableColumn{
					{
						Label: "Student",
						Name:  "Student.User.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Student.User.Name")},
						},
					},
					{
						Label: "Program",
						Name:  "Program.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
						},
					},
					{
						Label: "Status",
						Name:  "Status",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Status")},
						},
					},
					{
						Label: "Term",
						Name:  "Term",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[int]("$row.Term"))),
							},
						},
					},
				},
			},
		},
	})
}
