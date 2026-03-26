package p_nirmancampus_courses

import (
	"context"
	"log"
	"net/http"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	"gorm.io/gorm"
)

func init() {
	registerCourseProgramMenuPages()
	registerCourseProgramFilterPages()
	registerCourseProgramFormPages()
	registerCourseProgramTablePages()
	registerCourseProgramDetailPages()
	registerCourseProgramSelectionPages()
	registerCourseProgramProgramPatches()
}

func courseProgramDraftGetter() getters.Getter[CourseProgram] {
	return func(ctx context.Context) (CourseProgram, error) {
		draft, ok := ctx.Value("courseprogram").(CourseProgram)
		if !ok {
			return CourseProgram{}, nil
		}
		return draft, nil
	}
}

func courseProgramRowsForCurrentCourseGetter() getters.Getter[components.ObjectList[CourseProgram]] {
	return func(ctx context.Context) (components.ObjectList[CourseProgram], error) {
		courseID, err := getters.GetterKey[uint]("$in.ID")(ctx)
		if err != nil || courseID == 0 {
			return components.ObjectList[CourseProgram]{Number: 1, NumPages: 1}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return components.ObjectList[CourseProgram]{}, nil
		}

		var rows []CourseProgram
		if err := db.Model(&CourseProgram{}).
			Preload("Course").
			Preload("Program").
			Where("course_id = ?", courseID).
			Order("semester ASC").
			Order("program_id ASC").
			Find(&rows).Error; err != nil {
			return components.ObjectList[CourseProgram]{}, err
		}

		return components.ObjectList[CourseProgram]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    int64(len(rows)),
		}, nil
	}
}

func courseProgramRowsForCurrentProgramGetter() getters.Getter[components.ObjectList[CourseProgram]] {
	return func(ctx context.Context) (components.ObjectList[CourseProgram], error) {
		programID, err := getters.GetterKey[uint]("$in.ID")(ctx)
		if err != nil || programID == 0 {
			return components.ObjectList[CourseProgram]{Number: 1, NumPages: 1}, nil
		}

		db, ok := ctx.Value("$db").(*gorm.DB)
		if !ok {
			return components.ObjectList[CourseProgram]{}, nil
		}

		var rows []CourseProgram
		if err := db.Model(&CourseProgram{}).
			Preload("Course").
			Preload("Program").
			Where("program_id = ?", programID).
			Order("semester ASC").
			Order("course_id ASC").
			Find(&rows).Error; err != nil {
			return components.ObjectList[CourseProgram]{}, err
		}

		return components.ObjectList[CourseProgram]{
			Items:    rows,
			Number:   1,
			NumPages: 1,
			Total:    int64(len(rows)),
		}, nil
	}
}

func courseProgramSemesterGetter(key string) getters.Getter[int] {
	return func(ctx context.Context) (int, error) {
		if value, err := getters.GetterKey[int](key)(ctx); err == nil {
			return value, nil
		}
		if value, err := getters.GetterKey[uint](key)(ctx); err == nil {
			return int(value), nil
		}
		return 0, nil
	}
}

func courseProgramFormFields() *components.ContainerColumn {
	return &components.ContainerColumn{
		Page: components.Page{Key: "courses.CourseProgramFormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.GetterKey[error]("$error.CourseID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[Course]{
								Label:       "Course",
								Name:        "CourseID",
								Required:    true,
								Url:         lago.GetterRoutePath("courses.SelectRoute", nil),
								Placeholder: "Select a course...",
								Display:     getters.GetterKey[string]("$in.Name"),
								Getter: getters.GetterAssociation[Course](
									getters.GetterKey[uint]("$in.CourseID"),
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
								Placeholder: "Select a program...",
								Display:     getters.GetterKey[string]("$in.Name"),
								Getter: getters.GetterAssociation[p_nirmancampus_programs.Program](
									getters.GetterKey[uint]("$in.ProgramID"),
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
						Error: getters.GetterKey[error]("$error.Semester"),
						Children: []components.PageInterface{
							&components.InputNumber{
								Label:    "Semester",
								Name:     "Semester",
								Required: true,
								Getter:   courseProgramSemesterGetter("$in.Semester"),
							},
						},
					},
				},
			},
		},
	}
}

