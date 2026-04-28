package p_nirmancampus_examregistrations

import (
	"context"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
)

func examRegistrationFormCourseAndAcademicRecordRow() *components.ContainerRow {
	return &components.ContainerRow{
		Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
		Children: []components.PageInterface{
			&components.ContainerError{
				Error: getters.Key[error]("$error.CourseID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_nirmancampus_courses.Course]{
						Label:       "Course",
						Name:        "CourseID",
						Required:    true,
						Url:         lago.RoutePath("courses.SelectRoute", nil),
						Display:     getters.Key[string]("$in.Name"),
						Placeholder: "Select a course...",
						Getter: getters.Association[p_nirmancampus_courses.Course](
							getters.Key[uint]("$in.CourseID"),
						),
					},
				},
			},
			&components.ContainerError{
				Error: getters.Key[error]("$error.AcademicRecordID"),
				Children: []components.PageInterface{
					&components.InputForeignKey[p_nirmancampus_academicrecords.AcademicRecord]{
						Label:       "Academic record",
						Name:        "AcademicRecordID",
						Required:    true,
						Url:         lago.RoutePath("academicrecords.SelectRoute", nil),
						Display:     getters.Key[string]("$in.Student.StudentNo"),
						Placeholder: "Select an academic record...",
						Getter:      academicRecordForInputForeignKey(),
					},
				},
			},
		},
	}
}

func examRegistrationUpdateFormFields() *components.ContainerColumn {
	return &components.ContainerColumn{
		Page: components.Page{Key: "examregistrations.FormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.ExamTitle"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Exam title",
								Name:     "ExamTitle",
								Required: true,
								Getter:   getters.Key[string]("$in.ExamTitle"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.RegistrationStatus"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Registration status",
								Name:     "RegistrationStatus",
								Required: true,
								Choices:  getters.Static(ExamRegistrationStatusChoices),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.RegistrationStatus")(ctx)
									if err != nil || s == "" {
										if p, ok := registry.PairFromPairs(ExamRegistrationStatusNotRegisteredKey, ExamRegistrationStatusChoices); ok {
											return p, nil
										}
										return registry.Pair[string, string]{Key: ExamRegistrationStatusNotRegisteredKey, Value: "Not Registered"}, nil
									}
									if p, ok := registry.PairFromPairs(s, ExamRegistrationStatusChoices); ok {
										return p, nil
									}
									return registry.Pair[string, string]{Key: s, Value: s}, nil
								},
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.MaxMarks"),
						Children: []components.PageInterface{
							&components.InputNumber[int]{
								Label:    "Max marks",
								Name:     "MaxMarks",
								Required: true,
								Getter:   getters.Key[int]("$in.MaxMarks"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Marks"),
						Children: []components.PageInterface{
							&components.InputNumber[int]{
								Label:    "Marks",
								Name:     "Marks",
								Required: true,
								Getter:   getters.Key[int]("$in.Marks"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.Fee"),
						Children: []components.PageInterface{
							&components.InputNumber[uint]{
								Label:    "Fee (₹)",
								Name:     "Fee",
								Required: false,
								Getter:   getters.Key[uint]("$in.Fee"),
							},
						},
					},
				},
			},
			examRegistrationFormCourseAndAcademicRecordRow(),
			&components.ContainerError{
				Error: getters.Key[error]("$error.Assets"),
				Children: []components.PageInterface{
					&p_filesystem.InputMultiVNode{
						Label:            "Assets",
						Name:             "Assets",
						VNode:            getters.Key[[]p_filesystem.VNode]("$in.Assets"),
						AllowedFiletypes: []string{".pdf", ".jpg", ".jpeg", ".png", ".webp", ".doc", ".docx"},
					},
				},
			},
		},
	}
}

func registerFormPages() {
	lago.RegistryPage.Register("examregistrations.FormFields", examRegistrationUpdateFormFields())

	lago.RegistryPage.Register("examregistrations.BulkCreateFromAcademicRecordForm", &components.Modal{
		Page: components.Page{
			Key:   "examregistrations.BulkCreateFromAcademicRecordModal",
			Roles: []string{"admin", "superuser"},
		},
		UID: "examregistrations-bulk-create-academic-record-modal",
		Children: []components.PageInterface{
			&components.FormComponent[academicRecordBulkRegistrationsForm]{
				Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

				Title:    "Create exam registrations for student",
				Subtitle: "Select compulsory and/or optional courses. Title defaults to course name; fee defaults to the course fee where set.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "examregistrations.BulkCreateFromAcademicRecordFormBody"},
						Children: []components.PageInterface{
							&components.LabelInline{
								Title: "Student",
								Children: []components.PageInterface{
									&components.FieldText{Getter: bulkAcademicRecordStudentLineGetter()},
								},
							},
							&components.ContainerError{
								Error: getters.Key[error]("$error.AcademicRecordID"),
								Children: []components.PageInterface{
									&components.InputForeignKey[p_nirmancampus_academicrecords.AcademicRecord]{
										Hidden:   true,
										Name:     "AcademicRecordID",
										Required: true,
										Url:      lago.RoutePath("academicrecords.SelectRoute", nil),
										Display:  getters.Key[string]("$in.Student.StudentNo"),
										Getter:   academicRecordForInputForeignKey(),
									},
								},
							},
							&components.ContainerError{
								Error: getters.Key[error]("$error.BulkSelectedCourseIDs"),
								Children: []components.PageInterface{
									&InputBulkAcademicRecordCourses{
										Page:  components.Page{Key: "examregistrations.BulkCreateCourseSelection"},
										Label: "Courses on this academic record",
										Name:  bulkSelectedCourseIDsFieldName,
									},
								},
							},
						},
					},
				},
				ChildrenAction: []components.PageInterface{
					&components.ContainerRow{
						Classes: "flex justify-end gap-2 mt-2",
						Children: []components.PageInterface{
							&components.ButtonSubmit{Label: "Create registrations", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("examregistrations.UpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "examregistrations.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("examregistrations.UpdateForm"),
				ActionURL: lago.RoutePath("examregistrations.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("examregistration.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[ExamRegistration]{
						Getter: getters.Key[ExamRegistration]("examregistration"),
						Attr:   getters.FormBubbling(getters.Static("examregistrations.UpdateForm")),

						Title:    "Edit registration",
						Subtitle: "Update exam registration details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							examRegistrationUpdateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save registration"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("examregistrations.DeleteForm"),
												Url:         lago.RoutePath("examregistrations.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("examregistration.ID"))}),
												FormPostURL: lago.RoutePath("examregistrations.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("examregistration.ID"))}),
												ModalUID:    "examregistration-delete-modal",
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
