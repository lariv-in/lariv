package p_nirmancampus_assignmentsubmissions

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/lariv-in/lago/components"
	"github.com/lariv-in/lago/getters"
	"github.com/lariv-in/lago/lago"
	"github.com/lariv-in/lago/plugins/p_filesystem"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_academicrecords"
	"github.com/lariv-in/lago/plugins/p_nirmancampus_courses"
	"github.com/lariv-in/lago/registry"
	"gorm.io/gorm"
)

func bulkAcademicRecordStudentLineGetter() getters.Getter[string] {
	return func(ctx context.Context) (string, error) {
		rec, ok := bulkAcademicRecordFromContext(ctx)
		if !ok || rec.ID == 0 {
			return "", nil
		}
		return fmt.Sprintf("%s — %s", rec.Student.Name, rec.Student.StudentNo), nil
	}
}

// assignmentSubmissionFormAcademicRecordGetter loads the academic record for the FK input with
// preloads so Display (Student.StudentNo) works; plain getters.Association does not preload Student.
func assignmentSubmissionFormAcademicRecordGetter() getters.Getter[p_nirmancampus_academicrecords.AcademicRecord] {
	return func(ctx context.Context) (p_nirmancampus_academicrecords.AcademicRecord, error) {
		var zero p_nirmancampus_academicrecords.AcademicRecord
		id, err := getters.Key[uint]("$in.AcademicRecordID")(ctx)
		if err != nil || id == 0 {
			return zero, nil
		}
		db, dberr := getters.DBFromContext(ctx)
		if dberr != nil {
			return zero, nil
		}
		var rec p_nirmancampus_academicrecords.AcademicRecord
		err = db.Preload("Student").Preload("Program").First(&rec, id).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return zero, nil
			}
			slog.Error("assignmentSubmissionFormAcademicRecordGetter: load failed", "error", err, "id", id)
			return zero, err
		}
		return rec, nil
	}
}

func assignmentSubmissionFormCourseAndAcademicRecordRow() *components.ContainerRow {
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
						Getter:      assignmentSubmissionFormAcademicRecordGetter(),
					},
				},
			},
		},
	}
}

func assignmentSubmissionUpdateFormFields() *components.ContainerColumn {
	return &components.ContainerColumn{
		Page: components.Page{Key: "assignmentsubmissions.FormFieldsBody"},
		Children: []components.PageInterface{
			&components.ContainerRow{
				Classes: "grid grid-cols-1 gap-1 @md:grid-cols-2",
				Children: []components.PageInterface{
					&components.ContainerError{
						Error: getters.Key[error]("$error.AssignmentTitle"),
						Children: []components.PageInterface{
							&components.InputText{
								Label:    "Assignment title",
								Name:     "AssignmentTitle",
								Required: true,
								Getter:   getters.Key[string]("$in.AssignmentTitle"),
							},
						},
					},
					&components.ContainerError{
						Error: getters.Key[error]("$error.SubmissionStatus"),
						Children: []components.PageInterface{
							&components.InputSelect[string]{
								Label:    "Submission status",
								Name:     "SubmissionStatus",
								Required: true,
								Choices:  getters.Static(AssignmentSubmissionStatusChoices),
								Getter: func(ctx context.Context) (registry.Pair[string, string], error) {
									s, err := getters.Key[string]("$in.SubmissionStatus")(ctx)
									if err != nil || s == "" {
										if p, ok := registry.PairFromPairs(AssignmentSubmissionStatusCreatedKey, AssignmentSubmissionStatusChoices); ok {
											return p, nil
										}
										return registry.Pair[string, string]{Key: AssignmentSubmissionStatusCreatedKey, Value: "Created"}, nil
									}
									if p, ok := registry.PairFromPairs(s, AssignmentSubmissionStatusChoices); ok {
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
				},
			},
			assignmentSubmissionFormCourseAndAcademicRecordRow(),
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
	lago.RegistryPage.Register("assignmentsubmissions.FormFields", assignmentSubmissionUpdateFormFields())

	lago.RegistryPage.Register("assignmentsubmissions.BulkCreateFromAcademicRecordForm", &components.Modal{
		Page: components.Page{
			Key:   "assignmentsubmissions.BulkCreateFromAcademicRecordModal",
			Roles: []string{"admin", "superuser"},
		},
		UID: "assignmentsubmissions-bulk-create-academic-record-modal",
		Children: []components.PageInterface{
			&components.FormComponent[academicRecordBulkSubmissionsForm]{
				Attr: getters.FormBubbling(getters.Key[string]("$get.name")),

				Title:    "Create submissions for student",
				Subtitle: "Select compulsory and/or optional courses. Title defaults to course name; marks stay zero until you edit each submission.",
				Classes:  "@container",
				ChildrenInput: []components.PageInterface{
					&components.ContainerColumn{
						Page: components.Page{Key: "assignmentsubmissions.BulkCreateFromAcademicRecordFormBody"},
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
										Getter:   assignmentSubmissionFormAcademicRecordGetter(),
									},
								},
							},
							&components.ContainerError{
								Error: getters.Key[error]("$error.BulkSelectedCourseIDs"),
								Children: []components.PageInterface{
									&InputBulkAcademicRecordCourses{
										Page:  components.Page{Key: "assignmentsubmissions.BulkCreateCourseSelection"},
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
							&components.ButtonSubmit{Label: "Create submissions", Classes: "btn-primary"},
						},
					},
				},
			},
		},
	})

	lago.RegistryPage.Register("assignmentsubmissions.UpdateForm", &components.ShellScaffold{
		Page: components.Page{Roles: []string{"admin", "superuser"}},
		Sidebar: []components.PageInterface{
			lago.DynamicPage{Name: "assignmentsubmissions.DetailMenu"},
		},
		Children: []components.PageInterface{
			&components.FormListenBoostedPost{
				Name: getters.Static("assignmentsubmissions.UpdateForm"),
				ActionURL: lago.RoutePath("assignmentsubmissions.UpdateRoute", map[string]getters.Getter[any]{
					"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID")),
				}),
				Children: []components.PageInterface{
					&components.FormComponent[AssignmentSubmission]{
						Getter: getters.Key[AssignmentSubmission]("assignmentsubmission"),
						Attr:   getters.FormBubbling(getters.Static("assignmentsubmissions.UpdateForm")),

						Title:    "Edit submission",
						Subtitle: "Update assignment submission details",
						Classes:  "@container",
						ChildrenInput: []components.PageInterface{
							assignmentSubmissionUpdateFormFields(),
						},
						ChildrenAction: []components.PageInterface{
							&components.ContainerRow{
								Classes: "flex flex-wrap justify-between gap-2 mt-2 items-center",
								Children: []components.PageInterface{
									&components.ContainerRow{
										Classes: "flex justify-end gap-2",
										Children: []components.PageInterface{
											&components.ButtonSubmit{Label: "Save submission"},
											&components.ButtonModalForm{
												Page:        components.Page{Roles: []string{"admin", "superuser"}},
												Label:       "Delete",
												Icon:        "trash",
												Name:        getters.Static("assignmentsubmissions.DeleteForm"),
												Url:         lago.RoutePath("assignmentsubmissions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID"))}),
												FormPostURL: lago.RoutePath("assignmentsubmissions.DeleteRoute", map[string]getters.Getter[any]{"id": getters.Any(getters.Key[uint]("assignmentsubmission.ID"))}),
												ModalUID:    "assignmentsubmission-delete-modal",
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