func courseProgramCourseDetailSection() components.PageInterface {
	return &components.DataTable[CourseProgram]{
		Page:    components.Page{Key: "courses.CourseDetailProgramMappings"},
		UID:     "course-detail-program-mappings-table",
		Title:   "Program Mappings",
		Classes: "w-full mt-4",
		Data:    courseProgramRowsForCurrentCourseGetter(),
		CreateUrl: getters.GetterFormat(
			"%s?CourseID=%d",
			getters.GetterAny(lago.GetterRoutePath("courses.CourseProgramCreateRoute", nil)),
			getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
		),
		OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
		})),
		Columns: []components.TableColumn{
			{
				Label: "Program",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Name")},
				},
			},
			{
				Label: "Code",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Program.Code")},
				},
			},
			{
				Label: "Semester",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$row.Semester"))),
					},
				},
			},
		},
	}
}

func courseProgramProgramDetailSection() components.PageInterface {
	return &components.DataTable[CourseProgram]{
		Page:    components.Page{Key: "courses.ProgramDetailCourseMappings"},
		UID:     "program-detail-course-mappings-table",
		Title:   "Course Mappings",
		Classes: "w-full mt-4",
		Data:    courseProgramRowsForCurrentProgramGetter(),
		CreateUrl: getters.GetterFormat(
			"%s?ProgramID=%d",
			getters.GetterAny(lago.GetterRoutePath("courses.CourseProgramCreateRoute", nil)),
			getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
		),
		OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
			"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
		})),
		Columns: []components.TableColumn{
			{
				Label: "Course",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Name")},
				},
			},
			{
				Label: "Code",
				Children: []components.PageInterface{
					&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Code")},
				},
			},
			{
				Label: "Semester",
				Children: []components.PageInterface{
					&components.FieldText{
						Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$row.Semester"))),
					},
				},
			},
		},
	}
}

func registerCourseProgramProgramPatches() {
	lago.RegistryPage.Patch("programs.ProgramDetailMenu", func(page components.PageInterface) components.PageInterface {
		menu, ok := page.(*components.SidebarMenu)
		if !ok {
			log.Panic("programs.ProgramDetailMenu is not *components.SidebarMenu")
		}
		menu.Children = append(menu.Children,
			&components.SidebarMenuItem{
				Page:  components.Page{Key: "courses.program_detail_menu_course_mappings"},
				Title: getters.GetterStatic("Course Mappings"),
				Url: getters.GetterFormat(
					"%s?ProgramID=%d",
					getters.GetterAny(lago.GetterRoutePath("courses.CourseProgramDefaultRoute", nil)),
					getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				),
			},
			&components.SidebarMenuItem{
				Page:  components.Page{Key: "courses.program_detail_menu_add_mapping"},
				Title: getters.GetterStatic("Add Course Mapping"),
				Url: getters.GetterFormat(
					"%s?ProgramID=%d",
					getters.GetterAny(lago.GetterRoutePath("courses.CourseProgramCreateRoute", nil)),
					getters.GetterAny(getters.GetterKey[uint]("program.ID")),
				),
			},
		)
		return menu
	})

	lago.RegistryPage.Patch("programs.ProgramDetail", func(page components.PageInterface) components.PageInterface {
		scaffold, ok := page.(*components.ShellScaffold)
		if !ok {
			log.Panic("programs.ProgramDetail was not ShellScaffold")
		}
		components.ReplaceChild(scaffold, "programs.ProgramDetailContent", func(column components.ContainerColumn) components.ContainerColumn {
			column.Children = append(column.Children, courseProgramProgramDetailSection())
			return column
		})
		return scaffold
	})
}

