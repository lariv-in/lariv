package p_nirmancampus_academicrecords

import (
	"context"
	"time"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_programs"
	sessions "github.com/lariv-in/lago/plugins/p_nirmancampus_sessions"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_students"
	"github.com/lariv-in/lago/registry"
)

func createFormFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateFormFieldsBody",
		},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.SessionID"),
						Children: []components.PageInterface{
							&components.InputForeignKey[sessions.AdmissionSession]{
								Label:       "Admission session",
								Name:        "SessionID",
								Required:    true,
								Url:         lago.RoutePath("sessions.SelectRoute", nil),
								Display:     getters.Key[string]("$in.Name"),
								Placeholder: "Select an admission session…",
								Getter: getters.Association[sessions.AdmissionSession](
									getters.Key[uint]("$in.SessionID"),
								),
							},
						},
					},
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
								Label:    "Program",
								Name:     "ProgramID",
								Required: true,
								Url:      lago.RoutePath("programs.SelectRoute", nil),
								Display: p_nirmancampus_programs.ProgramDisplayLabel(
									getters.Key[string]("$in.Name"),
									getters.Key[string]("$in.University"),
								),
								Placeholder: "Select a program...",
								Getter: getters.Association[p_nirmancampus_programs.Program](
									getters.Key[uint]("$in.ProgramID"),
								),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Status"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Status",
								Name:     "Status",
								Required: true,
								Choices:  getters.Static(AcademicRecordStatusChoices),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.Status")(ctx)
									if err != nil || s == "" {
										if p, ok := registry.PairFromPairs("Enrolled", AcademicRecordStatusChoices); ok {
											return p, nil
										}
										return registry.Pair[string, string]{Key: "Enrolled", Value: "Enrolled"}, nil
									}
									if p, ok := registry.PairFromPairs(s, AcademicRecordStatusChoices); ok {
										return p, nil
									}
									return registry.Pair[string, string]{Key: s, Value: s}, nil
								},
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Date"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:    "Admission date",
								Name:     "Date",
								Required: true,
								Getter:   academicRecordDefaultGetter(getters.Key[time.Time]("$in.Date")),
							},
						},
					},
				},
			},
		},
	}
}

func createProgramStructureUnitFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateProgramStructureUnitBody",
		},
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.ProgramStructureUnitID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_nirmancampus_programs.ProgramStructureUnit]{
						Label:    "Term",
						Name:     "ProgramStructureUnitID",
						Required: true,
						Url:      programStructureUnitSelectURLGetter(),
						Display:  programStructureUnitDisplayGetter(),
						Getter: getters.Association[p_nirmancampus_programs.ProgramStructureUnit](
							getters.Key[uint]("$in.ProgramStructureUnitID"),
						),
					},
				},
			},
		},
	}
}

func createCourseSelectionFields() components.ContainerColumn {
	return components.ContainerColumn{
		Page: components.Page{
			Key: "academicrecords.AcademicRecordCreateCourseSelectionBody",
		},
		Children: []components.PageInterface{
			&components.LabelNewline{
				Title: "Compulsory courses",
				Children: []components.PageInterface{
					&components.FieldManyToMany[p_nirmancampus_courses.Course]{
						Getter:  getters.Key[[]p_nirmancampus_courses.Course](academicRecordProgramStructureUnitContextKey + ".CompulsoryCourses"),
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
								getters.Any(getters.Key[string]("$in.Student.Name")),
							)},
						},
					},
					&components.LabelInline{
						Title: "Program",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: p_nirmancampus_programs.ProgramDisplayLabel(
									getters.Key[string]("$in.Program.Name"),
									getters.Key[string]("$in.Program.University"),
								),
							},
						},
					},
					&components.LabelInline{
						Title: "Admission session",
						Children: []components.PageInterface{
							&components.FieldText{Getter: getters.Key[string]("$in.AdmissionSession.Name")},
						},
					},
					&components.LabelInline{
						Title: "Term",
						Children: []components.PageInterface{
							&components.FieldText{
								Getter: getters.Format("%d", getters.Any(getters.Key[uint]("$in.ProgramStructureUnit.TermNumber"))),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Date"),
						Children: []components.PageInterface{
							&components.InputDate{
								Label:    "Admission date",
								Name:     "Date",
								Required: true,
								Getter:   getters.Key[time.Time]("$in.Date"),
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
				Display: p_nirmancampus_programs.ProgramDisplayLabel(
					getters.Key[string]("$in.Name"),
					getters.Key[string]("$in.University"),
				),
				Getter: getters.Association[p_nirmancampus_programs.Program](
					getters.Key[uint]("$in.ProgramID"),
				),
			},
			&components.InputForeignKey[p_nirmancampus_programs.ProgramStructureUnit]{
				Hidden:  true,
				Name:    "ProgramStructureUnitID",
				Display: programStructureUnitDisplayGetter(),
				Getter: getters.Association[p_nirmancampus_programs.ProgramStructureUnit](
					getters.Key[uint]("$in.ProgramStructureUnitID"),
				),
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
								Choices:  getters.Static(AcademicRecordStatusChoices),
								Getter:   registry.PairFromGetter(getters.Key[string]("$in.Status"), AcademicRecordStatusChoices),
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
			&components.MultiStepForm{
				MultiStageURL: academicRecordCreateStageURLGetter(),
				Stages: []components.FormInterface{
					&components.FormComponent[AcademicRecord]{
						Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

						Title:    "Create Academic Record",
						Subtitle: "Pick student, program, admission session, and status.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							createFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex justify-end gap-2 mt-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Continue", Classes: "btn-primary"},
								},
							},
						},
					},
					&components.FormComponent[AcademicRecord]{
						Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

						Title:    "Select Term",
						Subtitle: "Choose the program term from the selected program structure.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							createProgramStructureUnitFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex justify-end gap-2 mt-2",
								Children: []components.PageInterface{
									&components.ButtonSubmit{Label: "Continue", Classes: "btn-primary"},
								},
							},
						},
					},
					&components.FormComponent[AcademicRecord]{
						Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

						Title:    "Select Courses",
						Subtitle: "Compulsory courses are prefilled from the selected term. Choose optional courses from that term's pool.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							createCourseSelectionFields(),
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
			},
		},
	})

	lago.RegistryPage.Register("academicrecords.AcademicRecordUpdateForm", &components.ShellScaffold{
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "academicrecords.AcademicRecordDetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("academicrecords.AcademicRecordUpdateForm"),
				ActionURL: lago.RoutePath("academicrecords.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("academicrecord.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[AcademicRecord]{
						Getter: getters.Key[AcademicRecord]("academicrecord"),
						Attr:   getters.FormBubbling(getters.Static("academicrecords.AcademicRecordUpdateForm")),

						Title:    "Edit Academic Record",
						Subtitle: "Update status or course selections. Student, program, and term cannot be changed here.",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							editFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save Academic Record"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("academicrecords.AcademicRecordDeleteForm"),
												Url:         lago.RoutePath("academicrecords.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("academicrecord.ID"))}),
												FormPostURL: lago.RoutePath("academicrecords.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("academicrecord.ID"))}),
												ModalUID:    "academicrecord-delete-modal",
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

// --- Tables ---