func registerCourseProgramMenuPages() {
	lago.RegistryPage.Register("courses.CourseProgramMenu", &components.SidebarMenu{
		Title: getters.GetterStatic("Course Program Mappings"),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Courses"),
			Url:   lago.GetterRoutePath("courses.DefaultRoute", nil),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("All Mappings"),
				Url:   lago.GetterRoutePath("courses.CourseProgramDefaultRoute", nil),
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseProgramDetailMenu", &components.SidebarMenu{
		Title: getters.GetterFormat(
			"%s -> %s",
			getters.GetterAny(getters.GetterKey[string]("courseprogram.Course.Name")),
			getters.GetterAny(getters.GetterKey[string]("courseprogram.Program.Name")),
		),
		Back: &components.SidebarMenuItem{
			Title: getters.GetterStatic("Back to Course"),
			Url: lago.GetterRoutePath("courses.DetailRoute", map[string]getters.Getter[any]{
				"id": getters.GetterAny(getters.GetterKey[uint]("courseprogram.CourseID")),
			}),
		},
		Children: []components.PageInterface{
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Mapping Detail"),
				Url: lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("courseprogram.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Edit Mapping"),
				Url: lago.GetterRoutePath("courses.CourseProgramUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("courseprogram.ID")),
				}),
			},
			&components.SidebarMenuItem{
				Title: getters.GetterStatic("Delete Mapping"),
				Url: lago.GetterRoutePath("courses.CourseProgramDeleteRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("courseprogram.ID")),
				}),
			},
		},
	})
}

func registerCourseProgramFilterPages() {
	lago.RegistryPage.Register("courses.CourseProgramFilter", &components.FormComponent[CourseProgram]{
		Url:    lago.GetterRoutePath("courses.CourseProgramDefaultRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputForeignKey[Course]{
				Label:       "Course",
				Name:        "CourseID",
				Url:         lago.GetterRoutePath("courses.SelectRoute", nil),
				Placeholder: "Filter by course...",
				Display:     getters.GetterKey[string]("$in.Name"),
				Getter: getters.GetterAssociation[Course](
					getters.GetterKey[uint]("$get.CourseID"),
				),
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
			&components.InputNumber{
				Label:  "Semester",
				Name:   "Semester",
				Getter: courseProgramSemesterGetter("$get.Semester"),
			},
		},
		ChildrenAction: []components.PageInterface{
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply Filters"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})

	lago.RegistryPage.Register("courses.CourseProgramSelectionFilter", &components.FormComponent[CourseProgram]{
		Url:    lago.GetterRoutePath("courses.CourseProgramSelectRoute", nil),
		Method: http.MethodGet,
		ChildrenInput: []components.PageInterface{
			&components.InputForeignKey[Course]{
				Label:       "Course",
				Name:        "CourseID",
				Url:         lago.GetterRoutePath("courses.SelectRoute", nil),
				Placeholder: "Filter by course...",
				Display:     getters.GetterKey[string]("$in.Name"),
				Getter: getters.GetterAssociation[Course](
					getters.GetterKey[uint]("$get.CourseID"),
				),
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
			&components.ContainerRow{Classes: "flex gap-2", Children: []components.PageInterface{
				&components.ButtonSubmit{Label: "Apply"},
				&components.ButtonClear{Label: "Clear"},
			}},
		},
	})
}

func registerCourseProgramFormPages() {
	lago.RegistryPage.Register("courses.CourseProgramFormFields", courseProgramFormFields())

	lago.RegistryPage.Register("courses.CourseProgramCreateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[CourseProgram]{
				Getter:   courseProgramDraftGetter(),
				Url:      lago.GetterRoutePath("courses.CourseProgramCreateRoute", nil),
				Method:   http.MethodPost,
				Title:    "Add Program Mapping",
				Subtitle: "Attach a program to a course with its semester",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseProgramFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Mapping"},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseProgramUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormComponent[CourseProgram]{
				Getter: getters.GetterKey[CourseProgram]("courseprogram"),
				Url: lago.GetterRoutePath("courses.CourseProgramUpdateRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$in.ID")),
				}),
				Method:   http.MethodPost,
				Title:    "Edit Program Mapping",
				Subtitle: "Update the mapped program or semester",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					courseProgramFormFields(),
				},
				ChildrenAction: []components.PageInterface{
					&components.ButtonSubmit{Label: "Save Mapping"},
				},
			},
		},
	})
}

func registerCourseProgramTablePages() {
	lago.RegistryPage.Register("courses.CourseProgramTable", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseProgramMenu"},
		},
		Children: []components.PageInterface{
			&components.DataTable[CourseProgram]{
				Page:      components.Page{Key: "courses.CourseProgramTableBody"},
				UID:       "course-program-table",
				Title:     "Course Program Mappings",
				Classes:   "w-full",
				Data:      getters.GetterKey[components.ObjectList[CourseProgram]]("courseprograms"),
				CreateUrl: lago.GetterRoutePath("courses.CourseProgramCreateRoute", nil),
				OnClick: getters.GetterNavigateGetter(lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("$row.ID")),
				})),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseProgramFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Course",
						Name:  "Course.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Name")},
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
						Label: "Semester",
						Name:  "Semester",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$row.Semester"))),
							},
						},
					},
				},
			},
		},
	})
}

func registerCourseProgramDetailPages() {
	lago.RegistryPage.Register("courses.CourseProgramDetail", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.Detail[CourseProgram]{
				Getter: getters.GetterKey[CourseProgram]("courseprogram"),
				Children: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "courses.CourseProgramDetailContent"},
						Children: []components.PageInterface{
							&components.FieldTitle{Getter: getters.GetterKey[string]("$in.Course.Name")},
							&components.FieldSubtitle{Getter: getters.GetterKey[string]("$in.Program.Name")},
							&components.LabelInline{
								Title:   "Program Code",
								Classes: "mt-2",
								Children: []components.PageInterface{
									&components.FieldText{Getter: getters.GetterKey[string]("$in.Program.Code")},
								},
							},
							&components.LabelInline{
								Title: "Semester",
								Children: []components.PageInterface{
									&components.FieldText{
										Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$in.Semester"))),
									},
								},
							},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("courses.CourseProgramDeleteForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "courses.CourseProgramDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.DeleteConfirmation{
				Title:   "Confirm Deletion",
				Message: "Are you sure you want to remove this course-program mapping?",
				CancelUrl: lago.GetterRoutePath("courses.CourseProgramDetailRoute", map[string]getters.Getter[any]{
					"id": getters.GetterAny(getters.GetterKey[uint]("courseprogram.ID")),
				}),
			},
		},
	})
}

func registerCourseProgramSelectionPages() {
	lago.RegistryPage.Register("courses.CourseProgramSelectionTable", &components.Modal{
		UID:   "course-program-selection-modal",
		Title: "Select Course Program Mapping",
		Children: []components.PageInterface{
			&components.DataTable[CourseProgram]{
				Page: components.Page{Key: "courses.CourseProgramSelectionTableBody"},
				UID:  "course-program-selection-table",
				Data: getters.GetterKey[components.ObjectList[CourseProgram]]("courseprograms"),
				OnClick: getters.GetterSelect(
					"CourseProgramID",
					getters.GetterKey[uint]("$row.ID"),
					getters.GetterFormat(
						"%s / %s (Sem %d)",
						getters.GetterAny(getters.GetterKey[string]("$row.Course.Name")),
						getters.GetterAny(getters.GetterKey[string]("$row.Program.Name")),
						getters.GetterAny(getters.GetterKey[uint]("$row.Semester")),
					),
				),
				FilterComponent: lago.DynamicPage{Name: "courses.CourseProgramSelectionFilter"},
				Columns: []components.TableColumn{
					{
						Label: "Course",
						Name:  "Course.Name",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.GetterKey[string]("$row.Course.Name")},
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
						Label: "Semester",
						Name:  "Semester",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.GetterFormat("%d", getters.GetterAny(getters.GetterKey[uint]("$row.Semester"))),
							},
						},
					},
				},
			},
		},
	})
}
